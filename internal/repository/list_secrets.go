package repository

import (
	"context"
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func (s *PostgresStorage) ListSecrets(ctx context.Context, userID int64) ([]domain.SecretSummary, error) {
	var summaries []domain.SecretSummary

	err := s.withRetry(ctx, func() error {
		rows, err := s.db.QueryContext(ctx,
			`SELECT id, type, meta, version, created_at, updated_at
                         FROM secrets
                         WHERE user_id = $1 AND deleted = FALSE
                         ORDER BY created_at`,
			userID,
		)
		if err != nil {
			return fmt.Errorf("failed to list secrets for user %d: %w", userID, err)
		}
		defer rows.Close()

		// Build into a fresh local slice so a retry can't duplicate rows.
		var result []domain.SecretSummary
		for rows.Next() {
			var item domain.SecretSummary
			if err := rows.Scan(
				&item.ID,
				&item.Type,
				&item.Meta,
				&item.Version,
				&item.CreatedAt,
				&item.UpdatedAt,
			); err != nil {
				return fmt.Errorf("failed to scan secret for user %d: %w", userID, err)
			}
			result = append(result, item)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("row iteration failed for user %d: %w", userID, err)
		}

		summaries = result
		return nil
	})

	if err != nil {
		return nil, err
	}

	return summaries, nil
}
