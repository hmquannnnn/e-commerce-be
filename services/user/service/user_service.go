package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/pkg/storage"
	"github.com/hmquannnnn/e-commerce/user-service/model"
	"github.com/hmquannnnn/e-commerce/user-service/repository"
)

// UserService handles user management operations
type UserService interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, params UpdateUserParams) (*model.User, error)
	GetPublicURL(filePath string) string
}

// UpdateUserParams represents parameters for updating a user
type UpdateUserParams struct {
	Name        *string
	Phone       *string
	AvatarURL   *string
	DateOfBirth *time.Time
	Gender      *model.Gender
}

type userService struct {
	userRepo       repository.UserRepository
	storageManager *storage.Manager
	bucketName     string
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repository.UserRepository,
	storageManager *storage.Manager,
	bucketName string,
) UserService {
	return &userService{
		userRepo:       userRepo,
		storageManager: storageManager,
		bucketName:     bucketName,
	}
}

// GetUserByID retrieves a user by ID
func (s *userService) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// UpdateUser updates user information (PATCH operation - only updates provided fields)
func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, params UpdateUserParams) (*model.User, error) {
	// Convert service params to model params
	modelParams := &model.UpdateUserParams{
		Name:        params.Name,
		Phone:       params.Phone,
		AvatarURL:   params.AvatarURL,
		DateOfBirth: params.DateOfBirth,
		Gender:      params.Gender,
	}

	// Use Update method with dynamic field updates
	if err := s.userRepo.Update(ctx, id, modelParams); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Get updated user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}

	// Invalidate cache (optional)
	cacheKey := fmt.Sprintf("user:%s", id.String())
	_ = cacheKey // TODO: Implement cache invalidation

	return user, nil
}

// GetPublicURL generates full URL from filePath
func (s *userService) GetPublicURL(filePath string) string {
	return s.storageManager.GetPublicURL(filePath)
}
