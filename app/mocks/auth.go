package mocks

import (
	"context"

	"github.com/hoyci/book-store-api/types"
	"github.com/stretchr/testify/mock"
)

type MockAuthStore struct {
	mock.Mock
}

func (m *MockAuthStore) GetRefreshTokenByUserID(ctx context.Context, userID int) (*types.RefreshToken, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*types.RefreshToken), args.Error(1)
}

func (m *MockAuthStore) UpsertRefreshToken(ctx context.Context, payload types.UpdateRefreshTokenPayload) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}
