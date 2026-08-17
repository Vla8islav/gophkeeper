package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Vla8islav/gophkeeper/internal/audit"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/middlewares"
)

// SecretGetHandler godoc
// @Summary  Get one secret by id
// @Tags     secrets
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "secret UUID"
// @Success  200 {object} domain.GetSecretResponse
// @Failure  400
// @Failure  401
// @Failure  404
// @Failure  500
// @Router   /api/secret/get/{id} [get]
func (h *Handler) SecretGetHandler(w http.ResponseWriter, r *http.Request) {
	audit.SetOperation(r.Context(), "secret.get")
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
	audit.SetSecretID(r.Context(), id.String())

	secret, err := h.service.GetSecret(r.Context(), userID, id)
	if errors.Is(err, domain.ErrInvalidSecretID) {
		h.writeBadRequest(w, err.Error())
		return
	}
	if errors.Is(err, domain.ErrSecretNotFound) {
		h.writeNotFound(w, err.Error())
		return
	}
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	resp := domain.GetSecretResponse{
		ID:        secret.ID,
		Type:      secret.Type,
		Payload:   secret.Payload,
		Meta:      secret.Meta,
		Version:   secret.Version,
		CreatedAt: secret.CreatedAt,
		UpdatedAt: secret.UpdatedAt,
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}
