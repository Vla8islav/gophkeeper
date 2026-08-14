package service

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func (m gophkeeperService) ListSecrets(ctx context.Context, userID int64) ([]domain.SecretSummary, error) {

	secrets, err := m.repository.ListSecrets(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return secrets, nil
}
