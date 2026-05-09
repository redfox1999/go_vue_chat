package service

import (
	"context"

	"backend/dto"
	"backend/models"
	"backend/repository"

	"github.com/rs/zerolog"
)

type MessageService interface {
	GetByRoomID(ctx context.Context, roomID int, page, pageSize int) (*dto.MessageListResponse, error)
}

type messageService struct {
	repo   repository.MessageRepository
	logger zerolog.Logger
}

func NewMessageService(repo repository.MessageRepository, logger zerolog.Logger) MessageService {
	return &messageService{repo: repo, logger: logger}
}

func (s *messageService) GetByRoomID(ctx context.Context, roomID int, page, pageSize int) (*dto.MessageListResponse, error) {
	messages, total, err := s.repo.GetByRoomID(ctx, roomID, page, pageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.MessageResponse, 0, len(messages))
	for _, msg := range messages {
		responses = append(responses, *s.toResponse(&msg))
	}

	return &dto.MessageListResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}, nil
}

func (s *messageService) toResponse(msg *models.Message) *dto.MessageResponse {
	return &dto.MessageResponse{
		ID:       msg.ID,
		RoomID:   msg.RoomID,
		Sender:   msg.Sender,
		Nickname: msg.Nickname,
		Notify:   msg.Notify,
		Message:  msg.Message,
		SendTime: msg.SendTime,
	}
}
