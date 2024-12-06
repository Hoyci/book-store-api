package repository

import (
	"context"

	"github.com/hoyci/book-store-api/types"
)

type BookRepositoryInterface interface {
	Create(ctx context.Context, book types.CreateBookPayload) (int64, error)
	GetByID(ctx context.Context, id int) (*types.Book, error)
	DeleteByID(ctx context.Context, id int) error
}
