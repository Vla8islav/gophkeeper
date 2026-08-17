package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (s *PostgresStorage) GetUserSalt(ctx context.Context, userID int64) ([]byte, error) {
	var salt []byte
	err := s.withRetry(ctx, func() error {
		err := s.db.QueryRowContext(ctx,
			`SELECT kdf_salt FROM users WHERE id = $1`,
			userID,
		).Scan(&salt)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get salt for user %d: %w", userID, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return salt, nil // may be nil for a legacy (pre-migration) user
}
