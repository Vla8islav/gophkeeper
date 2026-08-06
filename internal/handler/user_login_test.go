package handler

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/mocks"
	"github.com/Vla8islav/gophkeeper/internal/repository"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func newTestLoginHandler(service domain.GophkeeperService) *Handler {
	return &Handler{
		service: service,
		logger:  zap.NewNop(),
	}
}

func TestUserLoginHandler_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophkeeperService(ctrl)
	service.EXPECT().
		LoginUser(gomock.Any(), domain.UserLoginRequest{
			Login:    "test-login",
			Password: "test-password",
		}).
		Return(&domain.AuthResult{
			UserID: 123,
			Token:  "test-token",
		}, nil)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)

	cookies := res.Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "auth_token", cookies[0].Name)
	require.Equal(t, "test-token", cookies[0].Value)
	require.True(t, cookies[0].HttpOnly)
	require.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)
	require.Equal(t, "/", cookies[0].Path)
}

func TestUserLoginHandler_AllowsJSONContentTypeWithCharset(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophkeeperService(ctrl)
	service.EXPECT().
		LoginUser(gomock.Any(), domain.UserLoginRequest{
			Login:    "test-login",
			Password: "test-password",
		}).
		Return(&domain.AuthResult{UserID: 123, Token: "test-token"}, nil)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestUserLoginHandler_MethodNotAllowed(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophkeeperService(ctrl)

	h := newTestLoginHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/api/user/login", nil)
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
}

func TestUserLoginHandler_BadContentType(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophkeeperService(ctrl)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUserLoginHandler_InvalidJSON(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophkeeperService(ctrl)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUserLoginHandler_EmptyLogin(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophkeeperService(ctrl)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUserLoginHandler_EmptyPassword(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophkeeperService(ctrl)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestUserLoginHandler_InvalidCredentials(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophkeeperService(ctrl)
	service.EXPECT().
		LoginUser(gomock.Any(), domain.UserLoginRequest{
			Login:    "test-login",
			Password: "wrong-password",
		}).
		Return(nil, repository.ErrUserNotFound)

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"wrong-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestUserLoginHandler_ServiceError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := mocks.NewMockGophkeeperService(ctrl)
	service.EXPECT().
		LoginUser(gomock.Any(), domain.UserLoginRequest{
			Login:    "test-login",
			Password: "test-password",
		}).
		Return(nil, errors.New("service error"))

	h := newTestLoginHandler(service)

	body := bytes.NewBufferString(`{"login":"test-login","password":"test-password"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UserLoginHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
