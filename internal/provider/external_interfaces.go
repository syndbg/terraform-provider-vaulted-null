package provider

import "github.com/sumup-oss/vaulted/pkg/vaulted/payload"

type PayloadEncrypter interface {
	Encrypt(payload *payload.Payload) (*payload.EncryptedPayload, error)
}

type PayloadSerializer interface {
	Serialize(encryptedPayload *payload.EncryptedPayload) ([]byte, error)
}

type PayloadDecrypter interface {
	Decrypt(encryptedPayload *payload.EncryptedPayload) (*payload.Payload, error)
}

type PayloadDeserializer interface {
	Deserialize(encodedContent []byte) (*payload.EncryptedPayload, error)
}
