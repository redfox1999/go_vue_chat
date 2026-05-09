package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"backend/config"
	"backend/repository"
	"backend/websocket"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type WebSocketHandler struct {
	manager  *websocket.Manager
	logger   zerolog.Logger
	userRepo repository.UserRepository
}

func NewWebSocketHandler(manager *websocket.Manager, logger zerolog.Logger, userRepo repository.UserRepository) *WebSocketHandler {
	return &WebSocketHandler{manager: manager, logger: logger, userRepo: userRepo}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 解析 JWT token（优先从 Authorization header，其次从 URL 参数）
	var userId int
	var tokenString string

	// 尝试从 Authorization header 获取
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString = parts[1]
		}
	}

	// 如果 header 中没有，尝试从 URL 参数获取
	if tokenString == "" {
		tokenString = r.URL.Query().Get("token")
	}

	// 解析 token
	if tokenString == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "token is required"})
		return
	}

	claims, err := config.ParseJWT(tokenString)
	if err != nil || claims == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid token"})
		return
	}

	userId = claims.UserID
	if userId <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID"})
		return
	}

	// 从数据库获取用户昵称
	nickName := ""
	user, err := h.userRepo.GetByID(r.Context(), userId)
	if err != nil || user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}
	nickName = user.Nickname

	conn, err := websocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}
	h.logger.Info().Msg(fmt.Sprintf("WebSocket connection established for user %d", userId))

	client := websocket.NewClient(conn, h.manager, userId, nickName)
	h.manager.Register(client)

	go client.Start()
}

func (h *WebSocketHandler) GetClientCount(w http.ResponseWriter, r *http.Request) {
	count := h.manager.GetClientCount()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{
		"connected_clients": count,
	})
}

func (h *WebSocketHandler) GetRoomToken(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "id")
	if roomID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "room id is required"})
		return
	}

	token, ok := h.manager.GetRoomToken(roomID)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "room not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"room_id": roomID,
		"token":   token.NewToken,
	})
}

func (h *WebSocketHandler) GetRoomUsers(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "id")
	if roomID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "room id is required"})
		return
	}

	users := h.manager.GetRoomUsers(roomID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"room_id": roomID,
		"users":   users,
	})
}
