package repository

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

// SyncSecrets returns  of a user's secrets, including deleted ones
func (s *PostgresStorage) SyncSecrets(ctx context.Context, userID int64) ([]domain.Secret, error) {
	var secrets []domain.Secret

	err := s.withRetry(ctx, func() error {
		rows, err := s.db.QueryContext(ctx,
			`SELECT id, user_id, type, payload, meta, version, deleted, created_at, updated_at
                         FROM secrets
                         WHERE user_id = $1
                         ORDER BY updated_at`,
			userID,
		)
		if err != nil {
			return fmt.Errorf("failed to sync secrets for user %d: %w", userID, err)
		}
		defer rows.Close()

		var result []domain.Secret
		for rows.Next() {
			var sec domain.Secret
			if err := rows.Scan(
				&sec.ID, &sec.UserID, &sec.Type, &sec.Payload, &sec.Meta,
				&sec.Version, &sec.Deleted, &sec.CreatedAt, &sec.UpdatedAt,
			); err != nil {
				return fmt.Errorf("failed to scan secret for user %d: %w", userID, err)
			}
			result = append(result, sec)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("row iteration failed for user %d: %w", userID, err)
		}

		secrets = result
		return nil
	})

	if err != nil {
		return nil, err
	}
	return secrets, nil
}
