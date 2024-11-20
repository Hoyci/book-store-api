package service

import (
	"context"

	"github.com/hoyci/book-store-api/repository"
	"github.com/hoyci/book-store-api/types"
)

type BookService struct {
	BookRepository *repository.BookStore
}

func NewBookService(repo *repository.BookStore) *BookService {
	return &BookService{BookRepository: repo}
}

func (s *BookService) Create(ctx context.Context, payload types.CreateBookPayload) (int64, error) {
	book := types.Book{
		Name:          payload.Name,
		Description:   payload.Description,
		Author:        payload.Author,
		Genres:        payload.Genres,
		ReleaseYear:   payload.ReleaseYear,
		NumberOfPages: payload.NumberOfPages,
		ImageUrl:      payload.ImageUrl,
	}

	id, err := s.BookRepository.Create(ctx, book)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *BookService) GetByID(ctx context.Context, id int) (*types.Book, error) {
	book, err := s.BookRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return book, nil
}
