package repository

import (
	"context"

	"backend/middleware"
	"backend/models"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetAll(ctx context.Context, page, pageSize int) ([]models.User, int64, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, id int, user *models.User) error
	Delete(ctx context.Context, id int) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type userRepository struct {
	db     *sqlx.DB
	logger zerolog.Logger
}

func NewUserRepository(db *sqlx.DB, logger zerolog.Logger) UserRepository {
	return &userRepository{db: db, logger: logger}
}

func (r *userRepository) getLogger(ctx context.Context) *zerolog.Logger {
	requestID := middleware.GetRequestID(ctx)
	l := r.logger.With().Str("request_id", requestID).Logger()
	return &l
}

func (r *userRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	var user models.User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = ?", id)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("id", id).Msg("Failed to get user by ID")
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAll(ctx context.Context, page, pageSize int) ([]models.User, int64, error) {
	offset := (page - 1) * pageSize
	var users []models.User
	err := r.db.SelectContext(ctx, &users, "SELECT * FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?", pageSize, offset)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("page", page).Int("pageSize", pageSize).Msg("Failed to get users")
		return nil, 0, err
	}

	var total int64
	err = r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM users")
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Msg("Failed to count users")
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (name, email, age, created_at, updated_at) 
	          VALUES (:name, :email, :age, :created_at, :updated_at)`
	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Str("email", user.Email).Msg("Failed to create user")
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Msg("Failed to get last insert id")
		return err
	}
	user.ID = int(id)
	r.getLogger(ctx).Info().Int("id", user.ID).Str("email", user.Email).Msg("User created successfully")
	return nil
}

func (r *userRepository) Update(ctx context.Context, id int, user *models.User) error {
	query := `UPDATE users SET name = :name, email = :email, age = :age, updated_at = :updated_at WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("id", id).Msg("Failed to update user")
		return err
	}
	r.getLogger(ctx).Info().Int("id", id).Str("email", user.Email).Msg("User updated successfully")
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Int("id", id).Msg("Failed to delete user")
		return err
	}
	r.getLogger(ctx).Info().Int("id", id).Msg("User deleted successfully")
	return nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email)
	if err != nil {
		r.getLogger(ctx).Error().Err(err).Str("email", email).Msg("Failed to check email existence")
		return false, err
	}
	return exists, nil
}
