package router

import (
	"net/http"

	"backend/handler"
	"backend/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func NewRouter(userHandler *handler.UserHandler, wsHandler *handler.WebSocketHandler, chatRoomHandler *handler.ChatRoomHandler, messageHandler *handler.MessageHandler, logger zerolog.Logger) *chi.Mux {
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
			r.Post("/login", userHandler.Login)
			r.Post("/register", userHandler.CreateUser)

			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(logger))
				r.Get("/", userHandler.GetAllUsers)
				r.Get("/get", userHandler.GetUserByID)
				r.Put("/", userHandler.UpdateUser)
				r.Delete("/", userHandler.DeleteUser)
			})
		})

		r.Route("/chat-rooms", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(logger))

			r.Get("/", chatRoomHandler.GetAllChatRooms)
			r.Post("/", chatRoomHandler.CreateChatRoom)
			r.Get("/{id}", chatRoomHandler.GetChatRoomByID)
			r.Put("/{id}", chatRoomHandler.UpdateChatRoom)
			r.Delete("/{id}", chatRoomHandler.DeleteChatRoom)
			r.Get("/group/{group}", chatRoomHandler.GetChatRoomsByGroup)
			r.Get("/owner/{owner_id}", chatRoomHandler.GetChatRoomsByOwner)
			r.Get("/{id}/token", wsHandler.GetRoomToken)
			r.Get("/{id}/users", wsHandler.GetRoomUsers)
			r.Get("/{id}/messages", messageHandler.GetRoomMessages)
		})

		r.Get("/ws/clients", wsHandler.GetClientCount)

		r.Post("/upload/room-logo", handler.UploadRoomLogo)
	})

	return r
}
