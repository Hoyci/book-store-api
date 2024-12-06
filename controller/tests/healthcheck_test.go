package controller

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hoyci/book-store-api/cmd/api"
	"github.com/hoyci/book-store-api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockhHealthcheckService struct {
	mock.Mock
}

func (m *MockhHealthcheckService) HandleHealthcheck(ctx context.Context) *types.HealthcheckResponse {
	args := m.Called(ctx)
	return args.Get(0).(*types.HealthcheckResponse)
}

func TestHandleHealthChec(t *testing.T) {
	mockService := new(MockhHealthcheckService)

	apiServer := api.NewApiServer(":8080", nil)
	router := apiServer.SetupRouter(mockService, nil)

	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("it should return environment as production", func(t *testing.T) {
		mockService.On("HandleHealthcheck", mock.Anything).Return(&types.HealthcheckResponse{
			Status: "available",
			SystemInfo: map[string]string{
				"environment": "production",
			},
		})

		req := httptest.NewRequest(http.MethodGet, ts.URL+"/api/v1/healthcheck", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)

		assert.Equal(t, http.StatusOK, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"status":"available","system_info":{"environment":"production"}}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

}
