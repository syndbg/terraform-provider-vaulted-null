package provider

import "github.com/sumup-oss/vaulted/pkg/vaulted/payload"

type PayloadDeserializer interface {
	Deserialize(encodedContent []byte) (*payload.EncryptedPayload, error)
}
