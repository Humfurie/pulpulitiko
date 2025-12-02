package models

import (
	"time"

	"github.com/google/uuid"
)

// SocialLinks represents social media profile URLs
type SocialLinks struct {
	Twitter   string `json:"twitter,omitempty"`
	Facebook  string `json:"facebook,omitempty"`
	LinkedIn  string `json:"linkedin,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	YouTube   string `json:"youtube,omitempty"`
	TikTok    string `json:"tiktok,omitempty"`
	Website   string `json:"website,omitempty"`
}

type Author struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Slug        string       `json:"slug"`
	Bio         *string      `json:"bio,omitempty"`
	Avatar      *string      `json:"avatar,omitempty"`
	Email       *string      `json:"email,omitempty"`
	Phone       *string      `json:"phone,omitempty"`
	Address     *string      `json:"address,omitempty"`
	SocialLinks *SocialLinks `json:"social_links,omitempty"`
	RoleID      *uuid.UUID   `json:"role_id,omitempty"`
	Role        string       `json:"role"`      // Role slug from join with roles table
	IsSystem    bool         `json:"is_system"` // System users cannot be deleted
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty"`
}

type CreateAuthorRequest struct {
	Name        string       `json:"name" validate:"required,min=2,max=200"`
	Slug        string       `json:"slug" validate:"required,min=2,max=200"`
	Bio         *string      `json:"bio,omitempty"`
	Avatar      *string      `json:"avatar,omitempty"`
	Email       *string      `json:"email,omitempty" validate:"omitempty,email"`
	Phone       *string      `json:"phone,omitempty" validate:"omitempty,max=50"`
	Address     *string      `json:"address,omitempty"`
	SocialLinks *SocialLinks `json:"social_links,omitempty"`
	RoleID      *string      `json:"role_id,omitempty"`
	Role        *string      `json:"role,omitempty"` // Role slug for convenience
}

type UpdateAuthorRequest struct {
	Name        *string      `json:"name,omitempty" validate:"omitempty,min=2,max=200"`
	Slug        *string      `json:"slug,omitempty" validate:"omitempty,min=2,max=200"`
	Bio         *string      `json:"bio,omitempty"`
	Avatar      *string      `json:"avatar,omitempty"`
	Email       *string      `json:"email,omitempty" validate:"omitempty,email"`
	Phone       *string      `json:"phone,omitempty" validate:"omitempty,max=50"`
	Address     *string      `json:"address,omitempty"`
	SocialLinks *SocialLinks `json:"social_links,omitempty"`
	RoleID      *string      `json:"role_id,omitempty"`
	Role        *string      `json:"role,omitempty"` // Role slug for convenience
}

// UserProfile represents a public user profile with comment activity
type UserProfile struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	Avatar       *string    `json:"avatar,omitempty"`
	Bio          *string    `json:"bio,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	CommentCount int        `json:"comment_count"`
	ReplyCount   int        `json:"reply_count"`
}
