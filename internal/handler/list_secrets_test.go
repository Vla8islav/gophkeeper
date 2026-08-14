package handler

import (
	"encoding/json"
	"errors"
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

func TestListSecretsHandler_Success(t *testing.T) {
	h, service := newTestHandler(t)
	const userID = int64(42)

	id1, id2 := uuid.New(), uuid.New()
	service.EXPECT().
		ListSecrets(gomock.Any(), userID). // pins userID → proves context extraction + forwarding
		Return([]domain.SecretSummary{
			{ID: id1, Type: domain.SecretTypeText, Meta: []byte("m1"), Version: 1},
			{ID: id2, Type: domain.SecretTypeCard, Version: 2},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/secret/list", nil)
	req = req.WithContext(middlewares.ContextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	h.SecretsListHandler(w, req)

	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var got []domain.SecretSummaryResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
	require.Len(t, got, 2)
	require.Equal(t, id1, got[0].ID)
	require.Equal(t, domain.SecretTypeText, got[0].Type)
	require.Equal(t, []byte("m1"), got[0].Meta)
	require.Equal(t, id2, got[1].ID)
	require.Equal(t, int64(2), got[1].Version)
}

func TestListSecretsHandler_EmptyReturnsJSONArray(t *testing.T) {
	h, service := newTestHandler(t)
	const userID = int64(42)

	service.EXPECT().
		ListSecrets(gomock.Any(), userID).
		Return(nil, nil) // empty vault → repo/service return nil

	req := httptest.NewRequest(http.MethodGet, "/api/secret/list", nil)
	req = req.WithContext(middlewares.ContextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	h.SecretsListHandler(w, req)

	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	// The crux: a nil service result must serialize as [] , NOT null.
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, "[]", string(body))
}

func TestListSecretsHandler_MissingUserID(t *testing.T) {
	h, _ := newTestHandler(t)

	// No EXPECT → service must never be called when context lacks a user id.
	req := httptest.NewRequest(http.MethodGet, "/api/secret/list", nil)
	// deliberately no ContextWithUserID
	w := httptest.NewRecorder()

	h.SecretsListHandler(w, req)
	require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestListSecretsHandler_ServiceError(t *testing.T) {
	h, service := newTestHandler(t)
	const userID = int64(42)

	service.EXPECT().
		ListSecrets(gomock.Any(), userID).
		Return(nil, errors.New("db down"))

	req := httptest.NewRequest(http.MethodGet, "/api/secret/list", nil)
	req = req.WithContext(middlewares.ContextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	h.SecretsListHandler(w, req)
	require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}
