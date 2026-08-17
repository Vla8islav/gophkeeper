package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/localstore"
)

func testClientCfg(t *testing.T, serverURL string) *config.OptionsClient {
	t.Helper()
	dir := t.TempDir()
	return &config.OptionsClient{
		ServerAddress: config.OptionalString{Value: serverURL},
		TokenFile:     config.OptionalString{Value: filepath.Join(dir, "token")},
	}
}

// setStdin replaces os.Stdin with a pipe holding input, for one interactive read.
func setStdin(t *testing.T, input string) {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	_, _ = w.WriteString(input)
	_ = w.Close()
	old := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = old; _ = r.Close() })
}

func TestRunLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(domain.UserLoginResponse{Token: "tok-xyz"})
	}))
	defer srv.Close()
	cfg := testClientCfg(t, srv.URL)
	setStdin(t, "master-pw\n") // login reads the password once

	require.NoError(t, runLogin(cfg, []string{"alice"}))

	tok, err := loadToken(cfg.TokenFile.Value)
	require.NoError(t, err)
	require.Equal(t, "tok-xyz", tok)
}

func TestRunLogin_BadArgs(t *testing.T) {
	require.Error(t, runLogin(testClientCfg(t, ""), nil))
}

func TestRunList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]domain.SecretSummaryResponse{
			{ID: uuid.New(), Type: domain.SecretTypeText, Version: 1},
		})
	}))
	defer srv.Close()
	cfg := testClientCfg(t, srv.URL)
	require.NoError(t, saveToken(cfg.TokenFile.Value, "tok"))
	require.NoError(t, runList(cfg, nil))
}

func TestRunList_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()
	cfg := testClientCfg(t, srv.URL)
	require.NoError(t, saveToken(cfg.TokenFile.Value, "tok"))
	require.NoError(t, runList(cfg, nil))
}

func TestRunSync(t *testing.T) {
	id := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]domain.SyncSecretResponse{
			{ID: id, Type: domain.SecretTypeText, Payload: []byte("cipher"), Version: 1},
		})
	}))
	defer srv.Close()
	cfg := testClientCfg(t, srv.URL)
	require.NoError(t, saveToken(cfg.TokenFile.Value, "tok"))

	require.NoError(t, runSync(cfg, nil))

	// The pulled secret should be in the local store.
	store, err := localstore.Open(dbPath(cfg))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	sec, err := store.GetSecret(id.String())
	require.NoError(t, err)
	require.Equal(t, []byte("cipher"), sec.Payload)
}

func TestRunDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	cfg := testClientCfg(t, srv.URL)
	require.NoError(t, saveToken(cfg.TokenFile.Value, "tok"))

	require.NoError(t, runDelete(cfg, []string{uuid.New().String()}))
}

func TestRunDelete_BadID(t *testing.T) {
	cfg := testClientCfg(t, "")
	require.NoError(t, saveToken(cfg.TokenFile.Value, "tok"))
	require.Error(t, runDelete(cfg, []string{"not-a-uuid"}))
}
