package book_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/cmd/api"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/service/book"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBookStore struct {
	mock.Mock
}

func (m *MockBookStore) Create(ctx context.Context, book types.CreateBookPayload) (int, error) {
	args := m.Called(ctx, book)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockBookStore) GetByID(ctx context.Context, id int) (*types.Book, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.Book), args.Error(1)
}

func (m *MockBookStore) UpdateByID(ctx context.Context, id int, newBook types.UpdateBookPayload) (*types.Book, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.Book), args.Error(1)
}

func (m *MockBookStore) DeleteByID(ctx context.Context, id int) (int, error) {
	args := m.Called(ctx, id)
	if val, ok := args.Get(0).(int); ok {
		return val, args.Error(1)
	}
	return 0, args.Error(1)
}

func generateTestToken(userID int, username, email string) string {
	claims := types.CustomClaims{
		ID:       "mocked-id",
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(1 * time.Hour)},
		},
	}
	token, err := utils.CreateJWTFromClaims(claims, config.Envs.JWTSecret)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate test token: %v", err))
	}
	return token
}

func TestHandleCreateBook(t *testing.T) {
	setupTestServer := func() (*MockBookStore, *httptest.Server, *mux.Router) {
		mockBookStore := new(MockBookStore)
		mockBookHandler := book.NewBookHandler(mockBookStore)
		apiServer := api.NewApiServer(":8080", nil)
		router := apiServer.SetupRouter(nil, mockBookHandler, nil, nil)
		ts := httptest.NewServer(router)
		return mockBookStore, ts, router
	}

	t.Run("it should throw an error when body is not a valid JSON", func(t *testing.T) {
		_, ts, router := setupTestServer()
		defer ts.Close()

		invalidBody := bytes.NewReader([]byte("INVALID JSON"))
		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/books", invalidBody)
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

		expectedResponse := `{"error":"Body is not a valid json"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when body is a valid JSON but missing key", func(t *testing.T) {
		_, ts, router := setupTestServer()
		defer ts.Close()

		payload := types.CreateBookPayload{
			Description:   "A book about Go programming",
			Author:        "John Doe",
			Genres:        []string{"Programming"},
			ReleaseYear:   2024,
			NumberOfPages: 300,
			ImageUrl:      "http://example.com/go.jpg",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/books", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":["Field 'Name' is invalid: required"]}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw a database insert error", func(t *testing.T) {
		mockBookStore, ts, router := setupTestServer()
		defer ts.Close()

		mockBookStore.On("Create", mock.Anything, mock.Anything).Return(0, fmt.Errorf("failed to insert entity 'book': database error"))

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

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/books", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

		responseBody, _ := io.ReadAll(res.Body)
		expected := `{"error":"An unexpected error occurred"}`
		assert.JSONEq(t, expected, string(responseBody))
	})

	t.Run("it should successfully create a book", func(t *testing.T) {
		mockBookStore, ts, router := setupTestServer()
		defer ts.Close()

		mockBookStore.On("Create", mock.Anything, mock.Anything).Return(int(1), nil)

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

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/books", bytes.NewBuffer(marshalled))
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
}

func TestHandleGetBookByID(t *testing.T) {
	setupTestServer := func() (*MockBookStore, *httptest.Server, *mux.Router) {
		mockBookStore := new(MockBookStore)
		mockBookHandler := book.NewBookHandler(mockBookStore)
		apiServer := api.NewApiServer(":8080", nil)
		router := apiServer.SetupRouter(nil, mockBookHandler, nil, nil)
		ts := httptest.NewServer(router)
		return mockBookStore, ts, router
	}

	t.Run("it should throw an error when call endpoint without book ID", func(t *testing.T) {
		_, ts, router := setupTestServer()
		defer ts.Close()

		req := httptest.NewRequest(http.MethodGet, ts.URL+"/api/v1/books", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("it should throw an error when call endpoint with wrong book ID", func(t *testing.T) {
		token := generateTestToken(1, "JohnDoe", "johndoe@example.com")
		_, ts, router := setupTestServer()
		defer ts.Close()

		req := httptest.NewRequest(http.MethodGet, ts.URL+"/api/v1/books/johndoe", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":"Book ID must be a positive integer"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when call endpoint with a non-existent book ID", func(t *testing.T) {
		token := generateTestToken(1, "JohnDoe", "johndoe@example.com")
		mockBookStore, ts, router := setupTestServer()
		defer ts.Close()

		mockBookStore.On("GetByID", mock.Anything, 1).Return(&types.Book{}, sql.ErrNoRows)

		req := httptest.NewRequest(http.MethodGet, ts.URL+"/api/v1/books/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error": "No book found with ID 1"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should return succssefully status and body when call endpoint with valid body", func(t *testing.T) {
		token := generateTestToken(1, "JohnDoe", "johndoe@example.com")
		mockBookStore, ts, router := setupTestServer()
		defer ts.Close()

		mockBookStore.On("GetByID", mock.Anything, 1).Return(&types.Book{
			ID:            1,
			Name:          "Go Programming",
			Description:   "A book about Go programming",
			Author:        "John Doe",
			Genres:        []string{"Programming"},
			ReleaseYear:   2024,
			NumberOfPages: 300,
			ImageUrl:      "http://example.com/go.jpg",
			CreatedAt:     time.Date(0001, 01, 01, 0, 0, 0, 0, time.UTC),
			DeletedAt:     nil,
			UpdatedAt:     nil,
		}, nil)

		req := httptest.NewRequest(http.MethodGet, ts.URL+"/api/v1/books/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
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
			"release_year": 2024,
			"number_of_pages": 300,
			"image_url": "http://example.com/go.jpg",
			"created_at": "0001-01-01T00:00:00Z",
			"deleted_at": null,
			"updated_at": null
		}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})
}

func TestHandleUpdateBookByID(t *testing.T) {
	setupTestServer := func() (*MockBookStore, *httptest.Server, *mux.Router) {
		mockBookStore := new(MockBookStore)
		mockBookHandler := book.NewBookHandler(mockBookStore)
		apiServer := api.NewApiServer(":8080", nil)
		router := apiServer.SetupRouter(nil, mockBookHandler, nil, nil)
		ts := httptest.NewServer(router)
		return mockBookStore, ts, router
	}

	t.Run("it should throw an error when call endpoint without book ID", func(t *testing.T) {
		_, ts, router := setupTestServer()
		defer ts.Close()

		req := httptest.NewRequest(http.MethodPut, ts.URL+"/api/v1/books", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("it should throw an error when call endpoint with wrong book ID", func(t *testing.T) {
		_, ts, router := setupTestServer()
		defer ts.Close()

		req := httptest.NewRequest(http.MethodPut, ts.URL+"/api/v1/books/johndoe", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":"Book ID must be a positive integer"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when no fields are provided for update", func(t *testing.T) {
		_, ts, router := setupTestServer()
		defer ts.Close()

		emptyPayload := `{}`
		req := httptest.NewRequest(http.MethodPut, ts.URL+"/api/v1/books/1", bytes.NewBufferString(emptyPayload))
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

		expectedResponse := `{"error":["Field validation for 'UpdateBookPayload' failed on the 'atleastonefield' tag"]}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when body is invalid", func(t *testing.T) {
		_, ts, router := setupTestServer()
		defer ts.Close()

		invalidPayload := `{"name": ""}`
		req := httptest.NewRequest(http.MethodPut, ts.URL+"/api/v1/books/1", bytes.NewBufferString(invalidPayload))
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

		expectedResponse := `{"error":["Field validation for 'Name' failed on the 'min' tag"]}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when call endpoint with a non-existent user ID", func(t *testing.T) {
		mockBookStore, ts, router := setupTestServer()
		defer ts.Close()

		mockBookStore.On("UpdateByID", mock.Anything, 1, mock.Anything).Return(&types.Book{}, sql.ErrNoRows)

		validPayload := `{
			"name": "Go Programming - Updated",
			"genres": ["Programming", "Go"],
			"image_url": "http://example.com/go_updated.jpg"
		}`
		req := httptest.NewRequest(http.MethodPut, ts.URL+"/api/v1/books/1", bytes.NewBufferString(validPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error": "No book found with ID 1"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should return successfully status and body when the book is updated", func(t *testing.T) {
		mockBookStore, ts, router := setupTestServer()
		defer ts.Close()

		mockBookStore.On("UpdateByID", mock.Anything, 1, mock.Anything).Return(&types.Book{
			ID:            1,
			Name:          "Go Programming - Updated",
			Description:   "Updated description",
			Author:        "John Doe",
			Genres:        []string{"Programming", "Go"},
			ReleaseYear:   2024,
			NumberOfPages: 350,
			ImageUrl:      "http://example.com/go_updated.jpg",
			CreatedAt:     time.Date(0001, 01, 01, 0, 0, 0, 0, time.UTC),
			DeletedAt:     nil,
			UpdatedAt:     utils.TimePtr(time.Date(0001, 01, 01, 0, 0, 0, 0, time.UTC)),
		}, nil)

		validPayload := `{
			"name": "Go Programming - Updated",
			"genres": ["Programming", "Go"],
			"image_url": "http://example.com/go_updated.jpg"
		}`
		req := httptest.NewRequest(http.MethodPut, ts.URL+"/api/v1/books/1", bytes.NewBufferString(validPayload))
		req.Header.Set("Content-Type", "application/json")
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
			"name": "Go Programming - Updated",
			"description": "Updated description",
			"author": "John Doe",
			"genres": ["Programming", "Go"],
			"release_year": 2024,
			"number_of_pages": 350,
			"image_url": "http://example.com/go_updated.jpg",
			"created_at": "0001-01-01T00:00:00Z",
			"updated_at": "0001-01-01T00:00:00Z",
			"deleted_at": null 
		}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})
}

func TestHandleDeleteBookByID(t *testing.T) {
	setupTestServer := func() (*MockBookStore, *httptest.Server, *mux.Router) {
		mockBookStore := new(MockBookStore)
		mockBookHandler := book.NewBookHandler(mockBookStore)
		apiServer := api.NewApiServer(":8080", nil)
		router := apiServer.SetupRouter(nil, mockBookHandler, nil, nil)
		ts := httptest.NewServer(router)
		return mockBookStore, ts, router
	}

	t.Run("it should throw an error when call endpoint without book ID", func(t *testing.T) {
		_, ts, router := setupTestServer()
		defer ts.Close()

		req := httptest.NewRequest(http.MethodDelete, ts.URL+"/api/v1/books", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("it should throw an error when call endpoint with wrong ID", func(t *testing.T) {
		_, ts, router := setupTestServer()
		defer ts.Close()

		req := httptest.NewRequest(http.MethodDelete, ts.URL+"/api/v1/books/johndoe", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":"Book ID must be a positive integer"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when call endpoint with a non-existent user ID", func(t *testing.T) {
		mockBookStore, ts, router := setupTestServer()
		defer ts.Close()

		mockBookStore.On("DeleteByID", mock.Anything, mock.Anything).Return(int(0), sql.ErrNoRows)

		req := httptest.NewRequest(http.MethodDelete, ts.URL+"/api/v1/books/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error": "No book found with ID 1"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should return succssefully status and body when call endpoint with valid body", func(t *testing.T) {
		mockBookStore, ts, router := setupTestServer()
		defer ts.Close()

		mockBookStore.On("DeleteByID", mock.Anything, mock.Anything).Return(int(1), nil)

		req := httptest.NewRequest(http.MethodDelete, ts.URL+"/api/v1/books/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"id":1}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})
}
