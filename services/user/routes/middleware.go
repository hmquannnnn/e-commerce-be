package routes

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/user-service/internal/util"
	"github.com/hmquannnnn/e-commerce/user-service/model"
)

// AuthMiddleware is a middleware that validates JWT tokens
type AuthMiddleware struct {
	jwtManager *util.JWTManager
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtManager *util.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth is a Gin middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "MISSING_TOKEN",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "INVALID_TOKEN_FORMAT",
				"message": "Authorization header must be Bearer token",
			})
			c.Abort()
			return
		}

		// Verify token
		claims, err := m.jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			if err == util.ErrExpiredToken {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "TOKEN_EXPIRED",
					"message": "Access token has expired",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "INVALID_TOKEN",
					"message": "Invalid access token",
				})
			}
			c.Abort()
			return
		}

		// Parse user ID
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "INVALID_TOKEN",
				"message": "Invalid user ID in token",
			})
			c.Abort()
			return
		}

		// Store user info in Gin context
		c.Set("userID", userID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", model.UserRole(claims.Role))

		// Continue to next handler
		c.Next()
	}
}

// RequireRole is a Gin middleware that requires a specific role
// Note: This should be used AFTER RequireAuth middleware
func (m *AuthMiddleware) RequireRole(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from Gin context (set by RequireAuth middleware)
		roleValue, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "UNAUTHORIZED",
				"message": "User role not found in context",
			})
			c.Abort()
			return
		}

		userRole, ok := roleValue.(model.UserRole)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "UNAUTHORIZED",
				"message": "Invalid user role type in context",
			})
			c.Abort()
			return
		}

		// Check if user has required role
		hasRole := false
		for _, role := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "FORBIDDEN",
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		// User has required role, continue
		c.Next()
	}
}

// RequireAdmin is a Gin middleware that requires admin role
// Note: This should be used AFTER RequireAuth middleware
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole(model.RoleAdmin)
}

// OptionalAuth is a Gin middleware that optionally authenticates the user
// If token is present and valid, it adds user info to context
// If token is missing or invalid, it continues without error
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token, continue without auth
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid token format, continue without auth
			c.Next()
			return
		}

		claims, err := m.jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			// Invalid token, continue without auth
			c.Next()
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			// Invalid user ID, continue without auth
			c.Next()
			return
		}

		// Add user info to Gin context
		c.Set("userID", userID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", model.UserRole(claims.Role))

		// Call next handler with updated context
		c.Next()
	}
}

// ============================================================================
// Other Common Middlewares
// ============================================================================

// CORSMiddleware handles CORS for Gin
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // In production, set specific origins
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggingMiddleware logs HTTP requests for Gin
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next() // Process request

		duration := time.Since(start)
		log.Printf("[%s] %s %s %d %s",
			time.Now().Format(time.RFC3339),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}

// RecoveryMiddleware recovers from panics for Gin
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "INTERNAL_ERROR",
					"message": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request for Gin
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get existing request ID from header
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate new request ID if not provided
			requestID = uuid.New().String()
		}

		// Add request ID to response header
		c.Writer.Header().Set("X-Request-ID", requestID)

		// Store request ID in Gin context
		c.Set("requestID", requestID)

		c.Next()
	}
}
