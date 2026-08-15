package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func newUpdateRequest(t *testing.T, userID int64, idParam string, body []byte) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/api/secret/update/"+idParam, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := middlewares.ContextWithUserID(req.Context(), userID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", idParam) // this is what chi.URLParam reads
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return req.WithContext(ctx)
}

func TestSecretUpdateHandler_Success(t *testing.T) {
	h, service := newTestHandler(t)
	const userID = int64(42)
	id := uuid.New()

	body, err := json.Marshal(domain.UpdateSecretRequest{Payload: []byte("v2"), Meta: []byte("m2"), Version: 1})
	require.NoError(t, err)

	service.EXPECT().
		UpdateSecret(gomock.Any(), userID, id, gomock.Any()).
		Return(int64(2), nil)

	req := newUpdateRequest(t, userID, id.String(), body)
	w := httptest.NewRecorder()
	h.SecretUpdateHandler(w, req)

	res := w.Result()
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	var resp domain.UpdateSecretResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&resp))
	require.Equal(t, int64(2), resp.Version)
}

func TestSecretUpdateHandler_ServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{"version conflict -> 409", domain.ErrVersionConflict, http.StatusConflict},
		{"not found -> 404", domain.ErrSecretNotFound, http.StatusNotFound},
		{"invalid id -> 400", domain.ErrInvalidSecretID, http.StatusBadRequest},
		{"generic -> 500", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, service := newTestHandler(t)
			const userID = int64(42)
			id := uuid.New()
			body, _ := json.Marshal(domain.UpdateSecretRequest{Payload: []byte("v2"), Version: 1})

			service.EXPECT().
				UpdateSecret(gomock.Any(), userID, id, gomock.Any()).
				Return(int64(0), tt.serviceErr)

			req := newUpdateRequest(t, userID, id.String(), body)
			w := httptest.NewRecorder()
			h.SecretUpdateHandler(w, req)

			require.Equal(t, tt.wantStatus, w.Result().StatusCode)
		})
	}
}

func TestSecretUpdateHandler_RejectsBeforeService(t *testing.T) {
	validBody, _ := json.Marshal(domain.UpdateSecretRequest{Payload: []byte("v2"), Version: 1})

	t.Run("bad content type -> 400", func(t *testing.T) {
		h, _ := newTestHandler(t)
		req := newUpdateRequest(t, 1, uuid.New().String(), validBody)
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		h.SecretUpdateHandler(w, req)
		require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("bad uuid -> 400", func(t *testing.T) {
		h, _ := newTestHandler(t)
		req := newUpdateRequest(t, 1, "not-a-uuid", validBody)
		w := httptest.NewRecorder()
		h.SecretUpdateHandler(w, req)
		require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("malformed json -> 400", func(t *testing.T) {
		h, _ := newTestHandler(t)
		req := newUpdateRequest(t, 1, uuid.New().String(), []byte(`{"version":`))
		w := httptest.NewRecorder()
		h.SecretUpdateHandler(w, req)
		require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("missing user id -> 500", func(t *testing.T) {
		h, _ := newTestHandler(t)
		id := uuid.New()
		req := httptest.NewRequest(http.MethodPut, "/api/secret/update/"+id.String(), bytes.NewReader(validBody))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		h.SecretUpdateHandler(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
	})
}
