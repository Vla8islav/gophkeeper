package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/Vla8islav/gophkeeper/internal/audit"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/middlewares"
)

// SecretCreateHandler godoc
// @Summary  Create an encrypted secret
// @Tags     secrets
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    request body domain.CreateSecretRequest true "encrypted secret (payload/meta are ciphertext)"
// @Success  201
// @Failure  400
// @Failure  401
// @Failure  409
// @Failure  500
// @Router   /api/secret/create [post]
func (h *Handler) SecretCreateHandler(w http.ResponseWriter, r *http.Request) {
	audit.SetOperation(r.Context(), "secret.create")

	mimeType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mimeType != "application/json" {
		h.writeBadRequest(w, "only application/json content type is supported")
		return
	}

	// Owner is injected by WithAuth via token
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		// a wiring bug guard
		h.writeInternalServerError(w, "user id missing from context")
		return
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeBadRequest(w, "failed to read request body: "+err.Error())
		return
	}

	var req domain.CreateSecretRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		h.writeBadRequest(w, "couldn't parse request body: "+err.Error())
		return
	}
	audit.SetSecretID(r.Context(), req.ID.String())

	err = h.service.CreateSecret(r.Context(), userID, req)
	if errors.Is(err, domain.ErrInvalidSecretType) {
		h.writeBadRequest(w, err.Error())
		return
	}
	if errors.Is(err, domain.ErrSecretAlreadyExists) {
		h.writeAlreadyExists(w, err.Error())
		return
	}
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}
