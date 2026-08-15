package service

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func (m gophkeeperService) GetUserSalt(ctx context.Context, userID int64) ([]byte, error) {
	salt, err := m.repository.GetUserSalt(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user salt: %w", err)
	}
	if len(salt) == 0 {
		return nil, domain.ErrSaltNotSet // legacy user with NULL salt — an anomaly
	}
	return salt, nil
}
