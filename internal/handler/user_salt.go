package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Vla8islav/gophkeeper/internal/audit"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/middlewares"
)

func (h *Handler) UserSaltHandler(w http.ResponseWriter, r *http.Request) {
	audit.SetOperation(r.Context(), "user.salt")
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		h.writeInternalServerError(w, "user id missing from context")
		return
	}

	salt, err := h.service.GetUserSalt(r.Context(), userID)
	if err != nil {
		h.writeInternalServerError(w, err.Error()) // ErrSaltNotSet / ErrUserNotFound are server-side anomalies - 500
		return
	}

	payload, err := json.Marshal(domain.SaltResponse{Salt: salt})
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}
