package domain

import (
	"context"

	"github.com/google/uuid"
)

type GophkeeperService interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, request UserRegisterRequest) (*AuthResult, error)
	LoginUser(ctx context.Context, request UserLoginRequest) (*AuthResult, error)
	CreateSecret(ctx context.Context, userID int64, req CreateSecretRequest) error
	ListSecrets(ctx context.Context, userID int64) ([]SecretSummary, error)
	UpdateSecret(ctx context.Context, params UpdateSecretParams) (int64, error)
	GetSecret(ctx context.Context, userID int64, id uuid.UUID) (*Secret, error)
}
