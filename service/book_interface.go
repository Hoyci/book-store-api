package service

import (
	"context"

	"github.com/hoyci/book-store-api/types"
)

type BookServiceInterface interface {
	Create(ctx context.Context, payload types.CreateBookPayload) (int64, error)
	GetByID(ctx context.Context, id int) (*types.Book, error)
	DeleteByID(ctx context.Context, id int) error
}
