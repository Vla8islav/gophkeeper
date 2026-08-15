package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func (s *PostgresStorage) CreateSecret(ctx context.Context, params domain.CreateSecretParams) error {
	err := s.withRetry(ctx, func() error {
		var returnedID uuid.UUID
		err := s.db.QueryRowContext(ctx,
			`INSERT INTO secrets (id, user_id, type, payload, meta)
                         VALUES ($1, $2, $3, $4, $5)
                         ON CONFLICT (id) DO NOTHING
                         RETURNING id`,
			params.ID,
			params.UserID,
			params.Type,
			params.Payload,
			params.Meta,
		).Scan(&returnedID)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrSecretAlreadyExists
		}
		if err != nil {
			return fmt.Errorf("failed to create secret %s for user %d: %w", params.ID, params.UserID, err)
		}
		return nil
	})

	return err
}
