package service

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

func (m gophkeeperService) GetSecret(ctx context.Context, userID int64, id uuid.UUID) (*domain.Secret, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("secret id is required: %w", domain.ErrInvalidSecretID)
	}

	secret, err := m.repository.GetSecret(ctx, userID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	return secret, nil
}
