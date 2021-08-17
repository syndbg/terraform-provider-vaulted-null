package provider

import "github.com/palantir/stacktrace"

type MetaClient struct {
	payloadDeserializer PayloadDeserializer
	payloadDecrypter    PayloadDecrypter
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
