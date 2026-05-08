package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/dto"
	"backend/middleware"
	"backend/service"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

type ChatRoomHandler struct {
	service service.ChatRoomService
	logger  zerolog.Logger
}

func NewChatRoomHandler(svc service.ChatRoomService, logger zerolog.Logger) *ChatRoomHandler {
	return &ChatRoomHandler{service: svc, logger: logger}
}

func (h *ChatRoomHandler) GetChatRoomByID(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logger := h.logger.With().Str("request_id", requestID).Logger()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Warn().Str("id", idStr).Msg("Invalid chat room ID")
		h.respondWithError(w, http.StatusBadRequest, "Invalid chat room ID")
		return
	}

	logger.Debug().Int("id", id).Msg("Getting chat room by ID")
	room, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		switch err {
		case service.ErrChatRoomNotFound:
			h.respondWithError(w, http.StatusNotFound, "Chat room not found")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Failed to get chat room")
		}
		return
	}

	logger.Info().Int("id", id).Str("name", room.Name).Msg("Chat room retrieved successfully")
	h.respondWithData(w, http.StatusOK, room)
}

func (h *ChatRoomHandler) GetAllChatRooms(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logger := h.logger.With().Str("request_id", requestID).Logger()

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	logger.Debug().Int("page", page).Int("pageSize", pageSize).Msg("Getting all chat rooms")
	result, err := h.service.GetAll(r.Context(), page, pageSize)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get chat rooms")
		return
	}

	h.respondWithData(w, http.StatusOK, result)
}

func (h *ChatRoomHandler) GetChatRoomsByGroup(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logger := h.logger.With().Str("request_id", requestID).Logger()

	group := chi.URLParam(r, "group")
	if group == "" {
		logger.Warn().Msg("Group parameter is required")
		h.respondWithError(w, http.StatusBadRequest, "Group parameter is required")
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
		pageSize = 10
	}

	logger.Debug().Str("group", group).Int("page", page).Int("pageSize", pageSize).Msg("Getting chat rooms by group")
	result, err := h.service.GetByGroup(r.Context(), group, page, pageSize)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get chat rooms by group")
		return
	}

	h.respondWithData(w, http.StatusOK, result)
}

func (h *ChatRoomHandler) GetChatRoomsByOwner(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logger := h.logger.With().Str("request_id", requestID).Logger()

	ownerIDStr := chi.URLParam(r, "owner_id")
	ownerID, err := strconv.Atoi(ownerIDStr)
	if err != nil || ownerID <= 0 {
		logger.Warn().Str("owner_id", ownerIDStr).Msg("Invalid owner ID")
		h.respondWithError(w, http.StatusBadRequest, "Invalid owner ID")
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
		pageSize = 10
	}

	logger.Debug().Int("owner_id", ownerID).Int("page", page).Int("pageSize", pageSize).Msg("Getting chat rooms by owner")
	result, err := h.service.GetByOwner(r.Context(), ownerID, page, pageSize)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get chat rooms by owner")
		return
	}

	h.respondWithData(w, http.StatusOK, result)
}

func (h *ChatRoomHandler) CreateChatRoom(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logger := h.logger.With().Str("request_id", requestID).Logger()

	var req dto.CreateChatRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn().Err(err).Msg("Invalid request body for create chat room")
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	logger.Debug().Str("name", req.Name).Msg("Creating new chat room")
	room, err := h.service.Create(r.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrChatRoomNameExists:
			h.respondWithError(w, http.StatusConflict, "Chat room name already exists")
		case service.ErrInvalidInput:
			h.respondWithError(w, http.StatusBadRequest, "Invalid input")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Failed to create chat room")
		}
		return
	}

	logger.Info().Int("id", room.ID).Str("name", room.Name).Msg("Chat room created successfully")
	h.respondWithData(w, http.StatusCreated, room)
}

func (h *ChatRoomHandler) UpdateChatRoom(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logger := h.logger.With().Str("request_id", requestID).Logger()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Warn().Str("id", idStr).Msg("Invalid chat room ID")
		h.respondWithError(w, http.StatusBadRequest, "Invalid chat room ID")
		return
	}

	var req dto.UpdateChatRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn().Err(err).Msg("Invalid request body for update chat room")
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	logger.Debug().Int("id", id).Msg("Updating chat room")
	room, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrChatRoomNotFound:
			h.respondWithError(w, http.StatusNotFound, "Chat room not found")
		case service.ErrChatRoomNameExists:
			h.respondWithError(w, http.StatusConflict, "Chat room name already exists")
		case service.ErrInvalidInput:
			h.respondWithError(w, http.StatusBadRequest, "Invalid input")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Failed to update chat room")
		}
		return
	}

	logger.Info().Int("id", id).Str("name", room.Name).Msg("Chat room updated successfully")
	h.respondWithData(w, http.StatusOK, room)
}

func (h *ChatRoomHandler) DeleteChatRoom(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logger := h.logger.With().Str("request_id", requestID).Logger()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Warn().Str("id", idStr).Msg("Invalid chat room ID")
		h.respondWithError(w, http.StatusBadRequest, "Invalid chat room ID")
		return
	}

	logger.Debug().Int("id", id).Msg("Deleting chat room")
	err = h.service.Delete(r.Context(), id)
	if err != nil {
		switch err {
		case service.ErrChatRoomNotFound:
			h.respondWithError(w, http.StatusNotFound, "Chat room not found")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Failed to delete chat room")
		}
		return
	}

	logger.Info().Int("id", id).Msg("Chat room deleted successfully")
	w.WriteHeader(http.StatusNoContent)
}

func (h *ChatRoomHandler) respondWithData(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *ChatRoomHandler) respondWithError(w http.ResponseWriter, statusCode int, errorMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": errorMessage})
}
