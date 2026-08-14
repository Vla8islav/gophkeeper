package domain

import "github.com/google/uuid"

type CreateSecretParams struct {
	ID      uuid.UUID
	UserID  int64
	Type    SecretType
	Payload []byte
	Meta    []byte
}

// no UserID — that comes from the token.
type CreateSecretRequest struct {
	ID      uuid.UUID  `json:"id"`
	Type    SecretType `json:"type"`
	Payload []byte     `json:"payload"`
	Meta    []byte     `json:"meta,omitempty"`
}
