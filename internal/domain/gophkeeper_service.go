package domain

import (
	"context"
)

type GophkeeperService interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, request UserRegisterRequest) (*AuthResult, error)
	CreateSecret(ctx context.Context, userID int64, req CreateSecretRequest) error
	LoginUser(ctx context.Context, request UserLoginRequest) (*AuthResult, error)
}
