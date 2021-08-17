package provider

import (
	"context"
	stdRsa "crypto/rsa"
	"errors"
	"fmt"
	"time"

	extaws "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/palantir/stacktrace"
	"github.com/sumup-oss/go-pkgs/os"
	"github.com/sumup-oss/vaulted/pkg/aes"
	"github.com/sumup-oss/vaulted/pkg/aws"
	"github.com/sumup-oss/vaulted/pkg/base64"
	"github.com/sumup-oss/vaulted/pkg/pkcs7"
	"github.com/sumup-oss/vaulted/pkg/rsa"
	"github.com/sumup-oss/vaulted/pkg/vaulted/content"
	"github.com/sumup-oss/vaulted/pkg/vaulted/passphrase"
	"github.com/sumup-oss/vaulted/pkg/vaulted/payload"
)

func New() func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"aws_kms_key_id": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("VAULTED_AWS_KMS_KEY_ID", ""),
					Description: "Either AWS KMS key ARN or AWS KMS key alias, used to decrypt. " +
						"Make sure AWS_REGION and/or AWS_PROFILE environment variables are pointing to an AWS account that has the given KMS key." +
						"This setting has higher priority than `private_key_content`.",
				},
				"aws_profile": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AWS_PROFILE", ""),
					Description: "AWS profile to use when authenticating against AWS. Equivalent of `AWS_PROFILE` env var that also works. " +
						"In practice only useful when `aws_kms_key_id` is provided",
				},
				// NOTE: Intentionally mimic the official `terraform-provider-aws` as much as possible
				// to make the use for anyone already familiar with it, smooth.
				"aws_assume_role": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"duration_seconds": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "Seconds to restrict the assume role session duration.",
							},
							"external_id": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Unique identifier that might be required for assuming a role in another account.",
							},
							"policy": {
								Type:         schema.TypeString,
								Optional:     true,
								Description:  "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
								ValidateFunc: validation.StringIsJSON,
							},
							"policy_arns": {
								Type:     schema.TypeSet,
								Optional: true,
								Description: "Amazon Resource Names (ARNs) of IAM Policies describing further restricting " +
									"permissions for the IAM Role being assumed.",
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validateArn,
								},
							},
							"role_arn": {
								Type:         schema.TypeString,
								Optional:     true,
								Description:  "Amazon Resource Name of an IAM Role to assume prior to making API calls.",
								ValidateFunc: validateArn,
							},
							"session_name": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Identifier for the assumed role session.",
							},
						},
					},
				},
				"aws_region": {
					Type:     schema.TypeString,
					Optional: true,
					DefaultFunc: schema.MultiEnvDefaultFunc([]string{
						"AWS_REGION",
						"AWS_DEFAULT_REGION",
					}, nil),
					Description: "AWS Region to use where `aws_kms_key_id` is present.",
				},
				"private_key_content": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("VAULTED_PRIVATE_KEY_CONTENT", ""),
					Description: "Content of private key used to decrypt. This setting has higher priority than `private_key_path`.",
				},
				"private_key_path": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("VAULTED_PRIVATE_KEY_PATH", ""),
					Description: "Path to private key used to decrypt. This setting has lower priority than `private_key_content`.",
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"vaulted-null_content": dataSourceContent(),
			},
		}

		p.ConfigureContextFunc = configure()

		return p
	}
}

func configure() func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		osExecutor := &os.RealOsExecutor{}
		rsaSvc := rsa.NewRsaService(osExecutor)
		aesSvc := aes.NewAesService(pkcs7.NewPkcs7Service())
		b64Svc := base64.NewBase64Service()

		contentSvc := content.NewV1Service(b64Svc, aesSvc)

		var payloadDecrypter PayloadDecrypter

		awsKMSkeyID, ok := d.Get("aws_kms_key_id").(string)
		if !ok {
			return nil, diag.FromErr(errors.New("unexpected non-string `aws_kms_key_id`"))
		}

		if awsKMSkeyID == "" {
			privateKey, err := readPrivateKey(d, osExecutor, rsaSvc)
			if err != nil {
				return nil, diag.FromErr(err)
			}

			passphraseDecrypter := passphrase.NewDecryptionRsaPKCS1v15Service(privateKey, rsaSvc)
			payloadDecrypter = payload.NewDecryptionService(passphraseDecrypter, contentSvc)
		} else {
			awsCfg, err := readAWScfg(ctx, d)
			if err != nil {
				return nil, diag.FromErr(err)
			}

			awsSvc, _ := aws.NewService(awsCfg)
			passphraseDecrypter := passphrase.NewDecryptionAwsKmsService(awsSvc, awsKMSkeyID)
			payloadDecrypter = payload.NewDecryptionService(passphraseDecrypter, contentSvc)
		}

		payloadDeserializer := payload.NewSerdeService(b64Svc)

		return &MetaClient{payloadDecrypter: payloadDecrypter, payloadDeserializer: payloadDeserializer}, nil
	}
}

