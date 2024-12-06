package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hoyci/book-store-api/service"
	"github.com/hoyci/book-store-api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBookRepository struct {
	mock.Mock
}

func (m *MockBookRepository) Create(ctx context.Context, book types.CreateBookPayload) (int64, error) {
	args := m.Called(ctx, book)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBookRepository) GetByID(ctx context.Context, id int) (*types.Book, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.Book), args.Error(1)
}

func (m *MockBookRepository) DeleteByID(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreate(t *testing.T) {
	mockRepo := new(MockBookRepository)
	bookService := service.NewBookService(mockRepo)

	payload := types.CreateBookPayload{
		Name:          "Test Book",
		Description:   "A great book",
		Author:        "John Doe",
		Genres:        []string{"Fiction"},
		ReleaseYear:   2024,
		NumberOfPages: 300,
		ImageUrl:      "http://example.com/book.jpg",
	}

	t.Run("successful book creation", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, payload).Once().Return(int64(1), nil)

		id, err := bookService.Create(context.Background(), payload)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), id)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error during creation", func(t *testing.T) {
		mockRepo.On("Create", mock.Anything, payload).Once().Return(int64(0), errors.New("repository error"))

		id, err := bookService.Create(context.Background(), payload)

		assert.Error(t, err)
		assert.Equal(t, int64(0), id)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetByID(t *testing.T) {
	mockRepo := new(MockBookRepository)
	bookService := service.NewBookService(mockRepo)

	expectedBook := &types.Book{
		ID:            1,
		Name:          "Test Book",
		Description:   "A great book",
		Author:        "John Doe",
		Genres:        []string{"Fiction"},
		ReleaseYear:   2024,
		NumberOfPages: 300,
		ImageUrl:      "http://example.com/book.jpg",
		CreatedAt:     time.Now(),
	}

	t.Run("successful book retrieval", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, 1).Once().Return(expectedBook, nil)

		book, err := bookService.GetByID(context.Background(), 1)

		assert.NoError(t, err)
		assert.Equal(t, expectedBook, book)
		mockRepo.AssertExpectations(t)
	})

	t.Run("book not found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, 2).Once().Return((*types.Book)(nil), errors.New("book not found"))

		book, err := bookService.GetByID(context.Background(), 2)

		assert.Error(t, err)
		assert.Nil(t, book)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error during GetByID", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, 3).Once().Return((*types.Book)(nil), errors.New("repository error"))

		book, err := bookService.GetByID(context.Background(), 3)

		assert.Error(t, err)
		assert.Nil(t, book)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteByID(t *testing.T) {
	mockRepo := new(MockBookRepository)
	bookService := service.NewBookService(mockRepo)

	t.Run("delete book successfully", func(t *testing.T) {
		mockRepo.On("DeleteByID", mock.Anything, 1).Once().Return(nil)

		err := bookService.DeleteByID(context.Background(), 1)

		assert.NoError(t, err)
		assert.Equal(t, err, nil)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error during GetByID", func(t *testing.T) {
		mockRepo.On("DeleteByID", mock.Anything, 3).Once().Return(errors.New("repository error"))

		err := bookService.DeleteByID(context.Background(), 3)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
