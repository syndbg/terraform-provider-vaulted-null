package provider

import (
	"context"
	"crypto/rsa"

	extaws "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	frameworktypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/palantir/stacktrace"
	"github.com/sumup-oss/go-pkgs/os"
	"github.com/sumup-oss/vaulted/pkg/aes"
	"github.com/sumup-oss/vaulted/pkg/aws"
	"github.com/sumup-oss/vaulted/pkg/base64"
	"github.com/sumup-oss/vaulted/pkg/pkcs7"
	vaultedRsa "github.com/sumup-oss/vaulted/pkg/rsa"
	"github.com/sumup-oss/vaulted/pkg/vaulted/content"
	"github.com/sumup-oss/vaulted/pkg/vaulted/passphrase"
	"github.com/sumup-oss/vaulted/pkg/vaulted/payload"

	"github.com/syndbg/terraform-provider-vaulted-null/internal/vaulted"
)

// Ensure Provider satisfies various provider interfaces.
var (
	_ provider.Provider                       = &Provider{}
	_ provider.ProviderWithEphemeralResources = &Provider{}
)

// Provider defines the provider implementation using Plugin Framework.
type Provider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ProviderModel describes the provider data model.
type ProviderModel struct {
	AwsKmsKeyId       frameworktypes.String `tfsdk:"aws_kms_key_id"`
	AwsProfile        frameworktypes.String `tfsdk:"aws_profile"`
	AwsRegion         frameworktypes.String `tfsdk:"aws_region"`
	PublicKeyPath     frameworktypes.String `tfsdk:"public_key_path"`
	PublicKeyContent  frameworktypes.String `tfsdk:"public_key_content"`
	PrivateKeyPath    frameworktypes.String `tfsdk:"private_key_path"`
	PrivateKeyContent frameworktypes.String `tfsdk:"private_key_content"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{
			version: version,
		}
	}
}

func (p *Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vaulted-null"
	resp.Version = p.version
}

func (p *Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aws_kms_key_id": schema.StringAttribute{
				MarkdownDescription: "Either AWS KMS key ARN or AWS KMS key alias, used to decrypt. " +
					"Make sure AWS_REGION and/or AWS_PROFILE environment variables are pointing to an AWS account that has the given KMS key." +
					"This setting has higher priority than `private_key_content`.",
				Optional: true,
			},
			"aws_profile": schema.StringAttribute{
				MarkdownDescription: "AWS profile to use when authenticating against AWS. Equivalent of `AWS_PROFILE` env var that also works. " +
					"In practice only useful when `aws_kms_key_id` is provided",
				Optional: true,
			},
			"aws_region": schema.StringAttribute{
				MarkdownDescription: "AWS Region to use where `aws_kms_key_id` is present.",
				Optional:            true,
			},
			"public_key_content": schema.StringAttribute{
				MarkdownDescription: "Content of public key used to encrypt. This setting has higher priority than `public_key_path`.",
				Optional:            true,
			},
			"public_key_path": schema.StringAttribute{
				MarkdownDescription: "Path to public key used to decrypt. This setting has lower priority than `public_key_content`.",
				Optional:            true,
			},
			"private_key_content": schema.StringAttribute{
				MarkdownDescription: "Content of private key used to decrypt. This setting has higher priority than `private_key_path`.",
				Optional:            true,
				Sensitive:           true,
			},
			"private_key_path": schema.StringAttribute{
				MarkdownDescription: "Path to private key used to decrypt. This setting has lower priority than `private_key_content`.",
				Optional:            true,
			},
		},
	}
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Initialize services - replicate SDK v2 logic
	osExecutor := &os.RealOsExecutor{}
	rsaSvc := vaultedRsa.NewRsaService(osExecutor)
	aesSvc := aes.NewAesService(pkcs7.NewPkcs7Service())
	b64Svc := base64.NewBase64Service()

	contentSvc := content.NewV1Service(b64Svc, aesSvc)

	var payloadEncrypter PayloadEncrypter
	var payloadDecrypter PayloadDecrypter

	awsKMSkeyID := data.AwsKmsKeyId.ValueString()

	if awsKMSkeyID == "" {
		// RSA key path
		publicKey, err := p.readPublicKeyFromConfig(data, osExecutor, rsaSvc)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to read public key",
				err.Error(),
			)
			return
		}

		privateKey, err := p.readPrivateKeyFromConfig(data, osExecutor, rsaSvc)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to read private key",
				err.Error(),
			)
			return
		}

		passphraseEncrypter := passphrase.NewEncryptionRsaPKCS1v15Service(rsaSvc, publicKey)
		contentEncrypter := content.NewV1Service(b64Svc, aesSvc)
		payloadEncrypter = payload.NewEncryptionService(passphraseEncrypter, contentEncrypter)

		// If there's no private key provided, the provider would be used for encryption-only resources
		if privateKey == nil {
			payloadDecrypter = vaulted.NewErrorDecryptionService(
				"failed to read RSA private key from either `private_key_content` or " +
					"`private_key_path` provider attributes",
			)
		} else {
			passphraseDecrypter := passphrase.NewDecryptionRsaPKCS1v15Service(privateKey, rsaSvc)
			payloadDecrypter = payload.NewDecryptionService(passphraseDecrypter, contentSvc)
		}
	} else {
		// AWS KMS path
		publicKey, err := p.readPublicKeyFromConfig(data, osExecutor, rsaSvc)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to read public key",
				err.Error(),
			)
			return
		}

		awsCfg, err := p.readAWSConfigFromConfig(ctx, data)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to configure AWS",
				err.Error(),
			)
			return
		}

		awsSvc, _ := aws.NewService(awsCfg)
		contentEncrypter := content.NewV1Service(b64Svc, aesSvc)
		passphraseEncrypter := passphrase.NewEncRsaOaepService(rsaSvc, publicKey)
		payloadEncrypter = payload.NewEncryptionService(passphraseEncrypter, contentEncrypter)

		passphraseDecrypter := passphrase.NewDecryptionAwsKmsService(awsSvc, awsKMSkeyID)
		payloadDecrypter = payload.NewDecryptionService(passphraseDecrypter, contentSvc)
	}

	payloadSerdeSvc := payload.NewSerdeService(b64Svc)

	metaClient := &MetaClient{
		payloadSerializer:   payloadSerdeSvc,
		payloadEncrypter:    payloadEncrypter,
		payloadDeserializer: payloadSerdeSvc,
		payloadDecrypter:    payloadDecrypter,
	}

	resp.DataSourceData = metaClient
	resp.ResourceData = metaClient
	resp.EphemeralResourceData = metaClient
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewEncryptContentResource,
	}
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewContentDataSource,
	}
}

func (p *Provider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		NewEphemeralEncryptResource,
		NewEphemeralDecryptResource,
	}
}

// Helper functions replicated from SDK provider

func (p *Provider) readPublicKeyFromConfig(
	data ProviderModel,
	osExecutor os.OsExecutor,
	rsaSvc *vaultedRsa.Service,
) (*rsa.PublicKey, error) {
	var publicKey *rsa.PublicKey

	// Try public_key_content first (higher priority)
	publicKeyContent := data.PublicKeyContent.ValueString()
	if publicKeyContent != "" {
		fd, nestedErr := osExecutor.TempFile("", "vaulted-public-key-from-content")
		if nestedErr != nil {
			return nil, stacktrace.NewError(
				"failed to create temporary file for vaulted public key from content: %s",
				nestedErr,
			)
		}

		_, nestedErr = fd.WriteString(publicKeyContent)
		if nestedErr != nil {
			return nil, stacktrace.NewError(
				"failed to write public key content to temporary file for vaulted public key: %s",
				nestedErr,
			)
		}

		nestedErr = fd.Sync()
		if nestedErr != nil {
			return nil, stacktrace.NewError(
				"failed to sync public key content to temporary file for vaulted public key: %s",
				nestedErr,
			)
		}

		nestedErr = fd.Close()
		if nestedErr != nil {
			return nil, stacktrace.NewError(
				"failed to close temporary file for vaulted public key from content: %s",
				nestedErr,
			)
		}

		key, readErr := rsaSvc.ReadPublicKeyFromPath(fd.Name())
		if readErr != nil {
			return nil, stacktrace.Propagate(readErr, "failed to read public key from path")
		}

		publicKey = key

		// Clean up the temporary file
		nestedErr = osExecutor.Remove(fd.Name())
		if nestedErr != nil {
			return nil, stacktrace.NewError(
				"failed to remove temporary file for vaulted public key from content: %s",
				nestedErr,
			)
		}
	}

	// Try public_key_path if content wasn't provided
	if publicKey == nil {
		publicKeyPath := data.PublicKeyPath.ValueString()
		if publicKeyPath != "" {
			key, readErr := rsaSvc.ReadPublicKeyFromPath(publicKeyPath)
			if readErr != nil {
				return nil, stacktrace.Propagate(readErr, "failed to read public key from path %s", publicKeyPath)
			}
			publicKey = key
		}
	}

	if publicKey == nil {
		return nil, stacktrace.NewError(
			"failed to read RSA public key from either `public_key_content` or" +
				" `public_key_path` provider attributes",
		)
	}

	return publicKey, nil
}

func (p *Provider) readPrivateKeyFromConfig(
	data ProviderModel,
	osExecutor os.OsExecutor,
	rsaSvc *vaultedRsa.Service,
) (*rsa.PrivateKey, error) {
	var privateKey *rsa.PrivateKey

	// Try private_key_content first (higher priority)
	privateKeyContent := data.PrivateKeyContent.ValueString()
	if privateKeyContent != "" {
		fd, nestedErr := osExecutor.TempFile("", "vaulted-private-key-from-content")
		if nestedErr != nil {
			return nil, stacktrace.NewError(
				"failed to create temporary file for vaulted private key from content: %s",
				nestedErr,
			)
		}

		_, nestedErr = fd.WriteString(privateKeyContent)
		if nestedErr != nil {
			return nil, stacktrace.NewError(
				"failed to write private key content to temporary file for vaulted private key: %s",
				nestedErr,
			)
		}

		nestedErr = fd.Sync()
		if nestedErr != nil {
			return nil, stacktrace.NewError(
				"failed to sync private key content to temporary file for vaulted private key: %s",
				nestedErr,
			)
		}

		nestedErr = fd.Close()
		if nestedErr != nil {
			return nil, stacktrace.NewError(
				"failed to close temporary file for vaulted private key from content: %s",
				nestedErr,
			)
		}

		key, readErr := rsaSvc.ReadPrivateKeyFromPath(fd.Name())
		if readErr != nil {
			return nil, stacktrace.Propagate(readErr, "failed to read private key from path")
		}

		privateKey = key

		// Clean up the temporary file
		nestedErr = osExecutor.Remove(fd.Name())
		if nestedErr != nil {
			return nil, stacktrace.NewError(
				"failed to remove temporary file for vaulted private key from content: %s",
				nestedErr,
			)
		}
	}

	// Try private_key_path if content wasn't provided
	if privateKey == nil {
		privateKeyPath := data.PrivateKeyPath.ValueString()
		if privateKeyPath != "" {
			key, readErr := rsaSvc.ReadPrivateKeyFromPath(privateKeyPath)
			if readErr != nil {
				return nil, stacktrace.Propagate(readErr, "failed to read private key from path %s", privateKeyPath)
			}
			privateKey = key
		}
	}

	// Note: privateKey can be nil for encryption-only use cases
	return privateKey, nil
}

func (p *Provider) readAWSConfigFromConfig(ctx context.Context, data ProviderModel) (*extaws.Config, error) {
	awsRegion := data.AwsRegion.ValueString()
	awsProfile := data.AwsProfile.ValueString()

	awsCfgResolvers := make([]func(*awsconfig.LoadOptions) error, 0)

	if awsRegion != "" {
		awsCfgResolvers = append(awsCfgResolvers, awsconfig.WithRegion(awsRegion))
	}

	if awsProfile != "" {
		awsCfgResolvers = append(awsCfgResolvers, awsconfig.WithSharedConfigProfile(awsProfile))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsCfgResolvers...)
	if err != nil {
		return nil, err
	}

	return &awsCfg, nil
}
