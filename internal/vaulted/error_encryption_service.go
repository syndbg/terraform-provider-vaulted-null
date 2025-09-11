package vaulted

import (
	"errors"

	"github.com/sumup-oss/vaulted/pkg/vaulted/payload"
)

type ErrorEncryptionService struct {
	errorMsg string
}

func NewErrorEncryptionService(errorMsg string) *ErrorEncryptionService {
	return &ErrorEncryptionService{
		errorMsg: errorMsg,
	}
}

func (e *ErrorEncryptionService) Encrypt(payload *payload.Payload) (*payload.EncryptedPayload, error) {
	return nil, errors.New(e.errorMsg)
}
