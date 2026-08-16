package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDBPing_OK(t *testing.T) {
	h, service := newTestHandler(t)
	service.EXPECT().Ping(gomock.Any()).Return(nil)

	w := httptest.NewRecorder()
	h.DBPing(w, httptest.NewRequest(http.MethodGet, "/api/ping", nil))
	require.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestDBPing_Error(t *testing.T) {
	h, service := newTestHandler(t)
	service.EXPECT().Ping(gomock.Any()).Return(errors.New("db down"))

	w := httptest.NewRecorder()
	h.DBPing(w, httptest.NewRequest(http.MethodGet, "/api/ping", nil))
	require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestDBPing_MethodNotAllowed(t *testing.T) {
	h, _ := newTestHandler(t)

	w := httptest.NewRecorder()
	h.DBPing(w, httptest.NewRequest(http.MethodPost, "/api/ping", nil))
	require.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
}
