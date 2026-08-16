package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/localstore"
)

func TestFetchSalt_CacheHit(t *testing.T) {
	store, err := localstore.Open(filepath.Join(t.TempDir(), "s.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.SaveSalt([]byte("cached-salt")))

	cfg := &config.OptionsClient{}
	salt, err := fetchSalt(cfg, store, "tok")
	require.NoError(t, err)
	require.Equal(t, []byte("cached-salt"), salt)
}

func TestFetchSalt_CacheMissFetchesAndPersists(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
		_ = json.NewEncoder(w).Encode(domain.SaltResponse{Salt: []byte("server-salt")})
	}))
	defer srv.Close()

	store, err := localstore.Open(filepath.Join(t.TempDir(), "s.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	cfg := &config.OptionsClient{ServerAddress: config.OptionalString{Value: srv.URL}}
	salt, err := fetchSalt(cfg, store, "tok")
	require.NoError(t, err)
	require.Equal(t, []byte("server-salt"), salt)

	cached, err := store.Salt()
	require.NoError(t, err)
	require.Equal(t, []byte("server-salt"), cached)
}
