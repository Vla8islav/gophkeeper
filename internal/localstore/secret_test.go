package localstore

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// openTestStore gives each test a fresh SQLite file in its own temp dir.
func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "store.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// Happy path
func TestStore_GetSecret_RoundTrip(t *testing.T) {
	s := openTestStore(t)

	sec := Secret{
		ID: "abc", Type: "text",
		Payload: []byte{0x00, 0x01, 0xff}, Meta: []byte("m"),
		Version: 2,
	}
	require.NoError(t, s.SaveSecret(sec))

	got, err := s.GetSecret("abc")
	require.NoError(t, err)
	require.Equal(t, sec.ID, got.ID)
	require.Equal(t, sec.Type, got.Type)
	require.Equal(t, sec.Payload, got.Payload) // binary round-trips, null byte survives
	require.Equal(t, sec.Meta, got.Meta)
	require.Equal(t, int64(2), got.Version)
	require.False(t, got.Deleted)
}

// A missing id is ErrNotCached
func TestStore_GetSecret_NotCached(t *testing.T) {
	s := openTestStore(t)
	_, err := s.GetSecret("nope")
	require.ErrorIs(t, err, ErrNotCached)
}

// A soft-deleted secret must read as absent
func TestStore_GetSecret_SoftDeletedIsHidden(t *testing.T) {
	s := openTestStore(t)
	require.NoError(t, s.SaveSecret(Secret{
		ID: "gone", Type: "text", Payload: []byte("x"), Version: 1, Deleted: true,
	}))
	_, err := s.GetSecret("gone")
	require.ErrorIs(t, err, ErrNotCached) // deleted reads as absent
}

func TestStore_ListSecrets(t *testing.T) {
	s := openTestStore(t)
	require.NoError(t, s.SaveSecret(Secret{ID: "a", Type: "text", Payload: []byte("p"), Version: 1}))
	require.NoError(t, s.SaveSecret(Secret{ID: "b", Type: "card", Payload: []byte("p"), Version: 1}))
	require.NoError(t, s.SaveSecret(Secret{ID: "gone", Type: "text", Payload: []byte("p"), Version: 1, Deleted: true}))

	list, err := s.ListSecrets()
	require.NoError(t, err)
	require.Len(t, list, 2) // deleted one excluded

	ids := map[string]bool{}
	for _, sec := range list {
		ids[sec.ID] = true
	}
	require.True(t, ids["a"])
	require.True(t, ids["b"])
	require.False(t, ids["gone"]) // tombstone not listed
}

func TestStore_ListSecrets_Empty(t *testing.T) {
	s := openTestStore(t)
	list, err := s.ListSecrets()
	require.NoError(t, err)
	require.Empty(t, list)
}
