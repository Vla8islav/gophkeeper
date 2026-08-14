package domain

import "github.com/google/uuid"

type CreateSecretParams struct {
	ID      uuid.UUID
	UserID  int64
	Type    SecretType
	Payload []byte
	Meta    []byte
}
