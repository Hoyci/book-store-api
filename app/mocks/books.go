package mocks

import (
	"context"

	"github.com/hoyci/book-store-api/types"
	"github.com/stretchr/testify/mock"
)

type MockBookStore struct {
	mock.Mock
}

func (m *MockBookStore) Create(ctx context.Context, book types.CreateBookPayload) (int, error) {
	args := m.Called(ctx, book)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockBookStore) GetByID(ctx context.Context, id int) (*types.Book, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.Book), args.Error(1)
}

func (m *MockBookStore) GetMany(ctx context.Context) ([]*types.Book, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*types.Book), args.Error(1)
}

func (m *MockBookStore) UpdateByID(ctx context.Context, id int, newBook types.UpdateBookPayload) (*types.Book, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.Book), args.Error(1)
}

func (m *MockBookStore) DeleteByID(ctx context.Context, id int) error {
	args := m.Called(ctx, id)

	return args.Error(0)
}
