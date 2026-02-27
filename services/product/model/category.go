package model

import "time"

type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateCategoryParams struct {
	Name        string
	Description *string
}

type UpdateCategoryParams struct {
	Name        *string
	Description *string
}
