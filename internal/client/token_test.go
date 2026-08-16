package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveAndLoadToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "token") // nested dir needed

	require.NoError(t, saveToken(path, "my-token"))

	got, err := loadToken(path)
	require.NoError(t, err)
	require.Equal(t, "my-token", got)

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestLoadToken_Missing(t *testing.T) {
	_, err := loadToken(filepath.Join(t.TempDir(), "does-not-exist"))
	require.Error(t, err)
}
