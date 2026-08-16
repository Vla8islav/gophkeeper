package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/middlewares"
)

func newGetRequest(t *testing.T, userID int64, idParam string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/secret/get/"+idParam, nil)
	ctx := middlewares.ContextWithUserID(req.Context(), userID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", idParam)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

func TestSecretGetHandler_OK(t *testing.T) {
	h, service := newTestHandler(t)
	const userID = int64(42)
	id := uuid.New()
	service.EXPECT().GetSecret(gomock.Any(), userID, id).
		Return(&domain.Secret{ID: id, Type: domain.SecretTypeText, Payload: []byte("cipher"), Version: 1}, nil)

	w := httptest.NewRecorder()
	h.SecretGetHandler(w, newGetRequest(t, userID, id.String()))

	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
	var got domain.GetSecretResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&got))
	require.Equal(t, id, got.ID)
	require.Equal(t, []byte("cipher"), got.Payload)
}

func TestSecretGetHandler_NotFound(t *testing.T) {
	h, service := newTestHandler(t)
	const userID = int64(42)
	id := uuid.New()
	service.EXPECT().GetSecret(gomock.Any(), userID, id).Return(nil, domain.ErrSecretNotFound)

	w := httptest.NewRecorder()
	h.SecretGetHandler(w, newGetRequest(t, userID, id.String()))
	require.Equal(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestSecretGetHandler_BadID(t *testing.T) {
	h, _ := newTestHandler(t) // no EXPECT -> service not reached
	w := httptest.NewRecorder()
	h.SecretGetHandler(w, newGetRequest(t, 1, "not-a-uuid"))
	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestSecretGetHandler_MissingUser(t *testing.T) {
	h, _ := newTestHandler(t)
	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/secret/get/"+id.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.SecretGetHandler(w, req)
	require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}
