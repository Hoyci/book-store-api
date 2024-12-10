package service

import (
	"context"

	"github.com/hoyci/book-store-api/types"
)

type BookServiceInterface interface {
	Create(ctx context.Context, payload types.CreateBookPayload) (int, error)
	GetByID(ctx context.Context, id int) (*types.Book, error)
	UpdateByID(ctx context.Context, id int, newBook types.UpdateBookPayload) (*types.Book, error)
	DeleteByID(ctx context.Context, id int) (int, error)
}
