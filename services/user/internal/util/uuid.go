package util

import (
	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID v4
func GenerateUUID() uuid.UUID {
	return uuid.New()
}

// ParseUUID parses a UUID string
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
