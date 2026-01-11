package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/user-service/model"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrRefreshTokenRevoked  = errors.New("refresh token revoked")
)

// RefreshTokenRepository defines the interface for refresh token data access
type RefreshTokenRepository interface {
	// Create creates a new refresh token
	Create(ctx context.Context, token *model.RefreshToken) error

	// GetByToken retrieves a refresh token by token string
	GetByToken(ctx context.Context, token string) (*model.RefreshToken, error)

	// Revoke revokes a refresh token
	Revoke(ctx context.Context, token string) error

	// RevokeByUserID revokes all refresh tokens for a user
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error

	// DeleteExpired deletes expired refresh tokens
	DeleteExpired(ctx context.Context) error
}

type refreshTokenRepository struct {
	db *sql.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *sql.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (
			user_id, token, expires_at
		) VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	query := `
		SELECT 
			id, user_id, token,
			expires_at, revoked, revoked_at, created_at
		FROM refresh_tokens
		WHERE token = $1
	`

	rt := &model.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.Revoked,
		&rt.RevokedAt, &rt.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	// Check if token is valid
	if rt.Revoked {
		return nil, ErrRefreshTokenRevoked
	}

	if rt.ExpiresAt.Before(time.Now()) {
		return nil, ErrRefreshTokenExpired
	}

	return rt, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, token string) error {
	query := `
		UPDATE refresh_tokens SET
			revoked = TRUE,
			revoked_at = $1
		WHERE token = $2 AND revoked = FALSE
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), token)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrRefreshTokenNotFound
	}

	return nil
}

func (r *refreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens SET
			revoked = TRUE,
			revoked_at = $1
		WHERE user_id = $2 AND revoked = FALSE
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh tokens by user ID: %w", err)
	}

	return nil
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW() OR (revoked = TRUE AND revoked_at < NOW() - INTERVAL '30 days')
	`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return nil
}
