package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func (m gophkeeperService) CreateSecret(ctx context.Context, userID int64, req domain.CreateSecretRequest) error {
	if req.ID == uuid.Nil {
		return fmt.Errorf("secret id is required: %w", domain.ErrInvalidSecretType)
	}
	if !req.Type.Valid() {
		return fmt.Errorf("unknown secret type %q: %w", req.Type, domain.ErrInvalidSecretType)
	}

	params := domain.CreateSecretParams{
		ID:      req.ID,
		UserID:  userID,
		Type:    req.Type,
		Payload: req.Payload,
		Meta:    req.Meta,
	}

	if err := m.repository.CreateSecret(ctx, params); err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}
