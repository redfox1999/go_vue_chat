package repository

import (
	"context"

	"backend/middleware"
	"backend/models"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type ChatRoomRepository interface {
	GetByID(ctx context.Context, id int) (*models.ChatRoom, error)
	GetAll(ctx context.Context, page, pageSize int) ([]models.ChatRoom, int64, error)
	GetByGroup(ctx context.Context, group string, page, pageSize int) ([]models.ChatRoom, int64, error)
	GetByOwner(ctx context.Context, ownerID int, page, pageSize int) ([]models.ChatRoom, int64, error)
	Create(ctx context.Context, room *models.ChatRoom) error
	Update(ctx context.Context, id int, room *models.ChatRoom) error
	Delete(ctx context.Context, id int) error
	ExistsByName(ctx context.Context, name string) (bool, error)
}

type chatRoomRepository struct {
	db     *sqlx.DB
	logger zerolog.Logger
}

func NewChatRoomRepository(db *sqlx.DB, logger zerolog.Logger) ChatRoomRepository {
	return &chatRoomRepository{db: db, logger: logger}
}

func (r *chatRoomRepository) getLogger(ctx context.Context) *zerolog.Logger {
	requestID := middleware.GetRequestID(ctx)
	l := r.logger.With().Str("request_id", requestID).Logger()
	return &l
}

func (r *chatRoomRepository) GetByID(ctx context.Context, id int) (*models.ChatRoom, error) {
	var room models.ChatRoom
	err := r.db.GetContext(ctx, &room, "SELECT * FROM chat_rooms WHERE id = ?", id)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("id", id).Msg("Failed to get chat room by ID")
		return nil, err
	}
	return &room, nil
}

func (r *chatRoomRepository) GetAll(ctx context.Context, page, pageSize int) ([]models.ChatRoom, int64, error) {
	offset := (page - 1) * pageSize
	var rooms []models.ChatRoom
	err := r.db.SelectContext(ctx, &rooms, "SELECT * FROM chat_rooms WHERE status = 1 ORDER BY create_at DESC LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("page", page).Int("pageSize", pageSize).Msg("Failed to get chat rooms")
		return nil, 0, err
	}

	var total int64
	err = r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM chat_rooms WHERE status = 1")
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Msg("Failed to count chat rooms")
		return nil, 0, err
	}

	return rooms, total, nil
}

func (r *chatRoomRepository) GetByGroup(ctx context.Context, group string, page, pageSize int) ([]models.ChatRoom, int64, error) {
	offset := (page - 1) * pageSize
	var rooms []models.ChatRoom
	err := r.db.SelectContext(ctx, &rooms, "SELECT * FROM chat_rooms WHERE \"group\" = ? AND status = 1 ORDER BY create_at DESC LIMIT ? OFFSET ?", group, pageSize, offset)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Str("group", group).Msg("Failed to get chat rooms by group")
		return nil, 0, err
	}

	var total int64
	err = r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM chat_rooms WHERE \"group\" = ? AND status = 1", group)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Str("group", group).Msg("Failed to count chat rooms by group")
		return nil, 0, err
	}

	return rooms, total, nil
}

func (r *chatRoomRepository) GetByOwner(ctx context.Context, ownerID int, page, pageSize int) ([]models.ChatRoom, int64, error) {
	offset := (page - 1) * pageSize
	var rooms []models.ChatRoom
	err := r.db.SelectContext(ctx, &rooms, "SELECT * FROM chat_rooms WHERE owner_id = ? AND status = 1 ORDER BY create_at DESC LIMIT ? OFFSET ?", ownerID, pageSize, offset)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("owner_id", ownerID).Msg("Failed to get chat rooms by owner")
		return nil, 0, err
	}

	var total int64
	err = r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM chat_rooms WHERE owner_id = ? AND status = 1", ownerID)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("owner_id", ownerID).Msg("Failed to count chat rooms by owner")
		return nil, 0, err
	}

	return rooms, total, nil
}

func (r *chatRoomRepository) Create(ctx context.Context, room *models.ChatRoom) error {
	query := `INSERT INTO chat_rooms (name, logo, "desc", owner_id, "group", status, create_at, update_at)
	          VALUES (:name, :logo, :desc, :owner_id, :group, :status, :create_at, :update_at)`
	result, err := r.db.NamedExecContext(ctx, query, room)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Str("name", room.Name).Msg("Failed to create chat room")
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Msg("Failed to get last insert id")
		return err
	}
	room.ID = int(id)
	r.getLogger(ctx).Info().Int("id", room.ID).Str("name", room.Name).Msg("Chat room created successfully")
	return nil
}

func (r *chatRoomRepository) Update(ctx context.Context, id int, room *models.ChatRoom) error {
	query := `UPDATE chat_rooms SET name = :name, logo = :logo, "desc" = :desc, 
	          owner_id = :owner_id, "group" = :group, status = :status, update_at = :update_at WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, room)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("id", id).Msg("Failed to update chat room")
		return err
	}
	r.getLogger(ctx).Info().Int("id", id).Str("name", room.Name).Msg("Chat room updated successfully")
	return nil
}

func (r *chatRoomRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM chat_rooms WHERE id = ?", id)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("id", id).Msg("Failed to delete chat room")
		return err
	}
	r.getLogger(ctx).Info().Int("id", id).Msg("Chat room deleted successfully")
	return nil
}

func (r *chatRoomRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM chat_rooms WHERE name = ?)", name)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Str("name", name).Msg("Failed to check chat room name existence")
		return false, err
	}
	return exists, nil
}
