package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/models"
	"backend/service"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{service: svc}
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		if err == service.ErrUserNotFound {
			h.respondWithError(w, http.StatusNotFound, "User not found")
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to get user")
		}
		return
	}

	h.respondWithSuccess(w, http.StatusOK, "Success", user)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.service.GetAllUsers(r.Context(), page, pageSize)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	h.respondWithSuccess(w, http.StatusOK, "Success", result)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.service.CreateUser(r.Context(), &req)
	if err != nil {
		if err == service.ErrEmailAlreadyExist {
			h.respondWithError(w, http.StatusConflict, "Email already exists")
		} else if err == service.ErrInvalidInput {
			h.respondWithError(w, http.StatusBadRequest, "Invalid input")
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		}
		return
	}

	h.respondWithSuccess(w, http.StatusCreated, "User created successfully", user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.service.UpdateUser(r.Context(), id, &req)
	if err != nil {
		if err == service.ErrUserNotFound {
			h.respondWithError(w, http.StatusNotFound, "User not found")
		} else if err == service.ErrEmailAlreadyExist {
			h.respondWithError(w, http.StatusConflict, "Email already exists")
		} else if err == service.ErrInvalidInput {
			h.respondWithError(w, http.StatusBadRequest, "Invalid input")
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to update user")
		}
		return
	}

	h.respondWithSuccess(w, http.StatusOK, "User updated successfully", user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = h.service.DeleteUser(r.Context(), id)
	if err != nil {
		if err == service.ErrUserNotFound {
			h.respondWithError(w, http.StatusNotFound, "User not found")
		} else if err == service.ErrInvalidInput {
			h.respondWithError(w, http.StatusBadRequest, "Invalid input")
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		}
		return
	}

	h.respondWithSuccess(w, http.StatusNoContent, "User deleted successfully", nil)
}

func (h *UserHandler) respondWithSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := models.Response{
		Success: true,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) respondWithError(w http.ResponseWriter, statusCode int, errorMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := models.Response{
		Success: false,
		Message: errorMessage,
		Error:   errorMessage,
	}
	json.NewEncoder(w).Encode(response)
}
