package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/dto"
	"backend/middleware"
	"backend/service"

	"github.com/rs/zerolog"
)

type UserHandler struct {
	service service.UserService
	logger  zerolog.Logger
}

func NewUserHandler(svc service.UserService, logger zerolog.Logger) *UserHandler {
	return &UserHandler{service: svc, logger: logger}
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logger := h.logger.With().Str("request_id", requestID).Logger()

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Warn().Str("id", idStr).Msg("Invalid user ID parameter")
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	logger.Debug().Int("id", id).Msg("Getting user by ID")
	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			h.respondWithError(w, http.StatusNotFound, "User not found")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Failed to get user")
		}
		return
	}

	logger.Info().Int("id", id).Msg("User retrieved successfully")
	h.respondWithData(w, http.StatusOK, user)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
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

	logger.Debug().Int("page", page).Int("pageSize", pageSize).Msg("Getting users with pagination")
	result, err := h.service.GetAllUsers(r.Context(), page, pageSize)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	h.respondWithData(w, http.StatusOK, result)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logger := h.logger.With().Str("request_id", requestID).Logger()

	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn().Err(err).Msg("Invalid request body for create user")
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	logger.Debug().Str("email", req.Email).Msg("Creating new user")
	user, err := h.service.CreateUser(r.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrEmailAlreadyExist:
			h.respondWithError(w, http.StatusConflict, "Email already exists")
		case service.ErrInvalidInput:
			h.respondWithError(w, http.StatusBadRequest, "Invalid input")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		}
		return
	}

	logger.Info().Int("id", user.ID).Str("email", user.Email).Msg("User created via handler")
	h.respondWithData(w, http.StatusCreated, user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logger := h.logger.With().Str("request_id", requestID).Logger()

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Warn().Str("id", idStr).Msg("Invalid user ID for update")
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn().Err(err).Msg("Invalid request body for update user")
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	logger.Debug().Int("id", id).Msg("Updating user")
	user, err := h.service.UpdateUser(r.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			h.respondWithError(w, http.StatusNotFound, "User not found")
		case service.ErrEmailAlreadyExist:
			h.respondWithError(w, http.StatusConflict, "Email already exists")
		case service.ErrInvalidInput:
			h.respondWithError(w, http.StatusBadRequest, "Invalid input")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Failed to update user")
		}
		return
	}

	logger.Info().Int("id", id).Str("email", user.Email).Msg("User updated via handler")
	h.respondWithData(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	logger := h.logger.With().Str("request_id", requestID).Logger()

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		logger.Warn().Str("id", idStr).Msg("Invalid user ID for delete")
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	logger.Debug().Int("id", id).Msg("Deleting user")
	err = h.service.DeleteUser(r.Context(), id)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			h.respondWithError(w, http.StatusNotFound, "User not found")
		case service.ErrInvalidInput:
			h.respondWithError(w, http.StatusBadRequest, "Invalid input")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		}
		return
	}

	logger.Info().Int("id", id).Msg("User deleted via handler")
	h.respondWithSuccess(w, http.StatusNoContent, "User deleted successfully", nil)
}

func (h *UserHandler) respondWithSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := dto.Response{
		Success: true,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) respondWithData(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *UserHandler) respondWithError(w http.ResponseWriter, statusCode int, errorMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := dto.Response{
		Success: false,
		Message: errorMessage,
		Error:   errorMessage,
	}
	json.NewEncoder(w).Encode(response)
}
