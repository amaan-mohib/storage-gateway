package http

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/go-chi/render"
)

func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)
	r.Use(middleware.StripSlashes)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(httprate.LimitByIP(1000, 1*time.Minute))
	r.Use(cors.AllowAll().Handler)
	r.Use(middleware.Heartbeat("/health"))

	r.Get("/{bucket}/*", h.Download)
	r.With(AuthMiddleware).Post("/{bucket}/*", h.Upload)
	r.With(AuthMiddleware).Delete("/{bucket}/*", h.Delete)

	return r
}
