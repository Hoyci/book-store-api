package service

import (
	"context"

	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/types"
)

type HealthcheckService struct {
	Config config.Config
}

func NewHealthcheckService(cfg config.Config) *HealthcheckService {
	return &HealthcheckService{
		Config: cfg,
	}
}

func (s *HealthcheckService) HandleHealthcheck(ctx context.Context) *types.HealthcheckResponse {
	return &types.HealthcheckResponse{
		Status: "available",
		SystemInfo: map[string]string{
			"environment": s.Config.Environment,
		},
	}
}
