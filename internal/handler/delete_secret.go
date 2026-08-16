package handler

import (
	"errors"
	"net/http"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) SecretDeleteHandler(w http.ResponseWriter, r *http.Request) {
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

	err = h.service.DeleteSecret(r.Context(), userID, id)
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

	w.WriteHeader(http.StatusNoContent)
}
