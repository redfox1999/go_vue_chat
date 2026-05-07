package repository

import (
	"context"

	"backend/middleware"
	"backend/models"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type MessageRepository interface {
	GetByID(ctx context.Context, id int) (*models.Message, error)
	GetByRoomID(ctx context.Context, roomID int, page, pageSize int) ([]models.Message, int64, error)
	GetBySender(ctx context.Context, sender int, page, pageSize int) ([]models.Message, int64, error)
	Create(ctx context.Context, message *models.Message) error
	Delete(ctx context.Context, id int) error
	DeleteByRoomID(ctx context.Context, roomID int) error
}

type messageRepository struct {
	db     *sqlx.DB
	logger zerolog.Logger
}

func NewMessageRepository(db *sqlx.DB, logger zerolog.Logger) MessageRepository {
	return &messageRepository{db: db, logger: logger}
}

func (r *messageRepository) getLogger(ctx context.Context) *zerolog.Logger {
	requestID := middleware.GetRequestID(ctx)
	l := r.logger.With().Str("request_id", requestID).Logger()
	return &l
}

func (r *messageRepository) GetByID(ctx context.Context, id int) (*models.Message, error) {
	var message models.Message
	err := r.db.GetContext(ctx, &message, "SELECT * FROM messages WHERE id = ?", id)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("id", id).Msg("Failed to get message by ID")
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) GetByRoomID(ctx context.Context, roomID int, page, pageSize int) ([]models.Message, int64, error) {
	offset := (page - 1) * pageSize
	var messages []models.Message
	err := r.db.SelectContext(ctx, &messages, "SELECT * FROM messages WHERE room_id = ? ORDER BY send_time DESC LIMIT ? OFFSET ?", roomID, pageSize, offset)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("room_id", roomID).Msg("Failed to get messages by room ID")
		return nil, 0, err
	}

	var total int64
	err = r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM messages WHERE room_id = ?", roomID)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("room_id", roomID).Msg("Failed to count messages by room ID")
		return nil, 0, err
	}

	return messages, total, nil
}

func (r *messageRepository) GetBySender(ctx context.Context, sender int, page, pageSize int) ([]models.Message, int64, error) {
	offset := (page - 1) * pageSize
	var messages []models.Message
	err := r.db.SelectContext(ctx, &messages, "SELECT * FROM messages WHERE sender = ? ORDER BY send_time DESC LIMIT ? OFFSET ?", sender, pageSize, offset)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("sender", sender).Msg("Failed to get messages by sender")
		return nil, 0, err
	}

	var total int64
	err = r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM messages WHERE sender = ?", sender)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("sender", sender).Msg("Failed to count messages by sender")
		return nil, 0, err
	}

	return messages, total, nil
}

func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	query := `INSERT INTO messages (room_id, sender, notify, message, send_time)
	          VALUES (:room_id, :sender, :notify, :message, :send_time)`
	result, err := r.db.NamedExecContext(ctx, query, message)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("room_id", message.RoomID).Msg("Failed to create message")
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Msg("Failed to get last insert id")
		return err
	}
	message.ID = int(id)
	r.getLogger(ctx).Info().Int("id", message.ID).Int("room_id", message.RoomID).Msg("Message created successfully")
	return nil
}

func (r *messageRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM messages WHERE id = ?", id)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("id", id).Msg("Failed to delete message")
		return err
	}
	r.getLogger(ctx).Info().Int("id", id).Msg("Message deleted successfully")
	return nil
}

func (r *messageRepository) DeleteByRoomID(ctx context.Context, roomID int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM messages WHERE room_id = ?", roomID)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("room_id", roomID).Msg("Failed to delete messages by room ID")
		return err
	}
	r.getLogger(ctx).Info().Int("room_id", roomID).Msg("Messages deleted successfully")
	return nil
}
