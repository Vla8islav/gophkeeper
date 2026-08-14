package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/domain"
)

// If you already have an equivalent storage-setup helper, reuse it and drop this.
func newSecretStorage(t *testing.T) (*PostgresStorage, context.Context) {
	t.Helper()
	cfg, err := config.ReadFlagsServer(nil, zap.NewNop())
	require.NoError(t, err)
	return InitTestPostgresStorage(t, cfg), context.Background()
}

func TestPostgresStorage_GetSecret(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "get-secret-test")

	params := domain.CreateSecretParams{
		ID:      uuid.New(),
		UserID:  userID,
		Type:    domain.SecretTypeCard,
		Payload: []byte{0x00, 0x10, 0xff}, // binary incl. null byte
		Meta:    []byte("meta"),
	}
	require.NoError(t, storage.CreateSecret(ctx, params))

	got, err := storage.GetSecret(ctx, userID, params.ID)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.Equal(t, params.ID, got.ID)
	require.Equal(t, params.UserID, got.UserID)
	require.Equal(t, params.Type, got.Type)
	require.Equal(t, params.Payload, got.Payload) // byte-for-byte through SELECT
	require.Equal(t, params.Meta, got.Meta)
	require.Equal(t, int64(1), got.Version)
	require.False(t, got.Deleted)
	require.False(t, got.CreatedAt.IsZero())
	require.False(t, got.UpdatedAt.IsZero())
}

func TestPostgresStorage_GetSecret_NotFoundForWrongOwner(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	owner := createTestUser(t, ctx, storage, "get-secret-owner")
	other := createTestUser(t, ctx, storage, "get-secret-other")

	params := domain.CreateSecretParams{
		ID:      uuid.New(),
		UserID:  owner,
		Type:    domain.SecretTypeText,
		Payload: []byte("secret"),
	}
	require.NoError(t, storage.CreateSecret(ctx, params))

	// The isolation check: another user must NOT be able to read owner's secret,
	// and must get the SAME error as if it didn't exist (no ownership leak).
	got, err := storage.GetSecret(ctx, other, params.ID)
	require.ErrorIs(t, err, domain.ErrSecretNotFound)
	require.Nil(t, got)
}

func TestPostgresStorage_GetSecret_NotFoundForNonexistent(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "get-secret-missing")

	got, err := storage.GetSecret(ctx, userID, uuid.New()) // never inserted
	require.ErrorIs(t, err, domain.ErrSecretNotFound)
	require.Nil(t, got)
}

func TestPostgresStorage_GetSecret_NotFoundWhenSoftDeleted(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "get-secret-deleted")

	params := domain.CreateSecretParams{
		ID:      uuid.New(),
		UserID:  userID,
		Type:    domain.SecretTypeText,
		Payload: []byte("secret"),
	}
	require.NoError(t, storage.CreateSecret(ctx, params))

	// No DeleteSecret method yet — soft-delete directly to test the WHERE clause.
	_, err := storage.db.ExecContext(ctx,
		`UPDATE secrets SET deleted = TRUE WHERE id = $1`, params.ID)
	require.NoError(t, err)

	got, err := storage.GetSecret(ctx, userID, params.ID)
	require.ErrorIs(t, err, domain.ErrSecretNotFound)
	require.Nil(t, got)
}
