package service

import (
	"context"
	"errors"
	"time"

	"backend/models"
	"backend/repository"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExist = errors.New("email already exists")
	ErrInvalidInput      = errors.New("invalid input")
)

type UserService interface {
	GetUserByID(ctx context.Context, id int) (*models.UserResponse, error)
	GetAllUsers(ctx context.Context, page, pageSize int) (*models.PaginationResponse, error)
	CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.UserResponse, error)
	UpdateUser(ctx context.Context, id int, req *models.UpdateUserRequest) (*models.UserResponse, error)
	DeleteUser(ctx context.Context, id int) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetUserByID(ctx context.Context, id int) (*models.UserResponse, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *userService) GetAllUsers(ctx context.Context, page, pageSize int) (*models.PaginationResponse, error) {
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

	userResponses := make([]models.UserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &models.PaginationResponse{
		Data:       userResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *userService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.UserResponse, error) {
	if req.Name == "" || req.Email == "" {
		return nil, ErrInvalidInput
	}

	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
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

	response := createdUser.ToResponse()
	return &response, nil
}

func (s *userService) UpdateUser(ctx context.Context, id int, req *models.UpdateUserRequest) (*models.UserResponse, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
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

	response := updatedUser.ToResponse()
	return &response, nil
}

func (s *userService) DeleteUser(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidInput
	}

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	return s.repo.Delete(ctx, id)
}
