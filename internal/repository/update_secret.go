package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func (s *PostgresStorage) UpdateSecret(ctx context.Context, params domain.UpdateSecretParams) (int64, error) {
	var newVersion int64

	err := s.withRetry(ctx, func() error {
		// Optimistic update: only succeeds if the version still matches.
		err := s.db.QueryRowContext(ctx,
			`UPDATE secrets
                         SET payload = $1, meta = $2, version = version + 1, updated_at = now()
                         WHERE id = $3 AND user_id = $4 AND version = $5 AND deleted = FALSE
                         RETURNING version`,
			params.Payload,
			params.Meta,
			params.ID,
			params.UserID,
			params.Version,
		).Scan(&newVersion)
		if err == nil {
			return nil // updated successfully
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to update secret %s for user %d: %w",
				params.ID, params.UserID, err)
		}

		// No row updated
		var currentVersion int64
		checkErr := s.db.QueryRowContext(ctx,
			`SELECT version FROM secrets
                         WHERE id = $1 AND user_id = $2 AND deleted = FALSE`,
			params.ID,
			params.UserID,
		).Scan(&currentVersion)
		if errors.Is(checkErr, sql.ErrNoRows) {
			return domain.ErrSecretNotFound // doesn't exist/not yours/deleted: 404
		}
		if checkErr != nil {
			return fmt.Errorf("failed to check secret %s for user %d: %w",
				params.ID, params.UserID, checkErr)
		}
		return domain.ErrVersionConflict // exists, but a version was stale: 409
	})

	if err != nil {
		return 0, err
	}

	return newVersion, nil
}
