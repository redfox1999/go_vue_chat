package service

import (
	"context"
	"errors"
	"time"

	"backend/dto"
	"backend/middleware"
	"backend/models"
	"backend/repository"

	"github.com/jinzhu/copier"
	"github.com/rs/zerolog"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExist = errors.New("email already exists")
	ErrInvalidInput      = errors.New("invalid input")
)

type UserService interface {
	GetUserByID(ctx context.Context, id int) (*dto.UserResponse, error)
	GetAllUsers(ctx context.Context, page, pageSize int) (*dto.PaginationResponse, error)
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	UpdateUser(ctx context.Context, id int, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, id int) error
}

type userService struct {
	repo   repository.UserRepository
	logger zerolog.Logger
}

func NewUserService(repo repository.UserRepository, logger zerolog.Logger) UserService {
	return &userService{repo: repo, logger: logger}
}

func (s *userService) getLogger(ctx context.Context) zerolog.Logger {
	requestID := middleware.GetRequestID(ctx)
	return s.logger.With().Str("request_id", requestID).Logger()
}

func (s *userService) toResponse(user *models.User) (*dto.UserResponse, error) {
	var response dto.UserResponse
	err := copier.Copy(&response, user)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *userService) GetUserByID(ctx context.Context, id int) (*dto.UserResponse, error) {
	logger := s.getLogger(ctx)

	if id <= 0 {
		logger.Warn().Int("id", id).Msg("Invalid user ID")
		return nil, ErrInvalidInput
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Warn().Int("id", id).Msg("User not found")
		return nil, ErrUserNotFound
	}

	return s.toResponse(user)
}

func (s *userService) GetAllUsers(ctx context.Context, page, pageSize int) (*dto.PaginationResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	users, total, err := s.repo.GetAll(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	userResponses := make([]dto.UserResponse, 0, len(users))
	for i := range users {
		response, err := s.toResponse(&users[i])
		if err != nil {
			return nil, err
		}
		userResponses = append(userResponses, *response)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &dto.PaginationResponse{
		Data:       userResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *userService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	logger := s.getLogger(ctx)

	if req.Name == "" || req.Email == "" {
		logger.Warn().Str("email", req.Email).Msg("Invalid input for creating user")
		return nil, ErrInvalidInput
	}

	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		logger.Warn().Str("email", req.Email).Msg("Email already exists")
		return nil, ErrEmailAlreadyExist
	}

	now := time.Now()
	user := &models.User{
		Name:      req.Name,
		Email:     req.Email,
		Age:       req.Age,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	createdUser, err := s.repo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	response, err := s.toResponse(createdUser)
	if err != nil {
		return nil, err
	}
	logger.Info().Int("id", createdUser.ID).Str("email", createdUser.Email).Msg("User created via service")
	return response, nil
}

func (s *userService) UpdateUser(ctx context.Context, id int, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	logger := s.getLogger(ctx)

	if id <= 0 {
		logger.Warn().Int("id", id).Msg("Invalid user ID for update")
		return nil, ErrInvalidInput
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Warn().Int("id", id).Msg("User not found for update")
		return nil, ErrUserNotFound
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		exists, err := s.repo.ExistsByEmail(ctx, *req.Email)
		if err != nil {
			return nil, err
		}
		if exists && *req.Email != user.Email {
			logger.Warn().Str("email", *req.Email).Msg("Email already exists for update")
			return nil, ErrEmailAlreadyExist
		}
		user.Email = *req.Email
	}
	if req.Age != nil {
		user.Age = *req.Age
	}

	user.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, id, user)
	if err != nil {
		return nil, err
	}

	updatedUser, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response, err := s.toResponse(updatedUser)
	if err != nil {
		return nil, err
	}
	logger.Info().Int("id", id).Str("email", updatedUser.Email).Msg("User updated via service")
	return response, nil
}

func (s *userService) DeleteUser(ctx context.Context, id int) error {
	logger := s.getLogger(ctx)

	if id <= 0 {
		logger.Warn().Int("id", id).Msg("Invalid user ID for delete")
		return ErrInvalidInput
	}

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Warn().Int("id", id).Msg("User not found for delete")
		return ErrUserNotFound
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	logger.Info().Int("id", id).Msg("User deleted via service")
	return nil
}
