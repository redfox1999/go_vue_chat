package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/middleware"
	"backend/service"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type MessageHandler struct {
	service service.MessageService
	logger  zerolog.Logger
}

func NewMessageHandler(svc service.MessageService, logger zerolog.Logger) *MessageHandler {
	return &MessageHandler{service: svc, logger: logger}
}

func (h *MessageHandler) GetRoomMessages(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logger := h.logger.With().Str("request_id", requestID).Logger()

	roomIDStr := chi.URLParam(r, "id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil || roomID <= 0 {
		logger.Warn().Str("room_id", roomIDStr).Msg("Invalid room ID")
		h.respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		pageSize = 20
	}

	logger.Debug().Int("room_id", roomID).Int("page", page).Int("pageSize", pageSize).Msg("Getting room messages")
	result, err := h.service.GetByRoomID(r.Context(), roomID, page, pageSize)
	if err != nil {
		logger.Error().Err(err).Int("room_id", roomID).Msg("Failed to get room messages")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get room messages")
		return
	}

	h.respondWithData(w, http.StatusOK, result)
}

func (h *MessageHandler) respondWithData(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *MessageHandler) respondWithError(w http.ResponseWriter, statusCode int, errorMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": errorMessage})
}
