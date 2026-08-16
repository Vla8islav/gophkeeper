package repository

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

func (s *PostgresStorage) DeleteSecret(ctx context.Context, userID int64, id uuid.UUID) error {
	err := s.withRetry(ctx, func() error {
		res, err := s.db.ExecContext(ctx,
			`UPDATE secrets
                         SET deleted = TRUE, version = version + 1, updated_at = now()
                         WHERE id = $1 AND user_id = $2 AND deleted = FALSE`,
			id, userID,
		)
		if err != nil {
			return fmt.Errorf("failed to delete secret %s for user %d: %w", id, userID, err)
		}

		affected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("rows affected for secret %s: %w", id, err)
		}
		if affected == 0 {
			return domain.ErrSecretNotFound
		}
		return nil
	})

	return err
}
