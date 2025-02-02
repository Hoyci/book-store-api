package mocks

import (
	"context"

	"github.com/hoyci/book-store-api/types"
	"github.com/stretchr/testify/mock"
)

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Create(ctx context.Context, user types.CreateUserDatabasePayload) (*types.UserResponse, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*types.UserResponse), args.Error(1)
}

func (m *MockUserStore) GetByID(ctx context.Context, userID int) (*types.UserResponse, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*types.UserResponse), args.Error(1)
}

func (m *MockUserStore) GetByEmail(ctx context.Context, email string) (*types.UserResponse, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*types.UserResponse), args.Error(1)
}

func (m *MockUserStore) UpdateByID(ctx context.Context, userID int, newUser types.UpdateUserPayload) (*types.UserResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).(*types.UserResponse), args.Error(1)
}

func (m *MockUserStore) DeleteByID(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
