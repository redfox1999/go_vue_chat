package router

import (
	"net/http"

	"backend/handler"
	"backend/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func NewRouter(userHandler *handler.UserHandler, wsHandler *handler.WebSocketHandler, logger zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "message": "Server is running"}`))
	})

	r.Get("/ws", wsHandler.HandleWebSocket)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.RequestID)
		r.Use(middleware.RequestLogger(logger))

		r.Route("/users", func(r chi.Router) {
			r.Get("/", userHandler.GetAllUsers)
			r.Post("/", userHandler.CreateUser)
			r.Get("/get", userHandler.GetUserByID)
			r.Put("/", userHandler.UpdateUser)
			r.Delete("/", userHandler.DeleteUser)
			r.Post("/login", userHandler.Login)
		})

		r.Get("/ws/clients", wsHandler.GetClientCount)
	})

	return r
}
