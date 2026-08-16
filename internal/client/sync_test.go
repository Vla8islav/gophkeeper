package client

import (
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/localstore"
)

func TestReconcile(t *testing.T) {
	store, err := localstore.Open(filepath.Join(t.TempDir(), "store.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	oldID := uuid.New()
	require.NoError(t, store.SaveSecret(localstore.Secret{
		ID: oldID.String(), Type: "text", Payload: []byte("old"), Version: 1,
	}))

	newID := uuid.New()
	items := []domain.SyncSecretResponse{
		{ID: newID, Type: domain.SecretTypeText, Payload: []byte("new"), Version: 1},
		{ID: oldID, Version: 2, Deleted: true},
	}

	pulled, removed, err := reconcile(store, items)
	require.NoError(t, err)
	require.Equal(t, 1, pulled)
	require.Equal(t, 1, removed)

	got, err := store.GetSecret(newID.String())
	require.NoError(t, err)
	require.Equal(t, []byte("new"), got.Payload)

	_, err = store.GetSecret(oldID.String())
	require.ErrorIs(t, err, localstore.ErrNotCached)
}

func TestReconcile_Empty(t *testing.T) {
	store, err := localstore.Open(filepath.Join(t.TempDir(), "store.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	pulled, removed, err := reconcile(store, nil)
	require.NoError(t, err)
	require.Zero(t, pulled)
	require.Zero(t, removed)
}
