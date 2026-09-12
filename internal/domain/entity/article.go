package entity

import (
	"time"

	"github.com/google/uuid"
)

type Article struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     *string    `json:"excerpt"`
	Content     string     `json:"content"`
	CoverImage  *string    `json:"cover_image"`
	Status      string     `json:"status"`
	ViewCount   int32      `json:"view_count"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CategoryID  *uuid.UUID `json:"category_id"`
	WebsiteID   *uuid.UUID `json:"website_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Metadata sumber berita (fitur scheduler AI)
	SourceURL      *string `json:"source_url"`
	SourceType     *string `json:"source_type"`
	SourceName     *string `json:"source_name"`
	SourceHash     *string `json:"source_hash"`
	IsAiGenerated  bool    `json:"is_ai_generated"`
	ImageCredit    *string `json:"image_credit"`
	ImageSourceURL *string `json:"image_source_url"`
	ImageLicense   *string `json:"image_license"`

	// Joined fields
	AuthorName   *string `json:"author_name"`
	AuthorAvatar *string `json:"author_avatar"`
	CategoryName *string `json:"category_name"`
	CategorySlug *string `json:"category_slug"`
	WebsiteName  *string `json:"website_name"`
	WebsiteSlug  *string `json:"website_slug"`
	Tags         []Tag   `json:"tags,omitempty"`
}

type ArticleCreate struct {
	Title       string      `json:"title" binding:"required"`
	Slug        string      `json:"slug" binding:"required"`
	Excerpt     *string     `json:"excerpt"`
	Content     string      `json:"content" binding:"required"`
	CoverImage  *string     `json:"cover_image"`
	Status      string      `json:"status"`
	PublishedAt *time.Time  `json:"published_at"`
	CategoryID  *uuid.UUID  `json:"category_id"`
	WebsiteID   *uuid.UUID  `json:"website_id" binding:"required"`
	TagIDs      []uuid.UUID `json:"tag_ids"`
}

type ArticleUpdate struct {
	Title      *string     `json:"title"`
	Slug       *string     `json:"slug"`
	Excerpt    *string     `json:"excerpt"`
	Content    *string     `json:"content"`
	CoverImage *string     `json:"cover_image"`
	Status     *string     `json:"status"`
	CategoryID *uuid.UUID  `json:"category_id"`
	WebsiteID  *uuid.UUID  `json:"website_id"`
	TagIDs     []uuid.UUID `json:"tag_ids"`
}

type ArticleQuery struct {
	Page       int
	Limit      int
	Offset     int
	Status     string
	CategoryID *uuid.UUID
	WebsiteID  *uuid.UUID
	TagID      *uuid.UUID
}
