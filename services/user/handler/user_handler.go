package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hmquannnnn/e-commerce/user-service/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "UNAUTHORIZED",
			"message": "User not authenticated",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	// Parse date_of_birth from string to time.Time
	var dateOfBirth *time.Time
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "INVALID_REQUEST",
				"message": "Invalid date_of_birth format. Expected format: YYYY-MM-DD",
			})
			return
		}
		dateOfBirth = &parsed
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID, service.UpdateUserParams{
		Name:        req.Name,
		Phone:       req.Phone,
		AvatarURL:   req.AvatarURL,
		DateOfBirth: dateOfBirth,
		Gender:      req.Gender,
	})

	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "USER_NOT_FOUND",
				"message": "User not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "INTERNAL_ERROR",
				"message": "Failed to update profile",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
		"data":    ToUserResponse(user, h.userService.GetPublicURL),
	})
}
