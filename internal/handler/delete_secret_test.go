package handler

import (
	"context"
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

func newDeleteRequest(t *testing.T, userID int64, idParam string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodDelete,
		"/api/secret/delete/"+idParam, nil)
	ctx := middlewares.ContextWithUserID(req.Context(), userID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", idParam)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

func TestSecretDeleteHandler_NoContent(t *testing.T) {
	h, service := newTestHandler(t)
	const userID = int64(42)
	id := uuid.New()

	service.EXPECT().DeleteSecret(gomock.Any(), userID, id).Return(nil)

	req := newDeleteRequest(t, userID, id.String())
	w := httptest.NewRecorder()
	h.SecretDeleteHandler(w, req)

	require.Equal(t, http.StatusNoContent, w.Result().StatusCode)
}

func TestSecretDeleteHandler_NotFound(t *testing.T) {
	h, service := newTestHandler(t)
	const userID = int64(42)
	id := uuid.New()

	service.EXPECT().
		DeleteSecret(gomock.Any(), userID, id).
		Return(domain.ErrSecretNotFound)

	req := newDeleteRequest(t, userID, id.String())
	w := httptest.NewRecorder()
	h.SecretDeleteHandler(w, req)

	require.Equal(t, http.StatusNotFound, w.Result().StatusCode)
}

func TestSecretDeleteHandler_BadID(t *testing.T) {
	h, _ := newTestHandler(t) // no EXPECT

	req := newDeleteRequest(t, 1, "not-a-uuid")
	w := httptest.NewRecorder()
	h.SecretDeleteHandler(w, req)

	require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestSecretDeleteHandler_MissingUserID(t *testing.T) {
	h, _ := newTestHandler(t) // no EXPECT

	id := uuid.New()
	req := httptest.NewRequest(http.MethodDelete,
		"/api/secret/delete/"+id.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.SecretDeleteHandler(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}
