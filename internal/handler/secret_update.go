package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/middlewares"
)

func (h *Handler) SecretUpdateHandler(w http.ResponseWriter, r *http.Request) {
	mimeType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mimeType != "application/json" {
		h.writeBadRequest(w, "only application/json content type is supported")
		return
	}

	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		h.writeInternalServerError(w, "user id missing from context")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeBadRequest(w, "invalid secret id: "+err.Error())
		return
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeBadRequest(w, "failed to read request body: "+err.Error())
		return
	}

	var req domain.UpdateSecretRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		h.writeBadRequest(w, "couldn't parse request body: "+err.Error())
		return
	}

	newVersion, err := h.service.UpdateSecret(r.Context(), userID, id, req)
	if errors.Is(err, domain.ErrInvalidSecretID) {
		h.writeBadRequest(w, err.Error())
		return
	}
	if errors.Is(err, domain.ErrSecretNotFound) {
		h.writeNotFound(w, err.Error())
		return
	}
	if errors.Is(err, domain.ErrVersionConflict) {
		h.writeConflict(w, err.Error())
		return
	}
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	payload, err := json.Marshal(domain.UpdateSecretResponse{Version: newVersion})
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}
