package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/middlewares"
)

func TestSecretSyncHandler_OK(t *testing.T) {
	h, service := newTestHandler(t)
	const userID = int64(42)

	id1, id2 := uuid.New(), uuid.New()
	service.EXPECT().SyncSecrets(gomock.Any(), userID).Return([]domain.Secret{
		{ID: id1, Type: domain.SecretTypeText, Payload: []byte("cipher"), Version: 1},
		{ID: id2, Type: domain.SecretTypeText, Version: 2, Deleted: true}, // tombstone
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/secret/sync", nil)
	req = req.WithContext(middlewares.ContextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()
	h.SecretSyncHandler(w, req)

	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
	var got []domain.SyncSecretResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
	require.Len(t, got, 2)
	require.True(t, got[1].Deleted) // tombstone reaches the wire
}

func TestSecretSyncHandler_EmptyIsJSONArray(t *testing.T) {
	h, service := newTestHandler(t)
	const userID = int64(42)
	service.EXPECT().SyncSecrets(gomock.Any(), userID).Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/secret/sync", nil)
	req = req.WithContext(middlewares.ContextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()
	h.SecretSyncHandler(w, req)

	res := w.Result()
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, "[]", string(body))
}

func TestSecretSyncHandler_MissingUser(t *testing.T) {
	h, _ := newTestHandler(t)
	w := httptest.NewRecorder()
	h.SecretSyncHandler(w, httptest.NewRequest(http.MethodGet, "/api/secret/sync", nil))
	require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}
