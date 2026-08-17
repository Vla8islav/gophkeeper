package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/Vla8islav/gophkeeper/internal/audit"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/repository"
)

// UserLoginHandler godoc
// @Summary  Authenticate (returns a JWT)
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    request body domain.UserLoginRequest true "login + password"
// @Success  200 {object} domain.UserLoginResponse
// @Failure  400
// @Failure  401
// @Failure  500
// @Router   /api/user/login [post]
func (h *Handler) UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	audit.SetOperation(r.Context(), "user.login")

	if r.Method != http.MethodPost {
		h.writeMethodNotAllowed(w, "only POST method is allowed")
		return
	}

	mimeType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mimeType != "application/json" {
		h.writeBadRequest(w, "only application/json content type is supported")
		return
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeBadRequest(w, "failed to read request body: "+err.Error())
		return
	}

	var requestBodySerialised domain.UserLoginRequest
	err = json.Unmarshal(requestBody, &requestBodySerialised)
	if err != nil {
		h.writeBadRequest(w, "couldn't parse requestBody:"+err.Error())
		return
	}

	if requestBodySerialised.Login == "" {
		h.writeBadRequest(w, "login cannot be empty")
		return
	}

	if requestBodySerialised.Password == "" {
		h.writeBadRequest(w, "password cannot be empty")
		return
	}

	authResult, err := h.service.LoginUser(r.Context(), requestBodySerialised)
	if errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, domain.ErrInvalidUserCredentials) {
		h.writeUnauthorised(w, "invalid user login or password")
		return
	}

	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	loginResponse := domain.UserLoginResponse{Token: authResult.Token}
	payload, err := json.Marshal(loginResponse)
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(payload)
	if err != nil {
		return
	}
}
