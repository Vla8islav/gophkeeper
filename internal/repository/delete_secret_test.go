package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func TestPostgresStorage_DeleteSecret(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "delete-secret")

	id := uuid.New()
	require.NoError(t, storage.CreateSecret(ctx, domain.CreateSecretParams{
		ID: id, UserID: userID, Type: domain.SecretTypeText, Payload: []byte("v1"),
	}))

	require.NoError(t, storage.DeleteSecret(ctx, userID, id))

	_, err := storage.GetSecret(ctx, userID, id)
	require.ErrorIs(t, err, domain.ErrSecretNotFound)

	var deleted bool
	var version int64
	err = storage.db.QueryRowContext(ctx,
		`SELECT deleted, version FROM secrets WHERE id = $1`, id,
	).Scan(&deleted, &version)
	require.NoError(t, err)
	require.True(t, deleted)
	require.Equal(t, int64(2), version)
}

func TestPostgresStorage_DeleteSecret_NotFound(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "delete-notfound")
	require.ErrorIs(t, storage.DeleteSecret(ctx, userID, uuid.New()), domain.ErrSecretNotFound)
}

func TestPostgresStorage_DeleteSecret_WrongOwnerIsNotFound(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	owner := createTestUser(t, ctx, storage, "delete-owner")
	other := createTestUser(t, ctx, storage, "delete-other")

	id := uuid.New()
	require.NoError(t, storage.CreateSecret(ctx, domain.CreateSecretParams{
		ID: id, UserID: owner, Type: domain.SecretTypeText, Payload: []byte("v1"),
	}))

	// non-owner can't delete it
	require.ErrorIs(t, storage.DeleteSecret(ctx, other, id), domain.ErrSecretNotFound)
	got, err := storage.GetSecret(ctx, owner, id)
	require.NoError(t, err)
	require.Equal(t, id, got.ID)
}

func TestPostgresStorage_DeleteSecret_AlreadyDeleted(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "delete-twice")

	id := uuid.New()
	require.NoError(t, storage.CreateSecret(ctx, domain.CreateSecretParams{
		ID: id, UserID: userID, Type: domain.SecretTypeText, Payload: []byte("v1"),
	}))

	require.NoError(t, storage.DeleteSecret(ctx, userID, id))
	require.ErrorIs(t, storage.DeleteSecret(ctx, userID, id), domain.ErrSecretNotFound)
}
