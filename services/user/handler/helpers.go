package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/user-service/internal/util"
	"github.com/hmquannnnn/e-commerce/user-service/model"
)

// Context keys for storing user info in request context
type contextKey string

const (
	userIDKey    contextKey = "user_id"
	userEmailKey contextKey = "user_email"
	userRoleKey  contextKey = "user_role"
)

// ============================================================================
// JWT Token Helpers
// ============================================================================

// ExtractBearerToken extracts the bearer token from the Authorization header
func ExtractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("authorization header format must be Bearer {token}")
	}

	return parts[1], nil
}

// ValidateAndExtractClaims validates JWT token and extracts claims
func ValidateAndExtractClaims(tokenString string, jwtManager *util.JWTManager) (*util.JWTClaims, error) {
	claims, err := jwtManager.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// ============================================================================
// Context Helpers
// ============================================================================

// SetUserContext sets user information in request context
func SetUserContext(ctx context.Context, userID uuid.UUID, email, role string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID.String())
	ctx = context.WithValue(ctx, userEmailKey, email)
	ctx = context.WithValue(ctx, userRoleKey, role)
	return ctx
}

// GetUserIDFromContext extracts user ID from Gin context (set by RequireAuth middleware)
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, errors.New("user ID not found in context")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID type in context")
	}

	return userID, nil
}

// GetUserEmailFromContext extracts user email from Gin context
func GetUserEmailFromContext(c *gin.Context) (string, error) {
	emailValue, exists := c.Get("userEmail")
	if !exists {
		return "", errors.New("user email not found in context")
	}

	email, ok := emailValue.(string)
	if !ok {
		return "", errors.New("invalid user email type in context")
	}

	return email, nil
}

// GetUserRoleFromContext extracts user role from Gin context
func GetUserRoleFromContext(c *gin.Context) (model.UserRole, error) {
	roleValue, exists := c.Get("userRole")
	if !exists {
		return "", errors.New("user role not found in context")
	}

	role, ok := roleValue.(model.UserRole)
	if !ok {
		return "", errors.New("invalid user role type in context")
	}

	return role, nil
}

// ============================================================================
// Validation Helpers
// ============================================================================

// ValidateEmail performs basic email validation
func ValidateEmail(email string) bool {
	// Basic validation
	return strings.Contains(email, "@") && len(email) > 3
}

// ValidatePassword performs password validation
func ValidatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "password must be at least 8 characters long"
	}

	if !util.IsPasswordValid(password) {
		return false, "password must contain at least 3 of the following: uppercase, lowercase, number, special character"
	}

	return true, ""
}
