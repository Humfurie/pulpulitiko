package models

import (
	"time"

	"github.com/google/uuid"
)

type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
	ArticleStatusArchived  ArticleStatus = "archived"
)

type Article struct {
	ID                  uuid.UUID     `json:"id"`
	Slug                string        `json:"slug"`
	Title               string        `json:"title"`
	Summary             *string       `json:"summary,omitempty"`
	Content             string        `json:"content"`
	FeaturedImage       *string       `json:"featured_image,omitempty"`
	AuthorID            *uuid.UUID    `json:"author_id,omitempty"`
	CategoryID          *uuid.UUID    `json:"category_id,omitempty"`
	PrimaryPoliticianID *uuid.UUID    `json:"primary_politician_id,omitempty"`
	Status              ArticleStatus `json:"status"`
	ViewCount           int           `json:"view_count"`
	PublishedAt         *time.Time    `json:"published_at,omitempty"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`

	// Relations (populated when needed)
	Author               *Author      `json:"author,omitempty"`
	Category             *Category    `json:"category,omitempty"`
	Tags                 []Tag        `json:"tags,omitempty"`
	PrimaryPolitician    *Politician  `json:"primary_politician,omitempty"`
	MentionedPoliticians []Politician `json:"mentioned_politicians,omitempty"`
}

type ArticleListItem struct {
	ID            uuid.UUID     `json:"id"`
	Slug          string        `json:"slug"`
	Title         string        `json:"title"`
	Summary       *string       `json:"summary,omitempty"`
	FeaturedImage *string       `json:"featured_image,omitempty"`
	Status        ArticleStatus `json:"status"`
	ViewCount     int           `json:"view_count"`
	PublishedAt   *time.Time    `json:"published_at,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`

	AuthorName            *string `json:"author_name,omitempty"`
	AuthorSlug            *string `json:"author_slug,omitempty"`
	AuthorAvatar          *string `json:"author_avatar,omitempty"`
	CategoryName          *string `json:"category_name,omitempty"`
	CategorySlug          *string `json:"category_slug,omitempty"`
	PrimaryPoliticianName *string `json:"primary_politician_name,omitempty"`
	PrimaryPoliticianSlug *string `json:"primary_politician_slug,omitempty"`
}

type CreateArticleRequest struct {
	Slug                string   `json:"slug" validate:"required,min=3,max=255"`
	Title               string   `json:"title" validate:"required,min=3,max=500"`
	Summary             *string  `json:"summary,omitempty"`
	Content             string   `json:"content" validate:"required"`
	FeaturedImage       *string  `json:"featured_image,omitempty"`
	AuthorID            *string  `json:"author_id,omitempty" validate:"omitempty,uuid"`
	CategoryID          *string  `json:"category_id,omitempty" validate:"omitempty,uuid"`
	PrimaryPoliticianID *string  `json:"primary_politician_id,omitempty" validate:"omitempty,uuid"`
	Status              string   `json:"status,omitempty" validate:"omitempty,oneof=draft published archived"`
	TagIDs              []string `json:"tag_ids,omitempty" validate:"omitempty,dive,uuid"`
	PoliticianIDs       []string `json:"politician_ids,omitempty" validate:"omitempty,dive,uuid"`
}

type UpdateArticleRequest struct {
	Slug                *string  `json:"slug,omitempty" validate:"omitempty,min=3,max=255"`
	Title               *string  `json:"title,omitempty" validate:"omitempty,min=3,max=500"`
	Summary             *string  `json:"summary,omitempty"`
	Content             *string  `json:"content,omitempty"`
	FeaturedImage       *string  `json:"featured_image,omitempty"`
	AuthorID            *string  `json:"author_id,omitempty" validate:"omitempty,uuid"`
	CategoryID          *string  `json:"category_id,omitempty" validate:"omitempty,uuid"`
	PrimaryPoliticianID *string  `json:"primary_politician_id,omitempty" validate:"omitempty,uuid"`
	Status              *string  `json:"status,omitempty" validate:"omitempty,oneof=draft published archived"`
	TagIDs              []string `json:"tag_ids,omitempty" validate:"omitempty,dive,uuid"`
	PoliticianIDs       []string `json:"politician_ids,omitempty" validate:"omitempty,dive,uuid"`
}

type ArticleFilter struct {
	Status         *ArticleStatus
	CategoryID     *uuid.UUID
	TagID          *uuid.UUID
	AuthorID       *uuid.UUID
	PoliticianID   *uuid.UUID // Filter by primary or mentioned politician
	Search         *string
	IncludeDeleted bool
}

type PaginatedArticles struct {
	Articles   []ArticleListItem `json:"articles"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int               `json:"total_pages"`
}
