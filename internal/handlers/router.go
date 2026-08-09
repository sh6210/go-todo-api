package handlers

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(todoHandler *TodoHandler) *chi.Mux {
	r := chi.NewRouter()

	//r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(RequestIDMiddleware)

	r.Route("/todos", func(r chi.Router) {
		r.Post("/", todoHandler.Create)
		r.Get("/", todoHandler.GetAll)
		r.Get("/{id}", todoHandler.GetByID)
		r.Put("/{id}", todoHandler.Update)
		r.Delete("/{id}", todoHandler.Delete)
	})
	return r
}

func NewRouterWithLogger(todoHandler *TodoHandler, appLogger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(RequestIDMiddleware)
	r.Use(LoggingMiddleware(appLogger))

	r.Route("/todos", func(r chi.Router) {
		r.Post("/", todoHandler.Create)
		r.Get("/", todoHandler.GetAll)
		r.Get("/{id}", todoHandler.GetByID)
		r.Put("/{id}", todoHandler.Update)
		r.Delete("/{id}", todoHandler.Delete)
	})
	return r
}
