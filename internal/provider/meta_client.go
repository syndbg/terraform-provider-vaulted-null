package provider

import (
	"github.com/palantir/stacktrace"
	"github.com/sumup-oss/vaulted/pkg/vaulted/content"
	"github.com/sumup-oss/vaulted/pkg/vaulted/header"
	"github.com/sumup-oss/vaulted/pkg/vaulted/passphrase"
	"github.com/sumup-oss/vaulted/pkg/vaulted/payload"
)

type MetaClient struct {
	payloadSerializer PayloadSerializer
	payloadEncrypter  PayloadEncrypter

	payloadDeserializer PayloadDeserializer
	payloadDecrypter    PayloadDecrypter
}

func (m *MetaClient) EncryptValue(plaintext string) (string, error) {
	passphraseSvc := passphrase.NewService()

	generatedPassphrase, err := passphraseSvc.GeneratePassphrase(32)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"failed to generate random AES passphrase",
		)
	}

	payloadInstance := payload.NewPayload(
		header.NewHeader(),
		generatedPassphrase,
		content.NewContent([]byte(plaintext)),
	)

	encryptedPayload, err := m.payloadEncrypter.Encrypt(payloadInstance)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"failed to encrypt payload",
		)
	}

	serializedValue, err := m.payloadSerializer.Serialize(encryptedPayload)
	if err != nil {
		return "", stacktrace.Propagate(err, "unable to serialize encrypted payload")
	}

	return string(serializedValue), nil
}

func (m *MetaClient) DecryptValue(encryptedValue string) (string, error) {
	deserializedValue, err := m.payloadDeserializer.Deserialize([]byte(encryptedValue))
	if err != nil {
		return "", stacktrace.Propagate(err, "unable to serialize `value`")
	}

	decryptedValue, err := m.payloadDecrypter.Decrypt(deserializedValue)
	if err != nil {
		return "", stacktrace.Propagate(err, "unable to decrypt `value`")
	}

	plaintext := decryptedValue.Content.Plaintext

	return string(plaintext), nil
}
