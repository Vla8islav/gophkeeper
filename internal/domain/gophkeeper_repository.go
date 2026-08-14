package domain

import (
	"context"
)

type GophkeeperRepository interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, user CreateUserParams) (int64, error)
	CreateSecret(ctx context.Context, user CreateSecretParams) error
	GetUserByLogin(ctx context.Context, login string) (*User, error)
}
