package repository

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func TestPostgresStorage_ListSecrets(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userA := createTestUser(t, ctx, storage, "list-secrets-a")
	userB := createTestUser(t, ctx, storage, "list-secrets-b")

	// three secrets for A
	want := make(map[uuid.UUID]domain.CreateSecretParams)
	for i := 0; i < 3; i++ {
		p := domain.CreateSecretParams{
			ID:      uuid.New(),
			UserID:  userA,
			Type:    domain.SecretTypeText,
			Payload: []byte("payload-should-not-leak"),
			Meta:    []byte(fmt.Sprintf("meta-%d", i)),
		}
		require.NoError(t, storage.CreateSecret(ctx, p))
		want[p.ID] = p
	}

	// one secret for B - must NOT appear in A's list
	bSecret := domain.CreateSecretParams{
		ID:      uuid.New(),
		UserID:  userB,
		Type:    domain.SecretTypeCard,
		Payload: []byte("b-payload"),
	}
	require.NoError(t, storage.CreateSecret(ctx, bSecret))

	got, err := storage.ListSecrets(ctx, userA)
	require.NoError(t, err)
	require.Len(t, got, 3)

	// Membership check, NOT positional: ORDER BY created_at can tie at microsecond
	// resolution for rapid inserts, so asserting a strict order would be flaky.
	byID := make(map[uuid.UUID]domain.SecretSummary, len(got))
	for _, s := range got {
		byID[s.ID] = s
	}
	for id, p := range want {
		summary, ok := byID[id]
		require.True(t, ok, "secret %s missing from list", id)
		require.Equal(t, p.Type, summary.Type)
		require.Equal(t, p.Meta, summary.Meta) // metadata round-trips
		require.Equal(t, int64(1), summary.Version)
		require.False(t, summary.CreatedAt.IsZero())
	}
	require.NotContains(t, byID, bSecret.ID) // isolation: B's secret is filtered out
}

func TestPostgresStorage_ListSecrets_ExcludesSoftDeleted(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "list-deleted")

	live := domain.CreateSecretParams{ID: uuid.New(), UserID: userID, Type: domain.SecretTypeText, Payload: []byte("live")}
	gone := domain.CreateSecretParams{ID: uuid.New(), UserID: userID, Type: domain.SecretTypeText, Payload: []byte("gone")}
	require.NoError(t, storage.CreateSecret(ctx, live))
	require.NoError(t, storage.CreateSecret(ctx, gone))

	_, err := storage.db.ExecContext(ctx, `UPDATE secrets SET deleted = TRUE WHERE id = $1`, gone.ID)
	require.NoError(t, err)

	got, err := storage.ListSecrets(ctx, userID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, live.ID, got[0].ID)
}

func TestPostgresStorage_ListSecrets_Empty(t *testing.T) {
	storage, ctx := newSecretStorage(t)
	userID := createTestUser(t, ctx, storage, "list-empty")

	got, err := storage.ListSecrets(ctx, userID)
	require.NoError(t, err)
	require.Empty(t, got) // empty list is NOT an error (no 404 semantics)
}
