package provider

import (
	"context"
	stdRsa "crypto/rsa"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/palantir/stacktrace"
	"github.com/sumup-oss/go-pkgs/os"
	"github.com/sumup-oss/vaulted/pkg/aes"
	"github.com/sumup-oss/vaulted/pkg/base64"
	"github.com/sumup-oss/vaulted/pkg/pkcs7"
	"github.com/sumup-oss/vaulted/pkg/rsa"
	"github.com/sumup-oss/vaulted/pkg/vaulted/content"
	"github.com/sumup-oss/vaulted/pkg/vaulted/header"
	"github.com/sumup-oss/vaulted/pkg/vaulted/passphrase"
	"github.com/sumup-oss/vaulted/pkg/vaulted/payload"
)

type MetaClient struct {
	VaultedPrivateKey *stdRsa.PrivateKey
}

func (m *MetaClient) DecryptValue(encryptedValue string) (string, error) {
	osExecutor := &os.RealOsExecutor{}
	b64Svc := base64.NewBase64Service()
	rsaSvc := rsa.NewRsaService(osExecutor)
	aesSvc := aes.NewAesService(pkcs7.NewPkcs7Service())

	encPayloadSvc := payload.NewEncryptedPayloadService(
		header.NewHeaderService(),
		passphrase.NewEncryptedPassphraseService(b64Svc, rsaSvc),
		content.NewV1EncryptedContentService(b64Svc, aesSvc),
	)

	deserializedValue, err := encPayloadSvc.Deserialize([]byte(encryptedValue))
	if err != nil {
		return "", stacktrace.Propagate(err, "unable to serialize `value`")
	}

	decryptedValue, err := encPayloadSvc.Decrypt(m.VaultedPrivateKey, deserializedValue)
	if err != nil {
		return "", stacktrace.Propagate(err, "unable to decrypt `value`")
	}

	return string(decryptedValue.Content.Plaintext), nil
}

// nolint:gochecknoinits
func init() {
	// NOTE: Part of TF registry docs generation
	schema.DescriptionKind = schema.StringMarkdown
}

func New() func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"private_key_content": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("VAULTED_PRIVATE_KEY_CONTENT", ""),
					Description: "Content of private key used to decrypt `vaulted-tfe_variable` resources. " +
						"This setting has higher priority than `private_key_path`.",
				},
				"private_key_path": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("VAULTED_PRIVATE_KEY_PATH", ""),
					Description: "Path to private key used to decrypt `vaulted-tfe_variable` resources. " +
						"This setting has lower priority than `private_key_content`.",
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"vaulted_content": dataSourceContent(),
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

		privateKey, err := readPrivateKey(d, osExecutor, rsaSvc)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return &MetaClient{VaultedPrivateKey: privateKey}, nil
	}
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
