package types

import (
	"context"
	"time"
)

type BookStore interface {
	Create(ctx context.Context, book CreateBookPayload) (int, error)
	GetByID(ctx context.Context, id int) (*Book, error)
	GetMany(ctx context.Context) ([]*Book, error)
	UpdateByID(ctx context.Context, id int, book UpdateBookPayload) (*Book, error)
	DeleteByID(ctx context.Context, id int) error
}

type Book struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Author        string     `json:"author"`
	Genres        []string   `json:"genres"`
	ReleaseYear   int        `json:"release_year"`
	NumberOfPages int        `json:"number_of_pages"`
	ImageUrl      string     `json:"image_url"`
	CreatedAt     time.Time  `json:"created_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
}

type CreateBookPayload struct {
	Name          string   `json:"name" validate:"required,min=3"`
	Description   string   `json:"description" validate:"required,min=5"`
	Author        string   `json:"author" validate:"required,min=3"`
	Genres        []string `json:"genres" validate:"required,dive,min=1"`
	ReleaseYear   int      `json:"release_year" validate:"required,gte=1500,lte=2099"`
	NumberOfPages int      `json:"number_of_pages" validate:"required,gte=1"`
	ImageUrl      string   `json:"image_url" validate:"required,url"`
}

type CreateBookResponse struct {
	ID int `json:"id"`
}

type UpdateBookPayload struct {
	Name          string   `json:"name" validate:"required,min=3"`
	Description   string   `json:"description" validate:"required,min=5"`
	Author        string   `json:"author" validate:"required,min=3"`
	Genres        []string `json:"genres" validate:"required,dive,min=2"`
	ReleaseYear   int      `json:"release_year" validate:"required,gte=1500,lte=2099"`
	NumberOfPages int      `json:"number_of_pages" validate:"required,gte=1"`
	ImageUrl      string   `json:"image_url" validate:"required,url"`
}

type DeleteBookByIDResponse struct {
	ID int `json:"id"`
}

type GetBooksResponse struct {
	Books []*Book `json:"books"`
}
