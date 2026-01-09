package models

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateTagRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	Slug string `json:"slug" validate:"required,min=2,max=100"`
}

type UpdateTagRequest struct {
	Name *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Slug *string `json:"slug,omitempty" validate:"omitempty,min=2,max=100"`
}

type TagFilter struct {
	Search    *string
	SortBy    *string // name, created_at
	SortOrder *string // asc, desc
}

type PaginatedTags struct {
	Tags       []Tag `json:"data"`
	Total      int   `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}
