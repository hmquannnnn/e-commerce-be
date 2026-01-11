package handler

import "github.com/hmquannnnn/e-commerce/user-service/model"

type RegisterRequest struct {
	Email    string          `json:"email" binding:"required,email"`
	Password string          `json:"password" binding:"required,min=8"`
	Name     string          `json:"name" binding:"required,min=2,max=100"`
	Phone    *string         `json:"phone,omitempty"`
	Gender   *model.Gender   `json:"gender,omitempty"`
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
	ID          string         `json:"id"`
	Email       string         `json:"email"`
	Phone       *string        `json:"phone,omitempty"`
	Name        string         `json:"name"`
	AvatarURL   *string        `json:"avatar_url,omitempty"`
	DateOfBirth *string        `json:"date_of_birth,omitempty"`
	Gender      *model.Gender  `json:"gender,omitempty"`
	Role        model.UserRole `json:"role"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

type UpdateUserRequest struct {
	Name        *string       `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Phone       *string       `json:"phone,omitempty"`
	AvatarURL   *string       `json:"avatar_url,omitempty"`
	DateOfBirth *string       `json:"date_of_birth,omitempty" binding:"omitempty,datetime=2006-01-02"`
	Gender      *model.Gender `json:"gender,omitempty"`
}

func ToUserResponse(user *model.User, urlGenerator ...func(string) string) *UserResponse {
	var avatarURL *string

	// If user.AvatarURL contains filePath (relative path), convert to full URL
	if user.AvatarURL != nil && *user.AvatarURL != "" {
		// Check if it's already a full URL (contains http:// or https://)
		avatarPath := *user.AvatarURL
		if len(urlGenerator) > 0 && urlGenerator[0] != nil {
			// Generate full URL from filePath
			fullURL := urlGenerator[0](avatarPath)
			avatarURL = &fullURL
		} else {
			// If no URL generator provided, check if it's already a URL
			if len(avatarPath) > 7 && (avatarPath[:7] == "http://" || (len(avatarPath) > 8 && avatarPath[:8] == "https://")) {
				// Already a full URL, use as-is
				avatarURL = user.AvatarURL
			} else {
				// It's a filePath but no generator, use as-is (will need frontend to handle)
				avatarURL = user.AvatarURL
			}
		}
	}

	var dateOfBirth *string
	if user.DateOfBirth != nil {
		formatted := user.DateOfBirth.Format("2006-01-02")
		dateOfBirth = &formatted
	}

	return &UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		Phone:       user.Phone,
		Name:        user.Name,
		AvatarURL:   avatarURL,
		DateOfBirth: dateOfBirth,
		Gender:      &user.Gender,
		Role:        user.Role,
		CreatedAt:   user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
