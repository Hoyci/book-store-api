package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hoyci/book-store-api/cmd/api"
	"github.com/hoyci/book-store-api/repository"
	"github.com/hoyci/book-store-api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBookService struct {
	mock.Mock
}

func (m *MockBookService) Create(ctx context.Context, payload types.CreateBookPayload) (int64, error) {
	args := m.Called(ctx, payload)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBookService) GetByID(ctx context.Context, id int) (*types.Book, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.Book), args.Error(1)
}

func TestHandleGetBookByID(t *testing.T) {
	mockService := new(MockBookService)

	apiServer := api.NewApiServer(":8080", nil)
	router := apiServer.SetupRouter(nil, mockService)

	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("it should throw an error when call endpoint without book ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, ts.URL+"/api/v1/book", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
	})

	t.Run("it should throw an error when call endpoint with wrong ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, ts.URL+"/api/v1/book/johndoe", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":"invalid book ID: must be a positive integer"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should return succssefully status and body when call endpoint with valid body", func(t *testing.T) {
		mockService.On("GetByID", mock.Anything, 1).Return(&types.Book{
			ID:            1,
			Name:          "Go Programming",
			Description:   "A book about Go programming",
			Author:        "John Doe",
			Genres:        []string{"Programming"},
			ReleaseYear:   2024,
			NumberOfPages: 300,
			ImageUrl:      "http://example.com/go.jpg",
		}, nil)

		req := httptest.NewRequest(http.MethodGet, ts.URL+"/api/v1/book/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{
			"id": 1,
			"name": "Go Programming",
			"description": "A book about Go programming",
			"author": "John Doe",
			"genres": ["Programming"],
			"releaseYear": 2024,
			"numberOfPages": 300,
			"imageUrl": "http://example.com/go.jpg",
			"createdAt": "0001-01-01T00:00:00Z"
		}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})
}

func TestHandleCreateook(t *testing.T) {
	mockService := new(MockBookService)

	apiServer := api.NewApiServer(":8080", nil)
	router := apiServer.SetupRouter(nil, mockService)

	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("it should throw an error when body is not a valid JSON", func(t *testing.T) {
		invalidBody := bytes.NewReader([]byte("INVALID JSON"))
		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/book", invalidBody)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":"invalid input"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when body is a valid JSON but missing key", func(t *testing.T) {
		mockService.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)

		payload := types.CreateBookPayload{
			// Name:          "Go Programming",
			Description:   "A book about Go programming",
			Author:        "John Doe",
			Genres:        []string{"Programming"},
			ReleaseYear:   2024,
			NumberOfPages: 300,
			ImageUrl:      "http://example.com/go.jpg",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/book", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":"[Field 'Name' is invalid: required]"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should successfully create a book", func(t *testing.T) {
		mockService.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)

		payload := types.CreateBookPayload{
			Name:          "Go Programming",
			Description:   "A book about Go programming",
			Author:        "John Doe",
			Genres:        []string{"Programming"},
			ReleaseYear:   2024,
			NumberOfPages: 300,
			ImageUrl:      "http://example.com/go.jpg",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/book", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusCreated, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"id":1}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw a database insert error", func(t *testing.T) {
		mockService.On("Create", mock.Anything, mock.Anything).Return(int64(0), &repository.InsertError{
			Entity: "Go Programming",
			Err:    errors.New("duplicate key"),
		})

		payload := types.CreateBookPayload{
			Name:          "Go Programming",
			Description:   "A book about Go programming",
			Author:        "John Doe",
			Genres:        []string{"Programming"},
			ReleaseYear:   2024,
			NumberOfPages: 300,
			ImageUrl:      "http://example.com/go.jpg",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/book", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

		responseBody, _ := io.ReadAll(res.Body)
		expected := `{"error":"failed to insert entity 'Go Programming': duplicate key"}`
		assert.JSONEq(t, expected, string(responseBody))
	})
}
