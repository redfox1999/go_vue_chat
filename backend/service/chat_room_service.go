package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"backend/dto"
	"backend/middleware"
	"backend/models"
	"backend/repository"

	"github.com/rs/zerolog"
)

var (
	ErrChatRoomNotFound   = errors.New("chat room not found")
	ErrChatRoomNameExists = errors.New("chat room name already exists")
)

type ChatRoomService interface {
	GetByID(ctx context.Context, id int) (*dto.ChatRoomResponse, error)
	GetAll(ctx context.Context, page, pageSize int) (*dto.ChatRoomListResponse, error)
	GetByGroup(ctx context.Context, group string, page, pageSize int) (*dto.ChatRoomListResponse, error)
	GetByOwner(ctx context.Context, ownerID int, page, pageSize int) (*dto.ChatRoomListResponse, error)
	Create(ctx context.Context, req *dto.CreateChatRoomRequest) (*dto.ChatRoomResponse, error)
	Update(ctx context.Context, id int, req *dto.UpdateChatRoomRequest) (*dto.ChatRoomResponse, error)
	Delete(ctx context.Context, id int) error
}

type chatRoomService struct {
	repo   repository.ChatRoomRepository
	logger zerolog.Logger
}

func NewChatRoomService(repo repository.ChatRoomRepository, logger zerolog.Logger) ChatRoomService {
	return &chatRoomService{repo: repo, logger: logger}
}

func (s *chatRoomService) getLogger(ctx context.Context) zerolog.Logger {
	requestID := middleware.GetRequestID(ctx)
	return s.logger.With().Str("request_id", requestID).Logger()
}

func (s *chatRoomService) GetByID(ctx context.Context, id int) (*dto.ChatRoomResponse, error) {
	room, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrChatRoomNotFound
		}
		return nil, err
	}
	return s.toResponse(room, 0), nil
}

func (s *chatRoomService) GetAll(ctx context.Context, page, pageSize int) (*dto.ChatRoomListResponse, error) {
	rooms, total, err := s.repo.GetAll(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ChatRoomResponse, 0, len(rooms))
	for _, room := range rooms {
		responses = append(responses, *s.toResponse(&room, 0))
	}

	return &dto.ChatRoomListResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}, nil
}

func (s *chatRoomService) GetByGroup(ctx context.Context, group string, page, pageSize int) (*dto.ChatRoomListResponse, error) {
	rooms, total, err := s.repo.GetByGroup(ctx, group, page, pageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ChatRoomResponse, 0, len(rooms))
	for _, room := range rooms {
		responses = append(responses, *s.toResponse(&room, 0))
	}

	return &dto.ChatRoomListResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}, nil
}

func (s *chatRoomService) GetByOwner(ctx context.Context, ownerID int, page, pageSize int) (*dto.ChatRoomListResponse, error) {
	rooms, total, err := s.repo.GetByOwner(ctx, ownerID, page, pageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ChatRoomResponse, 0, len(rooms))
	for _, room := range rooms {
		responses = append(responses, *s.toResponse(&room, 0))
	}

	return &dto.ChatRoomListResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}, nil
}

func (s *chatRoomService) Create(ctx context.Context, req *dto.CreateChatRoomRequest) (*dto.ChatRoomResponse, error) {
	exists, err := s.repo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrChatRoomNameExists
	}

	now := time.Now()
	room := &models.ChatRoom{
		Name:     req.Name,
		Logo:     req.Logo,
		Desc:     req.Desc,
		Group:    req.Group,
		OwnerID:  req.OwnerID,
		Status:   1,
		CreateAt: now,
		UpdateAt: now,
	}

	err = s.repo.Create(ctx, room)
	if err != nil {
		return nil, err
	}

	return s.toResponse(room, 0), nil
}

func (s *chatRoomService) Update(ctx context.Context, id int, req *dto.UpdateChatRoomRequest) (*dto.ChatRoomResponse, error) {
	room, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrChatRoomNotFound
		}
		return nil, err
	}

	if req.Name != nil && *req.Name != room.Name {
		exists, err := s.repo.ExistsByName(ctx, *req.Name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrChatRoomNameExists
		}
		room.Name = *req.Name
	}

	if req.Logo != nil {
		room.Logo = *req.Logo
	}
	if req.Desc != nil {
		room.Desc = *req.Desc
	}
	if req.Group != nil {
		room.Group = *req.Group
	}
	if req.Status != nil {
		room.Status = *req.Status
	}
	if req.OwnerID != nil {
		room.OwnerID = *req.OwnerID
	}
	room.UpdateAt = time.Now()

	err = s.repo.Update(ctx, id, room)
	if err != nil {
		return nil, err
	}

	return s.toResponse(room, 0), nil
}

func (s *chatRoomService) Delete(ctx context.Context, id int) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrChatRoomNotFound
		}
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *chatRoomService) toResponse(room *models.ChatRoom, clientNum int) *dto.ChatRoomResponse {
	return &dto.ChatRoomResponse{
		ID:        room.ID,
		Name:      room.Name,
		Logo:      room.Logo,
		Desc:      room.Desc,
		OwnerID:   room.OwnerID,
		Group:     room.Group,
		Status:    room.Status,
		CreateAt:  room.CreateAt,
		UpdateAt:  room.UpdateAt,
		ClientNum: clientNum,
	}
}
