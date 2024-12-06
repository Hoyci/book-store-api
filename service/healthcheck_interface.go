package service

import (
	"context"

	"github.com/hoyci/book-store-api/types"
)

type HealthcheckServiceInterface interface {
	HandleHealthcheck(ctx context.Context) *types.HealthcheckResponse
}
