package models

import (
	"time"

	"github.com/google/uuid"
)

type Author struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Bio       *string   `json:"bio,omitempty"`
	Avatar    *string   `json:"avatar,omitempty"`
	Email     *string   `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateAuthorRequest struct {
	Name   string  `json:"name" validate:"required,min=2,max=200"`
	Slug   string  `json:"slug" validate:"required,min=2,max=200"`
	Bio    *string `json:"bio,omitempty"`
	Avatar *string `json:"avatar,omitempty"`
	Email  *string `json:"email,omitempty" validate:"omitempty,email"`
}

type UpdateAuthorRequest struct {
	Name   *string `json:"name,omitempty" validate:"omitempty,min=2,max=200"`
	Slug   *string `json:"slug,omitempty" validate:"omitempty,min=2,max=200"`
	Bio    *string `json:"bio,omitempty"`
	Avatar *string `json:"avatar,omitempty"`
	Email  *string `json:"email,omitempty" validate:"omitempty,email"`
}
