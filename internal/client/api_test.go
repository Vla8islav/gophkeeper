package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func TestAPIClient_RegisterAndLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		_ = json.NewEncoder(w).Encode(domain.UserLoginResponse{Token: "tok-123"})
	}))
	defer srv.Close()
	api, err := NewAPIClient(srv.URL, "")
	require.NoError(t, err)

	tok, err := api.Register("alice", "pw")
	require.NoError(t, err)
	require.Equal(t, "tok-123", tok)

	tok, err = api.Login("alice", "pw")
	require.NoError(t, err)
	require.Equal(t, "tok-123", tok)
}

func TestAPIClient_Register_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer srv.Close()
	api, _ := NewAPIClient(srv.URL, "")
	_, err := api.Register("alice", "pw")
	require.Error(t, err)
}

func TestAPIClient_ListSecrets(t *testing.T) {
	id := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
		_ = json.NewEncoder(w).Encode([]domain.SecretSummaryResponse{{ID: id, Type: domain.SecretTypeText, Version: 1}})
	}))
	defer srv.Close()
	api, _ := NewAPIClient(srv.URL, "")
	list, err := api.ListSecrets("tok")
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, id, list[0].ID)
}

func TestAPIClient_ListSecrets_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	api, _ := NewAPIClient(srv.URL, "")
	_, err := api.ListSecrets("tok")
	require.Error(t, err)
}

func TestAPIClient_CreateSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var req domain.CreateSecretRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	api, _ := NewAPIClient(srv.URL, "")
	err := api.CreateSecret("tok", domain.CreateSecretRequest{
		ID: uuid.New(), Type: domain.SecretTypeText, Payload: []byte("c"),
	})
	require.NoError(t, err)
}

func TestAPIClient_CreateSecret_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer srv.Close()
	api, _ := NewAPIClient(srv.URL, "")
	err := api.CreateSecret("tok", domain.CreateSecretRequest{ID: uuid.New(), Type: domain.SecretTypeText, Payload: []byte("c")})
	require.Error(t, err)
}

func TestAPIClient_GetUserSalt(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(domain.SaltResponse{Salt: []byte("sixteen-byte-slt")})
	}))
	defer srv.Close()
	api, _ := NewAPIClient(srv.URL, "")
	salt, err := api.GetUserSalt("tok")
	require.NoError(t, err)
	require.Equal(t, []byte("sixteen-byte-slt"), salt)
}

func TestAPIClient_DeleteSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	api, _ := NewAPIClient(srv.URL, "")
	require.NoError(t, api.DeleteSecret("tok", uuid.New()))
}

func TestAPIClient_DeleteSecret_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	api, _ := NewAPIClient(srv.URL, "")
	require.Error(t, api.DeleteSecret("tok", uuid.New()))
}

func TestAPIClient_SyncSecrets(t *testing.T) {
	id := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]domain.SyncSecretResponse{{ID: id, Version: 2, Deleted: true}})
	}))
	defer srv.Close()
	api, _ := NewAPIClient(srv.URL, "")
	items, err := api.SyncSecrets("tok")
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.True(t, items[0].Deleted)
}

func TestNewAPIClient_BadCAPath(t *testing.T) {
	_, err := NewAPIClient("https://example.com", "/no/such/ca.pem")
	require.Error(t, err)
}
