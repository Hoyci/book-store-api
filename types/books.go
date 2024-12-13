package types

import (
	"context"
	"time"
)

type BookStore interface {
	Create(ctx context.Context, book CreateBookPayload) (int, error)
	GetByID(ctx context.Context, id int) (*Book, error)
	UpdateByID(ctx context.Context, id int, book UpdateBookPayload) (*Book, error)
	DeleteByID(ctx context.Context, id int) (int, error)
}

type Book struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Author        string     `json:"author"`
	Genres        []string   `json:"genres"`
	ReleaseYear   int        `json:"releaseYear"`
	NumberOfPages int        `json:"numberOfPages"`
	ImageUrl      string     `json:"imageUrl"`
	CreatedAt     time.Time  `json:"createdAt"`
	DeletedAt     *time.Time `json:"deletedAt"`
	UpdatedAt     *time.Time `json:"updatedAt"`
}

type CreateBookPayload struct {
	Name          string   `json:"name" validate:"required"`
	Description   string   `json:"description" validate:"required"`
	Author        string   `json:"author" validate:"required"`
	Genres        []string `json:"genres" validate:"required"`
	ReleaseYear   int      `json:"releaseYear" validate:"required"`
	NumberOfPages int      `json:"numberOfPages" validate:"required"`
	ImageUrl      string   `json:"imageUrl" validate:"required"`
}

type UpdateBookPayload struct {
	Name          *string   `json:"name,omitempty" validate:"omitempty,min=3"`
	Description   *string   `json:"description,omitempty" validate:"omitempty,min=5"`
	Author        *string   `json:"author,omitempty" validate:"omitempty,min=3"`
	Genres        *[]string `json:"genres,omitempty" validate:"omitempty,dive,min=1"`
	ReleaseYear   *int      `json:"releaseYear,omitempty" validate:"omitempty,gte=1500,lte=2099"`
	NumberOfPages *int      `json:"numberOfPages,omitempty" validate:"omitempty,gte=1"`
	ImageUrl      *string   `json:"imageUrl,omitempty" validate:"omitempty,url"`
}
