package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Vla8islav/gophkeeper/internal/audit"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/middlewares"
)

func (h *Handler) SecretsListHandler(w http.ResponseWriter, r *http.Request) {
	audit.SetOperation(r.Context(), "secret.list")
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		h.writeInternalServerError(w, "user id missing from context")
		return
	}

	secrets, err := h.service.ListSecrets(r.Context(), userID)
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	items := make([]domain.SecretSummaryResponse, 0, len(secrets))
	for _, s := range secrets {
		items = append(items, domain.SecretSummaryResponse{
			ID:        s.ID,
			Type:      s.Type,
			Meta:      s.Meta,
			Version:   s.Version,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
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
