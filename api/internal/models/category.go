package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Slug        string  `json:"slug" validate:"required,min=2,max=100"`
	Description *string `json:"description,omitempty"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Slug        *string `json:"slug,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty"`
}

type CategoryFilter struct {
	Search    *string
	SortBy    *string // name, created_at
	SortOrder *string // asc, desc
}

type PaginatedCategories struct {
	Categories []Category `json:"data"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
	TotalPages int        `json:"total_pages"`
}
