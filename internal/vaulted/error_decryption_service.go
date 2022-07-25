package vaulted

import (
	"errors"

	"github.com/sumup-oss/vaulted/pkg/vaulted/payload"
)

type ErrorDecryptionService struct {
	errorMsg string
}

func NewErrorDecryptionService(errorMsg string) *ErrorDecryptionService {
	return &ErrorDecryptionService{}
}

func (e *ErrorDecryptionService) Decrypt(encryptedPayload *payload.EncryptedPayload) (*payload.Payload, error) {
	return nil, errors.New(e.errorMsg)
}
