package model

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleUser   UserRole = "user"
	RoleSeller UserRole = "seller"
	RoleAdmin  UserRole = "admin"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	Phone        *string    `json:"phone,omitempty"`
	PasswordHash string     `json:"-"`
	Name         string     `json:"name"`
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	DateOfBirth  *time.Time `json:"date_of_birth,omitempty"`
	Gender       Gender     `json:"gender"`
	Role         UserRole   `json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type CreateUserParams struct {
	Email    string
	Phone    *string
	Password string
	Name     string
	Role     UserRole
	Gender   Gender
}

type UpdateUserParams struct {
	Name        *string
	Phone       *string
	AvatarURL   *string
	DateOfBirth *time.Time
	Gender      *Gender
}
