package service

import (
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

func (s *HealthcheckService) CheckHealth() types.HealthCheckResponse {
	return types.HealthCheckResponse{
		Status: "available",
		SystemInfo: map[string]string{
			"environment": s.Config.Environment,
		},
	}
}
