package domain

import (
	"time"

	"github.com/google/uuid"
)

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

type Secret struct {
	ID        uuid.UUID
	UserID    int64
	Type      SecretType
	Payload   []byte
	Meta      []byte
	Version   int64
	Deleted   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SecretSummary struct {
	ID        uuid.UUID
	Type      SecretType
	Meta      []byte
	Version   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GetSecretResponse struct {
	ID        uuid.UUID  `json:"id"`
	Type      SecretType `json:"type"`
	Payload   []byte     `json:"payload"`
	Meta      []byte     `json:"meta,omitempty"`
	Version   int64      `json:"version"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
type SecretSummaryResponse struct {
	ID        uuid.UUID  `json:"id"`
	Type      SecretType `json:"type"`
	Meta      []byte     `json:"meta,omitempty"`
	Version   int64      `json:"version"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type UpdateSecretParams struct {
	ID      uuid.UUID
	UserID  int64
	Payload []byte
	Meta    []byte
	Version int64
}

type UpdateSecretRequest struct {
	Payload []byte `json:"payload"`
	Meta    []byte `json:"meta,omitempty"`
	Version int64  `json:"version"`
}
