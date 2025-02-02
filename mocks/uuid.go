package mocks

import "github.com/stretchr/testify/mock"

type MockUUIDGenerator struct {
	mock.Mock
}

func (m *MockUUIDGenerator) New() string {
	args := m.Called()
	return args.String(0)
}
