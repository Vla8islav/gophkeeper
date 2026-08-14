package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/middlewares"
	"github.com/Vla8islav/gophkeeper/internal/mocks"
)

func newTestHandler(t *testing.T) (*Handler, *mocks.MockGophkeeperService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	service := mocks.NewMockGophkeeperService(ctrl)
	h := &Handler{service: service, logger: zap.NewNop()}
	return h, service
}

func validCreateBody(t *testing.T) []byte {
	t.Helper()
	b, err := json.Marshal(domain.CreateSecretRequest{
		ID:      uuid.New(),
		Type:    domain.SecretTypeText,
		Payload: []byte("cipher"),
	})
	require.NoError(t, err)
	return b
}

func TestSecretCreateHandler_Created(t *testing.T) {
	h, service := newTestHandler(t)
	const userID = int64(42)

	// Asserting the 2nd arg == userID proves the handler pulled it from context
	service.EXPECT().
		CreateSecret(gomock.Any(), userID, gomock.Any()).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/secrets", bytes.NewReader(validCreateBody(t)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middlewares.ContextWithUserID(req.Context(), userID))

	w := httptest.NewRecorder()
	h.SecretCreateHandler(w, req)

	require.Equal(t, http.StatusCreated, w.Result().StatusCode)
}

func TestSecretCreateHandler_ServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
	}{
		{"invalid type -> 400", domain.ErrInvalidSecretType, http.StatusBadRequest},
		{"already exists -> 409", domain.ErrSecretAlreadyExists, http.StatusConflict},
		{"generic error -> 500", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, service := newTestHandler(t)
			const userID = int64(7)

			service.EXPECT().
				CreateSecret(gomock.Any(), userID, gomock.Any()).
				Return(tt.serviceErr)

			req := httptest.NewRequest(http.MethodPost, "/api/secrets", bytes.NewReader(validCreateBody(t)))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(middlewares.ContextWithUserID(req.Context(), userID))

			w := httptest.NewRecorder()
			h.SecretCreateHandler(w, req)

			require.Equal(t, tt.wantStatus, w.Result().StatusCode)
		})
	}
}

func TestSecretCreateHandler_RejectsBeforeService(t *testing.T) {
	// if the handler reaches the service - fail

	t.Run("wrong content type -> 400", func(t *testing.T) {
		h, _ := newTestHandler(t)
		req := httptest.NewRequest(http.MethodPost, "/api/secrets", bytes.NewReader(validCreateBody(t)))
		req.Header.Set("Content-Type", "text/plain")
		req = req.WithContext(middlewares.ContextWithUserID(req.Context(), 1))

		w := httptest.NewRecorder()
		h.SecretCreateHandler(w, req)
		require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("malformed json -> 400", func(t *testing.T) {
		h, _ := newTestHandler(t)
		req := httptest.NewRequest(http.MethodPost, "/api/secrets", bytes.NewReader([]byte(`{"type":`)))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(middlewares.ContextWithUserID(req.Context(), 1))

		w := httptest.NewRecorder()
		h.SecretCreateHandler(w, req)
		require.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("missing user id in context -> 500", func(t *testing.T) {
		h, _ := newTestHandler(t)
		req := httptest.NewRequest(http.MethodPost, "/api/secrets", bytes.NewReader(validCreateBody(t)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		h.SecretCreateHandler(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
	})
}
