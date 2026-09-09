package entity

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type TagCreate struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

type TagUpdate struct {
	Name *string `json:"name"`
	Slug *string `json:"slug"`
}

type Comment struct {
	ID         uuid.UUID  `json:"id"`
	ArticleID  uuid.UUID  `json:"article_id"`
	UserID     uuid.UUID  `json:"user_id"`
	ParentID   *uuid.UUID `json:"parent_id"`
	Content    string     `json:"content"`
	IsApproved bool       `json:"is_approved"`
	CreatedAt  time.Time  `json:"created_at"`

	UserName   string  `json:"user_name"`
	UserAvatar *string `json:"user_avatar"`
}

type CommentCreate struct {
	ArticleID uuid.UUID  `json:"article_id" binding:"required"`
	ParentID  *uuid.UUID `json:"parent_id"`
	Content   string     `json:"content" binding:"required"`
}
