package provider

import "github.com/sumup-oss/vaulted/pkg/vaulted/payload"

type PayloadDecrypter interface {
	Decrypt(encryptedPayload *payload.EncryptedPayload) (*payload.Payload, error)
}
