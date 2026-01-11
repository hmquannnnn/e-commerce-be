package model

import "time"

type City struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type District struct {
	ID        int       `json:"id"`
	CityID    int       `json:"city_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Ward struct {
	ID         int    `json:"id"`
	DistrictID int    `json:"district_id"`
	Name       string `json:"name"`
	Code       string `json:"code,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LocationInfo struct {
	City     City     `json:"city"`
	District District `json:"district"`
	Ward     Ward     `json:"ward"`
}
