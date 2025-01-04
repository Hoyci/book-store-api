package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/cmd/api"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/service/user"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Create(ctx context.Context, user types.CreateUserDatabasePayload) (*types.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*types.User), args.Error(1)
}

func (m *MockUserStore) GetByID(ctx context.Context, id int) (*types.UserResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.UserResponse), args.Error(1)
}

func (m *MockUserStore) UpdateByID(ctx context.Context, id int, newUser types.UpdateUserPayload) (*types.UserResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.UserResponse), args.Error(1)
}

func (m *MockUserStore) DeleteByID(ctx context.Context, id int) (int, error) {
	args := m.Called(ctx, id)
	if val, ok := args.Get(0).(int); ok {
		return val, args.Error(1)
	}
	return 0, args.Error(1)
}

func TestHandleCreateUser(t *testing.T) {
	setupTestServer := func() (*MockUserStore, *httptest.Server, *mux.Router, config.Config) {
		mockUserStore := new(MockUserStore)
		mockUserHandler := user.NewUserHandler(mockUserStore)
		apiServer := api.NewApiServer(":8080", nil)
		router := apiServer.SetupRouter(nil, nil, mockUserHandler)
		ts := httptest.NewServer(router)
		return mockUserStore, ts, router, apiServer.Config
	}

	t.Run("it should throw an error when body is not a valid JSON", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		invalidBody := bytes.NewReader([]byte("INVALID JSON"))
		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/user", invalidBody)
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

		expectedResponse := `{"error":"body is not a valid json"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when body is a valid JSON but missing key", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		payload := types.CreateUserRequestPayload{
			// Username:        "JohnDoe",
			// Email:           "johndoe@email.com",
			// Password:        "123mudar",
			// ConfirmPassword: "123mudar",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/user", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":["Field 'Username' is invalid: required", "Field 'Email' is invalid: required", "Field 'Password' is invalid: required", "Field 'ConfirmPassword' is invalid: required"]}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when body does not contain a valid email", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		payload := types.CreateUserRequestPayload{
			Username:        "JohnDoe",
			Email:           "johndoe",
			Password:        "123mudar",
			ConfirmPassword: "123mudar",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/user", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":["Field 'Email' is invalid: email"]}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when password or confirmPassword is smaller than 8 chars", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		payload := types.CreateUserRequestPayload{
			Username:        "JohnDoe",
			Email:           "johndoe@email.com",
			Password:        "12345",
			ConfirmPassword: "12345",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/user", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":["Field 'Password' is invalid: min", "Field 'ConfirmPassword' is invalid: min"]}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw a database insert error", func(t *testing.T) {
		mockUserStore, ts, router, _ := setupTestServer()
		defer ts.Close()

		mockUserStore.On("Create", mock.Anything, mock.Anything).Return((*types.User)(nil), fmt.Errorf("failed to insert entity 'user': database error"))

		payload := types.CreateUserRequestPayload{
			Username:        "JohnDoe",
			Email:           "johndoe@email.com",
			Password:        "123mudar",
			ConfirmPassword: "123mudar",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/user", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

		responseBody, _ := io.ReadAll(res.Body)
		expected := `{"error":"failed to insert entity 'user': database error"}`
		assert.JSONEq(t, expected, string(responseBody))
	})

	t.Run("it should successfully create a user and return a valida JWT Token", func(t *testing.T) {
		mockUserStore, ts, router, config := setupTestServer()
		defer ts.Close()

		mockUserStore.On("Create", mock.Anything, mock.Anything).Return(
			&types.User{
				ID:        1,
				Username:  "JohnDoe",
				Email:     "johndoe@email.com",
				CreatedAt: time.Date(0001, 01, 01, 0, 0, 0, 0, time.UTC),
				UpdatedAt: nil,
				DeletedAt: nil,
			},
			nil,
		)

		payload := types.CreateUserRequestPayload{
			Username:        "JohnDoe",
			Email:           "johndoe@email.com",
			Password:        "123mudar",
			ConfirmPassword: "123mudar",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/user", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusCreated, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var responseMap map[string]interface{}
		err = json.Unmarshal(responseBody, &responseMap)
		if err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		tokenString, ok := responseMap["token"].(string)
		if !ok {
			t.Fatalf("Token not found or not a string")
		}

		claims, err := utils.VerifyJWT(tokenString, config.JWTSecret)
		assert.NoError(t, err, "Failed to verify JWT token")

		t.Log(claims)

		assert.Equal(t, "johndoe@email.com", claims.Email, "Email claim mismatch")
		assert.Equal(t, "JohnDoe", claims.Username, "Username claim mismatch")
		assert.Equal(t, 1, claims.UserID, "UserID claim mismatch")
	})
}

func TestHandleGetUserById(t *testing.T) {
	setupTestServer := func() (*MockUserStore, *httptest.Server, *mux.Router, config.Config) {
		mockUserStore := new(MockUserStore)
		mockUserHandler := user.NewUserHandler(mockUserStore)
		apiServer := api.NewApiServer(":8080", nil)
		router := apiServer.SetupRouter(nil, nil, mockUserHandler)
		ts := httptest.NewServer(router)
		return mockUserStore, ts, router, apiServer.Config
	}

	t.Run("it should throw an error when call endpoint without user ID", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/user/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("it should throw an error when call endpoint with wrong user ID", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		req := httptest.NewRequest(http.MethodGet, ts.URL+"/api/v1/user/johndoe", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":"user ID must be a positive integer"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should return succssefully status and body when call endpoint with valid body", func(t *testing.T) {
		mockUserStore, ts, router, _ := setupTestServer()
		defer ts.Close()

		mockUserStore.On("GetByID", mock.Anything, 1).Return(&types.UserResponse{
			ID:        1,
			Username:  "johndoe",
			Email:     "johndoe@email.com",
			CreatedAt: time.Date(0001, 01, 01, 0, 0, 0, 0, time.UTC),
			DeletedAt: nil,
			UpdatedAt: nil,
		}, nil)

		req := httptest.NewRequest(http.MethodGet, ts.URL+"/api/v1/user/1", nil)
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
			"username": "johndoe",
			"email": "johndoe@email.com",
			"createdAt": "0001-01-01T00:00:00Z",
			"deletedAt": null,
			"updatedAt": null
		}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})
}

func TestHandleUpdateUserById(t *testing.T) {
	setupTestServer := func() (*MockUserStore, *httptest.Server, *mux.Router, config.Config) {
		mockUserStore := new(MockUserStore)
		mockUserHandler := user.NewUserHandler(mockUserStore)
		apiServer := api.NewApiServer(":8080", nil)
		router := apiServer.SetupRouter(nil, nil, mockUserHandler)
		ts := httptest.NewServer(router)
		return mockUserStore, ts, router, apiServer.Config
	}

	t.Run("it should throw an error when call endpoint without user ID", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/user/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("it should throw an error when call endpoint with wrong user ID", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		req := httptest.NewRequest(http.MethodGet, ts.URL+"/api/v1/user/johndoe", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":"user ID must be a positive integer"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when no fields are provided for update", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		emptyPayload := `{}`
		req := httptest.NewRequest(http.MethodPut, ts.URL+"/api/v1/user/1", bytes.NewBufferString(emptyPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		fmt.Println(res)

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error": ["Field validation for 'Username' failed on the 'required' tag", "Field validation for 'Email' failed on the 'required' tag"]}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when body is invalid", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		invalidPayload := `{"username": "", "email": ""}`
		req := httptest.NewRequest(http.MethodPut, ts.URL+"/api/v1/user/1", bytes.NewBufferString(invalidPayload))
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

		expectedResponse := `{"error":["Field validation for 'Username' failed on the 'min' tag", "Field validation for 'Email' failed on the 'email' tag"]}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should return successfully status and body when the user is updated", func(t *testing.T) {
		mockUserStore, ts, router, _ := setupTestServer()
		defer ts.Close()

		mockUserStore.On("UpdateByID", mock.Anything, 1, mock.Anything).Return(&types.UserResponse{
			ID:        1,
			Username:  "johndoe - updated",
			Email:     "johndoeupdated@email.com",
			CreatedAt: time.Date(0001, 01, 01, 0, 0, 0, 0, time.UTC),
			DeletedAt: nil,
			UpdatedAt: utils.TimePtr(time.Date(0001, 01, 01, 0, 0, 0, 0, time.UTC)),
		}, nil)

		validPayload := `{
			"username": "johndoe - updated",
			"email": "johndoeupdated@email.com"
		}`
		req := httptest.NewRequest(http.MethodPut, ts.URL+"/api/v1/user/1", bytes.NewBufferString(validPayload))
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
			"username":  "johndoe - updated",
			"email": "johndoeupdated@email.com",
			"createdAt": "0001-01-01T00:00:00Z",
			"updatedAt": "0001-01-01T00:00:00Z",
			"deletedAt": null 
		}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})
}

func TestHandleDeleteBookByID(t *testing.T) {
	setupTestServer := func() (*MockUserStore, *httptest.Server, *mux.Router, config.Config) {
		mockUserStore := new(MockUserStore)
		mockUserHandler := user.NewUserHandler(mockUserStore)
		apiServer := api.NewApiServer(":8080", nil)
		router := apiServer.SetupRouter(nil, nil, mockUserHandler)
		ts := httptest.NewServer(router)
		return mockUserStore, ts, router, apiServer.Config
	}

	t.Run("it should throw an error when call endpoint without book ID", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		req := httptest.NewRequest(http.MethodDelete, ts.URL+"/api/v1/user", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("it should throw an error when call endpoint with wrong ID", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		req := httptest.NewRequest(http.MethodDelete, ts.URL+"/api/v1/user/anything", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":"book ID must be a positive integer"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should return succssefully status and body when call endpoint with valid body", func(t *testing.T) {
		mockBookStore, ts, router, _ := setupTestServer()
		defer ts.Close()

		mockBookStore.On("DeleteByID", mock.Anything, mock.Anything).Return(int(1), nil)

		req := httptest.NewRequest(http.MethodDelete, ts.URL+"/api/v1/user/1", nil)
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
