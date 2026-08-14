package handler

import (
	"net/http"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"go.uber.org/zap"
)

type Handler struct {
	service domain.GophkeeperService
	logger  *zap.Logger
}

func NewHandler(service domain.GophkeeperService, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) writeUnauthorised(w http.ResponseWriter, msg string) {
	h.logger.Info("unauthorised", zap.String("msg", msg))
	http.Error(w, msg, http.StatusUnauthorized)
}

func (h *Handler) writeAlreadyExists(w http.ResponseWriter, msg string) {
	h.logger.Info("already exists", zap.String("msg", msg))
	http.Error(w, msg, http.StatusConflict)
}

func (h *Handler) writeInternalServerError(w http.ResponseWriter, msg string) {
	h.logger.Error("internal server error", zap.String("msg", msg))
	http.Error(w, msg, http.StatusInternalServerError)
}

func (h *Handler) writeBadRequest(w http.ResponseWriter, msg string) {
	h.logger.Info("bad request", zap.String("msg", msg))
	http.Error(w, msg, http.StatusBadRequest)
}

func (h *Handler) writeNotFound(w http.ResponseWriter, msg string) {
	h.logger.Info("not found request", zap.String("msg", msg))
	http.Error(w, msg, http.StatusNotFound)
}

func (h *Handler) writePaymentRequired(w http.ResponseWriter, msg string) {
	h.logger.Info("not enough money", zap.String("msg", msg))
	http.Error(w, msg, http.StatusPaymentRequired)
}

func (h *Handler) writeUnprocessableEntity(w http.ResponseWriter, msg string) {
	h.logger.Info("incorrect request checksum", zap.String("msg", msg))
	http.Error(w, msg, http.StatusUnprocessableEntity)
}

func (h *Handler) writeMethodNotAllowed(w http.ResponseWriter, msg string) {
	h.logger.Info("method not allowed", zap.String("msg", msg))
	http.Error(w, msg, http.StatusMethodNotAllowed)
}

func (h *Handler) writeNoContent(w http.ResponseWriter, msg string) {
	h.logger.Info("found no content", zap.String("msg", msg))
	w.WriteHeader(http.StatusNoContent)
}
