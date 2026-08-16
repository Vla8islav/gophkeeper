package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func TestPostgresStorage_SyncSecrets(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "sync-secrets")

	live, gone := uuid.New(), uuid.New()
	require.NoError(t, storage.CreateSecret(ctx,
		domain.CreateSecretParams{ID: live, UserID: userID, Type: domain.SecretTypeText, Payload: []byte("a")}))
	require.NoError(t, storage.CreateSecret(ctx,
		domain.CreateSecretParams{ID: gone, UserID: userID, Type: domain.SecretTypeText, Payload: []byte("b")}))
	require.NoError(t, storage.DeleteSecret(ctx, userID, gone)) // tombstone

	other := createTestUser(t, ctx, storage, "sync-other")
	require.NoError(t, storage.CreateSecret(ctx,
		domain.CreateSecretParams{ID: uuid.New(), UserID: other, Type: domain.SecretTypeText, Payload: []byte("x")}))

	secrets, err := storage.SyncSecrets(ctx, userID)
	require.NoError(t, err)
	require.Len(t, secrets, 2) // includes the tombstone

	byID := map[uuid.UUID]domain.Secret{}
	for _, s := range secrets {
		byID[s.ID] = s
	}
	require.False(t, byID[live].Deleted)
	require.True(t, byID[gone].Deleted) // tombstones are included
}

func TestPostgresStorage_SyncSecrets_Empty(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "sync-empty")
	secrets, err := storage.SyncSecrets(ctx, userID)
	require.NoError(t, err)
	require.Empty(t, secrets)
}
