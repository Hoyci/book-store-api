package service

import (
	"context"

	"github.com/hoyci/book-store-api/repository"
	"github.com/hoyci/book-store-api/types"
)

type BookService struct {
	BookRepository repository.BookRepositoryInterface
}

func NewBookService(repo repository.BookRepositoryInterface) *BookService {
	return &BookService{BookRepository: repo}
}

func (s *BookService) Create(ctx context.Context, payload types.CreateBookPayload) (int, error) {
	return s.BookRepository.Create(ctx, payload)
	// Here I can add send email when a user is created
}

func (s *BookService) GetByID(ctx context.Context, id int) (*types.Book, error) {
	return s.BookRepository.GetByID(ctx, id)
}

func (s *BookService) UpdateByID(ctx context.Context, id int, newBook types.UpdateBookPayload) (*types.Book, error) {
	return s.BookRepository.UpdateByID(ctx, id, newBook)
}

func (s *BookService) DeleteByID(ctx context.Context, id int) (int, error) {
	return s.BookRepository.DeleteByID(ctx, id)
	// Here I can add send email when a user is created
}
