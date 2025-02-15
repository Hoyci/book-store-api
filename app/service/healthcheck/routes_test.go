package healthcheck_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hoyci/book-store-api/cmd/api"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/service/healthcheck"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	t.Run("should return environment as production", func(t *testing.T) {
		mockConfig := config.Config{
			Environment: "production",
		}
		healthCheckHandler := healthcheck.NewHealthCheckHandler(mockConfig)

		apiServer := api.NewApiServer(":8080", nil)
		router := apiServer.SetupRouter(healthCheckHandler, nil, nil, nil)

		ts := httptest.NewServer(router)
		defer ts.Close()

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

	t.Run("should return environment as production", func(t *testing.T) {
		mockConfig := config.Config{
			Environment: "production",
		}
		healthCheckHandler := healthcheck.NewHealthCheckHandler(mockConfig)

		apiServer := api.NewApiServer(":8080", nil)
		router := apiServer.SetupRouter(healthCheckHandler, nil, nil, nil)

		ts := httptest.NewServer(router)
		defer ts.Close()

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
