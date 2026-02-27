package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID type in context")
	}
	return id, nil
}

func GetUserRoleFromContext(c *gin.Context) (string, error) {
	val, exists := c.Get("userRole")
	if !exists {
		return "", errors.New("user role not found in context")
	}
	role, ok := val.(string)
	if !ok {
		return "", errors.New("invalid user role type in context")
	}
	return role, nil
}

func respondOK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func respondCreated(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"error":   code,
		"message": message,
	})
}
