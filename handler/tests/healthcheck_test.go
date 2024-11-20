package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/handler"
	"github.com/hoyci/book-store-api/service"
	"github.com/stretchr/testify/assert"
)

func TestHandleHealthcheckProduction(t *testing.T) {
	mockConfig := config.Config{
		Environment: "production",
	}

	mockService := service.NewHealthcheckService(mockConfig)
	handler := handler.NewHealthcheckHandler(mockService)

	req, err := http.NewRequest(http.MethodGet, "/healthcheck", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.HandleHealthcheck(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	expectedResponse := `{"status":"available","system_info":{"environment":"production"}}`
	assert.JSONEq(t, expectedResponse, rr.Body.String())
}

func TestHandleHealthcheckDevelopment(t *testing.T) {
	mockConfig := config.Config{
		Environment: "development",
	}

	mockService := service.NewHealthcheckService(mockConfig)
	handler := handler.NewHealthcheckHandler(mockService)

	req, err := http.NewRequest(http.MethodGet, "/healthcheck", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.HandleHealthcheck(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	expectedResponse := `{"status":"available","system_info":{"environment":"development"}}`
	assert.JSONEq(t, expectedResponse, rr.Body.String())
}
