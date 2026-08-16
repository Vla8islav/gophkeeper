package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Vla8islav/gophkeeper/internal/audit"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/middlewares"
)

// SecretSyncHandler godoc
// @Summary  Full sync feed (all secrets incl. tombstones, with ciphertext)
// @Tags     secrets
// @Produce  json
// @Security BearerAuth
// @Success  200 {array} domain.SyncSecretResponse
// @Failure  401
// @Failure  500
// @Router   /api/secret/sync [get]
func (h *Handler) SecretSyncHandler(w http.ResponseWriter, r *http.Request) {
	audit.SetOperation(r.Context(), "secret.sync")
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		h.writeInternalServerError(w, "user id missing from context")
		return
	}

	secrets, err := h.service.SyncSecrets(r.Context(), userID)
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	items := make([]domain.SyncSecretResponse, 0, len(secrets))
	for _, s := range secrets {
		items = append(items, domain.SyncSecretResponse{
			ID: s.ID, Type: s.Type, Payload: s.Payload, Meta: s.Meta,
			Version: s.Version, Deleted: s.Deleted, UpdatedAt: s.UpdatedAt,
		})
	}

	payload, err := json.Marshal(items)
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}
