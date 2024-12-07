package service

import (
	"context"
	"testing"

	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/service"
	"github.com/stretchr/testify/assert"
)

func TestHealthcheckProduction(t *testing.T) {
	mockConfig := config.Config{
		Environment: "production",
	}

	service := service.NewHealthcheckService(mockConfig)
	response := service.HandleHealthcheck(context.Background())
	assert.Equal(t, "available", response.Status)

	assert.Equal(t, "production", response.SystemInfo["environment"])
}

func TestHealthcheckDevelopment(t *testing.T) {
	mockConfig := config.Config{
		Environment: "development",
	}

	service := service.NewHealthcheckService(mockConfig)
	response := service.HandleHealthcheck(context.Background())
	assert.Equal(t, "available", response.Status)

	assert.Equal(t, "development", response.SystemInfo["environment"])
}