func readAWScfg(ctx context.Context, d *schema.ResourceData) (*extaws.Config, error) {
	awsRegion, ok := d.Get("aws_region").(string)
	if !ok {
		return nil, errors.New("unexpected non-string `aws_region`")
	}

	awsCfgResolvers := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(awsRegion)}

	awsProfile, ok := d.Get("aws_profile").(string)
	if !ok {
		return nil, errors.New("unexpected non-string `aws_profile`")
	}

	if awsProfile != "" {
		awsCfgResolvers = append(awsCfgResolvers, awsconfig.WithSharedConfigProfile(awsProfile))
	}

	awsAssumeRole, ok := d.Get("aws_assume_role").([]interface{})
	if ok && len(awsAssumeRole) > 0 && awsAssumeRole[0] != nil {
		m, ok := awsAssumeRole[0].(map[string]interface{})
		if !ok {
			return nil, errors.New("unexpected non-map with key string, value interface{} - `aws_assume_role[0]`")
		}

		awsconfig.WithAssumeRoleCredentialOptions(func(opts *stscreds.AssumeRoleOptions) {
			if v, ok := m["duration_seconds"].(int); ok && v != 0 {
				opts.Duration = time.Second * time.Duration(v)
			}

			if v, ok := m["external_id"].(string); ok && v != "" {
				opts.ExternalID = extaws.String(v)
			}

			if v, ok := m["policy"].(string); ok && v != "" {
				opts.Policy = extaws.String(v)
			}

			if policyARNSet, ok := m["policy_arns"].(*schema.Set); ok && policyARNSet.Len() > 0 {
				for _, policyARNRaw := range policyARNSet.List() {
					policyARN, ok := policyARNRaw.(string)

					if !ok {
						continue
					}

					opts.PolicyARNs = append(
						opts.PolicyARNs,
						types.PolicyDescriptorType{
							Arn: extaws.String(policyARN),
						},
					)
				}
			}

			if v, ok := m["role_arn"].(string); ok && v != "" {
				opts.RoleARN = v
			}

			if v, ok := m["session_name"].(string); ok && v != "" {
				opts.RoleSessionName = v
			}
		})
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsCfgResolvers...)
	if err != nil {
		return nil, err
	}

	return &awsCfg, nil
}

func readPrivateKey(
	d *schema.ResourceData,
	osExecutor os.OsExecutor,
	rsaSvc *rsa.Service,
) (*stdRsa.PrivateKey, error) {
	var privateKey *stdRsa.PrivateKey

	privateKeyContentTypeless := d.Get("private_key_content")
	switch privateKeyContent := privateKeyContentTypeless.(type) {
	case string:
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

			// NOTE: Clean up the private key from the disk
			nestedErr = osExecutor.Remove(fd.Name())
			if nestedErr != nil {
				return nil, stacktrace.NewError(
					"failed to remove temporary file for vaulted private key from content: %s",
					nestedErr,
				)
			}
		}
	default: // NOTE: Do nothing, try with `private_key_path`.
	}

	if privateKey == nil {
		privateKeyPathTypeless := d.Get("private_key_path")
		switch privateKeyPath := privateKeyPathTypeless.(type) {
		case string:
			if privateKeyPath != "" {
				key, readErr := rsaSvc.ReadPrivateKeyFromPath(privateKeyPath)
				if readErr != nil {
					return nil, fmt.Errorf("failed to read private key from path %s, err: %w", privateKeyPath, readErr)
				}

				privateKey = key
			}
		default:
			return nil, stacktrace.NewError("non-string private_key_path. actual: %#v", privateKeyPath)
		}
	}

	if privateKey == nil {
		return nil, stacktrace.NewError(
			"failed to read RSA private key from either `private_key_content` or" +
				" `private_key_path` provider attributes",
		)
	}

	return privateKey, nil
}
