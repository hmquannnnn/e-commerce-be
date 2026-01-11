package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/user-service/internal/util"
	"github.com/hmquannnnn/e-commerce/user-service/model"
	"github.com/hmquannnnn/e-commerce/user-service/repository"
)

// AuthService handles authentication-related operations
type AuthService interface {
	Register(ctx context.Context, params RegisterParams) (*model.User, string, string, error)
	Login(ctx context.Context, email, password, ip, userAgent string) (*model.User, string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}

// RegisterParams represents parameters for user registration
type RegisterParams struct {
	Email     string
	Password  string
	Name      string
	IP        string
	UserAgent string
}

type authService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtManager       *util.JWTManager
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtManager *util.JWTManager,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
	}
}

// Register handles user registration
func (s *authService) Register(ctx context.Context, params RegisterParams) (*model.User, string, string, error) {
	if !util.IsPasswordValid(params.Password) {
		return nil, "", "", ErrInvalidPassword
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, params.Email)
	if err == nil && existingUser != nil {
		return nil, "", "", ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := util.HashPassword(params.Password)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		ID:           uuid.New(),
		Email:        params.Email,
		PasswordHash: hashedPassword,
		Name:         params.Name,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Role:         model.RoleUser,
		Gender:       model.GenderMale,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", "", fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

// Login handles user login
func (s *authService) Login(ctx context.Context, email, password, ip, userAgent string) (*model.User, string, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", "", ErrInvalidCredentials
		}
		return nil, "", "", fmt.Errorf("failed to get user: %w", err)
	}

	if err := util.VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

// RefreshToken generates a new access token using a refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, error) {
	// Get refresh token from database
	refreshToken, err := s.refreshTokenRepo.GetByToken(ctx, refreshTokenStr)
	if err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotFound) ||
			errors.Is(err, repository.ErrRefreshTokenExpired) ||
			errors.Is(err, repository.ErrRefreshTokenRevoked) {
			return "", ErrInvalidToken
		}
		return "", fmt.Errorf("failed to get refresh token: %w", err)
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, refreshToken.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Generate new access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, nil
}

// Logout revokes all refresh tokens for a user
func (s *authService) Logout(ctx context.Context, userID uuid.UUID) error {
	if err := s.refreshTokenRepo.RevokeByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}

	return nil
}

// ============================================================================
// Private Helper Methods
// ============================================================================

// createRefreshToken creates and stores a new refresh token
func (s *authService) createRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// Generate refresh token
	tokenString, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return "", err
	}

	// Create refresh token record
	refreshToken := &model.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     tokenString,
		ExpiresAt: time.Now().Add(s.jwtManager.GetRefreshTokenDuration()),
		Revoked:   false,
		CreatedAt: time.Now(),
	}

	// Save to database
	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return "", err
	}

	return tokenString, nil
}
