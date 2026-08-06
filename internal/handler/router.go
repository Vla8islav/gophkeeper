package handler

import (
	"net/http"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler, cfg *config.OptionsServer) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)

	r.Get("/api/ping", h.DBPing)
	r.Post("/api/user/register", h.UserRegisterHandler)
	r.Post("/api/user/login", h.UserLoginHandler)

	r.Group(func(r chi.Router) {

		r.Use(middlewares.WithAuth([]byte(cfg.AuthTokenSecret.Value)))

	})

	return r
}
