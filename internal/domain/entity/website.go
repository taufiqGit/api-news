package entity

import (
	"time"

	"github.com/google/uuid"
)

type Website struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Domain      *string    `json:"domain"`
	Description *string    `json:"description"`
	LogoURL     *string    `json:"logo_url"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type WebsiteCreate struct {
	Name        string  `json:"name" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
	Domain      *string `json:"domain"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
}

type WebsiteUpdate struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Domain      *string `json:"domain"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
	IsActive    *bool   `json:"is_active"`
}
