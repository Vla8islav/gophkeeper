package service

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func (m gophkeeperService) SyncSecrets(ctx context.Context, userID int64) ([]domain.Secret, error) {
	secrets, err := m.repository.SyncSecrets(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to sync secrets: %w", err)
	}
	return secrets, nil
}
