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

func TestAPIClient_UpdateSecret_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPut, r.Method)
		require.Equal(t, "Bearer tok", r.Header.Get("Authorization"))

		var req domain.UpdateSecretRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.Equal(t, int64(1), req.Version)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(domain.UpdateSecretResponse{Version: 2})
	}))
	defer srv.Close()

	api, err := NewAPIClient(srv.URL, "")
	require.NoError(t, err)

	v, err := api.UpdateSecret("tok", uuid.New(), domain.UpdateSecretRequest{
		Payload: []byte("cipher"), Version: 1,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), v)
}

func TestAPIClient_UpdateSecret_Conflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer srv.Close()

	api, err := NewAPIClient(srv.URL, "")
	require.NoError(t, err)

	_, err = api.UpdateSecret("tok", uuid.New(), domain.UpdateSecretRequest{Version: 1})
	require.Error(t, err)
}
