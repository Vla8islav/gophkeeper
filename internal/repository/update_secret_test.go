package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func TestPostgresStorage_UpdateSecret(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "update-secret")

	id := uuid.New()
	require.NoError(t, storage.CreateSecret(ctx, domain.CreateSecretParams{
		ID: id, UserID: userID, Type: domain.SecretTypeText,
		Payload: []byte("v1"), Meta: []byte("m1"),
	}))

	newVersion, err := storage.UpdateSecret(ctx, domain.UpdateSecretParams{
		ID: id, UserID: userID, Payload: []byte("v2"), Meta: []byte("m2"), Version: 1,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), newVersion) // 1 - 2

	got, err := storage.GetSecret(ctx, userID, id)
	require.NoError(t, err)
	require.Equal(t, []byte("v2"), got.Payload)
	require.Equal(t, []byte("m2"), got.Meta)
	require.Equal(t, int64(2), got.Version)
	require.Equal(t, domain.SecretTypeText, got.Type) // type is immutable
}

func TestPostgresStorage_UpdateSecret_VersionConflict(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "update-conflict")

	id := uuid.New()
	require.NoError(t, storage.CreateSecret(ctx, domain.CreateSecretParams{
		ID: id, UserID: userID, Type: domain.SecretTypeText, Payload: []byte("v1"),
	}))

	_, err := storage.UpdateSecret(ctx, domain.UpdateSecretParams{
		ID: id, UserID: userID, Payload: []byte("v2"), Version: 99, // stale
	})
	require.ErrorIs(t, err, domain.ErrVersionConflict)

	// The write mustn't happen
	got, err := storage.GetSecret(ctx, userID, id)
	require.NoError(t, err)
	require.Equal(t, []byte("v1"), got.Payload)
	require.Equal(t, int64(1), got.Version)
}

func TestPostgresStorage_UpdateSecret_NotFound(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "update-notfound")

	_, err := storage.UpdateSecret(ctx, domain.UpdateSecretParams{
		ID: uuid.New(), UserID: userID, Payload: []byte("v2"), Version: 1,
	})
	require.ErrorIs(t, err, domain.ErrSecretNotFound)
}

func TestPostgresStorage_UpdateSecret_WrongOwnerIsNotFoundNotConflict(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	owner := createTestUser(t, ctx, storage, "update-owner")
	other := createTestUser(t, ctx, storage, "update-other")

	id := uuid.New()
	require.NoError(t, storage.CreateSecret(ctx, domain.CreateSecretParams{
		ID: id, UserID: owner, Type: domain.SecretTypeText, Payload: []byte("v1"),
	}))

	_, err := storage.UpdateSecret(ctx, domain.UpdateSecretParams{
		ID: id, UserID: other, Payload: []byte("hacked"), Version: 1,
	})
	require.ErrorIs(t, err, domain.ErrSecretNotFound)

	got, err := storage.GetSecret(ctx, owner, id)
	require.NoError(t, err)
	require.Equal(t, []byte("v1"), got.Payload)
}

func TestPostgresStorage_UpdateSecret_SoftDeletedIsNotFound(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "update-deleted")

	id := uuid.New()
	require.NoError(t, storage.CreateSecret(ctx, domain.CreateSecretParams{
		ID: id, UserID: userID, Type: domain.SecretTypeText, Payload: []byte("v1"),
	}))
	_, err := storage.db.ExecContext(ctx,
		`UPDATE secrets SET deleted = TRUE WHERE id = $1`, id)
	require.NoError(t, err)

	_, err = storage.UpdateSecret(ctx, domain.UpdateSecretParams{
		ID: id, UserID: userID, Payload: []byte("v2"), Version: 1,
	})
	require.ErrorIs(t, err, domain.ErrSecretNotFound)
}

func TestPostgresStorage_UpdateSecret_SequentialVersions(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "update-sequential")

	id := uuid.New()
	require.NoError(t, storage.CreateSecret(ctx, domain.CreateSecretParams{
		ID: id, UserID: userID, Type: domain.SecretTypeText, Payload: []byte("v1"),
	}))

	v, err := storage.UpdateSecret(ctx, domain.UpdateSecretParams{ID: id, UserID: userID, Payload: []byte("v2"), Version: 1})
	require.NoError(t, err)
	require.Equal(t, int64(2), v)

	v, err = storage.UpdateSecret(ctx, domain.UpdateSecretParams{ID: id, UserID: userID, Payload: []byte("v3"), Version: 2})
	require.NoError(t, err)
	require.Equal(t, int64(3), v)

	_, err = storage.UpdateSecret(ctx, domain.UpdateSecretParams{ID: id, UserID: userID, Payload: []byte("v4"), Version: 2})
	require.ErrorIs(t, err, domain.ErrVersionConflict)
}
