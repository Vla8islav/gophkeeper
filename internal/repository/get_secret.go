package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func (s *PostgresStorage) GetSecret(ctx context.Context, userID int64, id uuid.UUID) (*domain.Secret, error) {
	var secret domain.Secret

	err := s.withRetry(ctx, func() error {
		err := s.db.QueryRowContext(ctx,
			`SELECT id, user_id, type, payload, meta, version, deleted, created_at, updated_at
                         FROM secrets
                         WHERE id = $1 AND user_id = $2 AND deleted = FALSE`,
			id,
			userID,
		).Scan(
			&secret.ID,
			&secret.UserID,
			&secret.Type,
			&secret.Payload,
			&secret.Meta,
			&secret.Version,
			&secret.Deleted,
			&secret.CreatedAt,
			&secret.UpdatedAt,
		)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrSecretNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get secret %s for user %d: %w", id, userID, err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &secret, nil
}
