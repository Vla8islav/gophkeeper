package service

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

func (m gophkeeperService) DeleteSecret(ctx context.Context, userID int64, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("secret id is required: %w", domain.ErrInvalidSecretID)
	}

	err := m.repository.DeleteSecret(ctx, userID, id)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}
