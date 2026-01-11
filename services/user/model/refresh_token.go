package model

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token in the system
type RefreshToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	Revoked   bool       `json:"revoked"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// IsValid checks if refresh token is valid (not expired and not revoked)
func (rt *RefreshToken) IsValid() bool {
	return !rt.Revoked && rt.ExpiresAt.After(time.Now())
}

// Revoke revokes the refresh token
func (rt *RefreshToken) Revoke() {
	now := time.Now()
	rt.Revoked = true
	rt.RevokedAt = &now
}

// CreateRefreshTokenParams represents parameters for creating a new refresh token
type CreateRefreshTokenParams struct {
	UserID     uuid.UUID
	DeviceName *string
	DeviceType *string
	UserAgent  *string
	IPAddress  *string
	Duration   time.Duration
}
