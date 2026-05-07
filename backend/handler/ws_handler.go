package handler

import (
	"encoding/json"
	"net/http"

	"backend/websocket"

	"github.com/rs/zerolog"
)

type WebSocketHandler struct {
	manager *websocket.Manager
	logger  zerolog.Logger
}

func NewWebSocketHandler(manager *websocket.Manager, logger zerolog.Logger) *WebSocketHandler {
	return &WebSocketHandler{manager: manager, logger: logger}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}

	client := websocket.NewClient(conn, h.manager)
	h.manager.Register(client)

	client.Start()
}

func (h *WebSocketHandler) GetClientCount(w http.ResponseWriter, r *http.Request) {
	count := h.manager.GetClientCount()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{
		"connected_clients": count,
	})
}
