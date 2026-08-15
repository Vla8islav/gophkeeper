package localstore

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore_SaltRoundTrip(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "store.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })

	// Nothing cached yet.
	_, err = s.Salt()
	require.ErrorIs(t, err, ErrNotCached)

	salt := []byte{0x01, 0x02, 0x03, 0x04}
	require.NoError(t, s.SaveSalt(salt))

	got, err := s.Salt()
	require.NoError(t, err)
	require.Equal(t, salt, got)

	// Upsert overwrites.
	salt2 := []byte{0x05, 0x06, 0x07, 0x08}
	require.NoError(t, s.SaveSalt(salt2))
	got, err = s.Salt()
	require.NoError(t, err)
	require.Equal(t, salt2, got)
}

func TestStore_PersistsAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.db")

	s, err := Open(path)
	require.NoError(t, err)
	require.NoError(t, s.SaveSalt([]byte("abc")))
	require.NoError(t, s.Close())

	// Reopen the same file — a cache must survive process restarts.
	s2, err := Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = s2.Close() })

	got, err := s2.Salt()
	require.NoError(t, err)
	require.Equal(t, []byte("abc"), got)
}
