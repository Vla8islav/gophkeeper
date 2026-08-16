package domain

import (
	"context"

	"github.com/google/uuid"
)

type GophkeeperRepository interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, user CreateUserParams) (int64, error)
	CreateSecret(ctx context.Context, user CreateSecretParams) error
	GetUserByLogin(ctx context.Context, login string) (*User, error)
	ListSecrets(ctx context.Context, userID int64) ([]SecretSummary, error)
	UpdateSecret(ctx context.Context, params UpdateSecretParams) (int64, error)
	GetSecret(ctx context.Context, userID int64, id uuid.UUID) (*Secret, error)
	GetUserSalt(ctx context.Context, userID int64) ([]byte, error)
	DeleteSecret(ctx context.Context, userID int64, id uuid.UUID) error
}
