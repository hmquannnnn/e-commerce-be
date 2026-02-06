package handler

import "github.com/hmquannnnn/e-commerce/user-service/model"

type RegisterRequest struct {
	Email    string          `json:"email" binding:"required,email"`
	Password string          `json:"password" binding:"required,min=8"`
	Name     string          `json:"name" binding:"required,min=2,max=100"`
	Phone    *string         `json:"phone,omitempty"`
	Role     *model.UserRole `json:"role,omitempty"` // Optional, defaults to "user"
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	User         *UserResponse `json:"user"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type UserResponse struct {
	ID        string         `json:"id"`
	Email     string         `json:"email"`
	Phone     *string        `json:"phone,omitempty"`
	Name      string         `json:"name"`
	Role      model.UserRole `json:"role"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Phone *string `json:"phone,omitempty"`
}

func ToUserResponse(user *model.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Phone:     user.Phone,
		Name:      user.Name,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
