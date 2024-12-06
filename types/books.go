package types

import (
	"context"
	"time"
)

type BookRepository interface {
	Create(ctx context.Context, book Book) (int64, error)
	GetByID(ctx context.Context, id int) (*Book, error)
}

type Book struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Author        string     `json:"author"`
	Genres        []string   `json:"genres"`
	ReleaseYear   int32      `json:"releaseYear"`
	NumberOfPages int32      `json:"numberOfPages"`
	ImageUrl      string     `json:"imageUrl"`
	CreatedAt     time.Time  `json:"createdAt"`
	DeletedAt     *time.Time `json:"deletedAt"`
}

type CreateBookPayload struct {
	Name          string   `json:"name" validate:"required"`
	Description   string   `json:"description" validate:"required"`
	Author        string   `json:"author" validate:"required"`
	Genres        []string `json:"genres" validate:"required"`
	ReleaseYear   int32    `json:"releaseYear" validate:"required"`
	NumberOfPages int32    `json:"numberOfPages" validate:"required"`
	ImageUrl      string   `json:"imageUrl" validate:"required"`
}

type BookUpdatePayload struct {
	Name          *string   `json:"name"`
	Description   *string   `json:"description"`
	Author        *string   `json:"author"`
	Genres        *[]string `json:"genres"`
	ReleaseYear   *int32    `json:"releaseYear"`
	NumberOfPages *int32    `json:"numberOfPages"`
	ImageUrl      *string   `json:"imageUrl"`
}
