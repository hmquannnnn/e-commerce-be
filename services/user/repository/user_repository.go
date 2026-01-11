package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/pkg/common/utils"
	"github.com/hmquannnnn/e-commerce/user-service/model"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error

	Update(ctx context.Context, id uuid.UUID, params *model.UpdateUserParams) error

	UpdateAvatar(ctx context.Context, id uuid.UUID, avatarUrl string) error

	GetByEmail(ctx context.Context, email string) (*model.User, error)

	GetByName(ctx context.Context, name string) (*model.User, error)

	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (
			id, email, phone, password_hash, name, avatar_url,
			date_of_birth, gender, role,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		user.ID, user.Email, user.Phone, user.PasswordHash, user.Name,
		user.AvatarURL, user.DateOfBirth, user.Gender, user.Role,
		user.CreatedAt, user.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation
		if isDuplicateKeyError(err) {
			if isDuplicateEmail(err) {
				return ErrEmailAlreadyExists
			}
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Update updates only the provided fields of a user (PATCH operation)
func (r *userRepository) Update(ctx context.Context, id uuid.UUID, params *model.UpdateUserParams) error {
	// Build updates map with only non-nil fields
	updates := make(map[string]interface{})

	if params.Name != nil {
		updates["name"] = *params.Name
	}

	if params.Phone != nil {
		updates["phone"] = *params.Phone
	}

	if params.AvatarURL != nil {
		updates["avatar_url"] = *params.AvatarURL
	}

	if params.DateOfBirth != nil {
		updates["date_of_birth"] = *params.DateOfBirth
	}

	if params.Gender != nil {
		updates["gender"] = *params.Gender
	}

	// Execute patch update using common utils
	rowsAffected, err := utils.ExecPatch(ctx, r.db, utils.PatchUpdate{
		Table:     "users",
		Updates:   updates,
		Where:     "id = $1",
		WhereArgs: []interface{}{id},
	})

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// Helper functions
func isDuplicateKeyError(err error) bool {
	// PostgreSQL duplicate key error code: 23505
	return err != nil && (err.Error() == "pq: duplicate key value violates unique constraint" ||
		err.Error() == "ERROR: duplicate key value violates unique constraint")
}

func isDuplicateEmail(err error) bool {
	return err != nil && (err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" ||
		err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\"")
}

func (r *userRepository) UpdateAvatar(ctx context.Context, id uuid.UUID, avatarUrl string) error {
	query := `
		UPDATE users SET
			avatar_url = $1,
			updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, avatarUrl, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update avatar: %w", err)
	}

	return nil
}

func (r *userRepository) getOneByField(
	ctx context.Context,
	field string,
	value any,
) (*model.User, error) {

	allowed := map[string]bool{
		"id":    true,
		"email": true,
		"name":  true,
	}
	if !allowed[field] {
		return nil, fmt.Errorf("invalid field: %s", field)
	}

	query := fmt.Sprintf(`
        SELECT id, email, phone, name, avatar_url, date_of_birth, gender, role, password_hash, 
        created_at, updated_at
        FROM users
        WHERE %s = $1
    `, field)

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, value).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.Name,
		&user.AvatarURL,
		&user.DateOfBirth,
		&user.Gender,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return r.getOneByField(ctx, "id", id.String())
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.getOneByField(ctx, "email", email)
}

func (r *userRepository) GetByName(ctx context.Context, name string) (*model.User, error) {
	return r.getOneByField(ctx, "name", name)
}
