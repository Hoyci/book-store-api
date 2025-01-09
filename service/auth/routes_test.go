package auth_test

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
	"github.com/hoyci/book-store-api/service/auth"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Create(ctx context.Context, user types.CreateUserDatabasePayload) (*types.UserResponse, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*types.UserResponse), args.Error(1)
}

func (m *MockUserStore) GetByID(ctx context.Context, id int) (*types.UserResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*types.UserResponse), args.Error(1)
}

func (m *MockUserStore) GetByEmail(ctx context.Context, email string) (*types.UserResponse, error) {
	args := m.Called(ctx, email)
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

func TestHandleUserLogin(t *testing.T) {
	setupTestServer := func() (*MockUserStore, *httptest.Server, *mux.Router, config.Config) {
		mockUserStore := new(MockUserStore)
		mockAuthHandler := auth.NewAuthHandler(mockUserStore)
		apiServer := api.NewApiServer(":8080", nil)
		router := apiServer.SetupRouter(nil, nil, nil, mockAuthHandler)
		ts := httptest.NewServer(router)
		return mockUserStore, ts, router, apiServer.Config
	}

	t.Run("it should throw an error when body is not a valid JSON", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		invalidBody := bytes.NewReader([]byte("INVALID JSON"))
		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/auth", invalidBody)
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
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		payload := types.UserLoginPayload{}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/auth", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":["Field 'Email' is invalid: required", "Field 'Password' is invalid: required"]}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when body does not contain a valid email", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		payload := types.UserLoginPayload{
			Email:    "johndoe",
			Password: "123mudar",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/auth", bytes.NewBuffer(marshalled))
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

		payload := types.UserLoginPayload{
			Email:    "johndoe@email.com",
			Password: "12345",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/auth", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":["Field 'Password' is invalid: min"]}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw a database find error", func(t *testing.T) {
		mockUserStore, ts, router, _ := setupTestServer()
		defer ts.Close()

		mockUserStore.On("GetByEmail", mock.Anything, mock.Anything).Return((*types.UserResponse)(nil), fmt.Errorf("no row found with email: 'johndoe@email.com'"))

		payload := types.UserLoginPayload{
			Email:    "johndoe@email.com",
			Password: "123mudar",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/auth", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

		responseBody, _ := io.ReadAll(res.Body)
		expected := `{"error":"An unexpected error occurred"}`
		assert.JSONEq(t, expected, string(responseBody))
	})

	t.Run("it should successfully authenticate a user", func(t *testing.T) {
		mockUserStore, ts, router, config := setupTestServer()
		defer ts.Close()

		mockUserStore.On("GetByEmail", mock.Anything, mock.Anything).Return(
			&types.UserResponse{
				ID:        1,
				Username:  "JohnDoe",
				Email:     "johndoe@email.com",
				CreatedAt: time.Date(0001, 01, 01, 0, 0, 0, 0, time.UTC),
				UpdatedAt: nil,
				DeletedAt: nil,
			},
			nil,
		)

		payload := types.UserLoginPayload{
			Email:    "johndoe@email.com",
			Password: "123mudar",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/auth", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var responseMap map[string]interface{}
		err = json.Unmarshal(responseBody, &responseMap)
		if err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		access_token, ok := responseMap["access_token"].(string)
		if !ok {
			t.Fatalf("access_token not found or not a string")
		}

		refresh_token, ok := responseMap["refresh_token"].(string)
		if !ok {
			t.Fatalf("refresh_token not found or not a string")
		}

		access_token_claims, err := utils.VerifyJWT(access_token, config.JWTSecret)
		assert.NoError(t, err, "Failed to verify JWT token")

		assert.Equal(t, "johndoe@email.com", access_token_claims.Email, "Email claim mismatch")
		assert.Equal(t, "JohnDoe", access_token_claims.Username, "Username claim mismatch")
		assert.Equal(t, 1, access_token_claims.UserID, "UserID claim mismatch")

		refresh_token_claims, err := utils.VerifyJWT(refresh_token, config.JWTSecret)
		assert.NoError(t, err, "Failed to verify JWT token")

		assert.Equal(t, 1, refresh_token_claims.UserID, "UserID claim mismatch")
	})
}

func TestHandleRefreshToken(t *testing.T) {
	setupTestServer := func() (*MockUserStore, *httptest.Server, *mux.Router, config.Config) {
		mockUserStore := new(MockUserStore)
		mockAuthHandler := auth.NewAuthHandler(mockUserStore)
		apiServer := api.NewApiServer(":8080", nil)
		router := apiServer.SetupRouter(nil, nil, nil, mockAuthHandler)
		ts := httptest.NewServer(router)
		return mockUserStore, ts, router, apiServer.Config
	}

	t.Run("it should throw an error when body is not a valid JSON", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		invalidBody := bytes.NewReader([]byte("INVALID JSON"))
		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/auth/refresh", invalidBody)
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
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		payload := types.UserLoginPayload{}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/auth/refresh", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":["Field 'AccessToken' is invalid: required"]}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should throw an error when body does not contain a valid token", func(t *testing.T) {
		_, ts, router, _ := setupTestServer()
		defer ts.Close()

		payload := types.RefreshTokenPayload{
			AccessToken: "123mudar",
		}
		marshalled, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/auth/refresh", bytes.NewBuffer(marshalled))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		expectedResponse := `{"error":"Refresh token is invalid or has been expired"}`
		assert.JSONEq(t, expectedResponse, string(responseBody))
	})

	t.Run("it should successfully refresh user token a user", func(t *testing.T) {
		mockUserStore, ts, router, config := setupTestServer()
		defer ts.Close()

		mockUserStore.On("GetByEmail", mock.Anything, mock.Anything).Return(
			&types.UserResponse{
				ID:        1,
				Username:  "JohnDoe",
				Email:     "johndoe@email.com",
				CreatedAt: time.Date(0001, 01, 01, 0, 0, 0, 0, time.UTC),
				UpdatedAt: nil,
				DeletedAt: nil,
			},
			nil,
		)

		userLoginPayload := types.UserLoginPayload{
			Email:    "johndoe@email.com",
			Password: "123mudar",
		}
		userLoginMarshalled, _ := json.Marshal(userLoginPayload)

		reqUserLogin := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/auth", bytes.NewBuffer(userLoginMarshalled))
		wUserLogin := httptest.NewRecorder()

		router.ServeHTTP(wUserLogin, reqUserLogin)

		resUserLogin := wUserLogin.Result()
		defer resUserLogin.Body.Close()

		assert.Equal(t, http.StatusOK, resUserLogin.StatusCode)

		responseUserLoginBody, err := io.ReadAll(resUserLogin.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var responseUserLoginMap types.UserLoginResponse
		err = json.Unmarshal(responseUserLoginBody, &responseUserLoginMap)
		if err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		assert.NotEmpty(t, responseUserLoginMap.AccessToken, "Access token should not be empty")
		assert.NotEmpty(t, responseUserLoginMap.RefreshToken, "Refresh token should not be empty")

		userRefreshTokenPayload := types.RefreshTokenPayload{
			AccessToken: responseUserLoginMap.AccessToken,
		}
		userRefreshTokenMarshalled, _ := json.Marshal(userRefreshTokenPayload)

		reqRefreshToken := httptest.NewRequest(http.MethodPost, ts.URL+"/api/v1/auth/refresh", bytes.NewBuffer(userRefreshTokenMarshalled))
		wRefreshToken := httptest.NewRecorder()

		router.ServeHTTP(wRefreshToken, reqRefreshToken)

		resRefreshToken := wRefreshToken.Result()
		defer resRefreshToken.Body.Close()

		assert.Equal(t, http.StatusOK, resRefreshToken.StatusCode)

		responseRefreshTokenBody, err := io.ReadAll(resRefreshToken.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var responseRefreshTokenMap map[string]interface{}
		err = json.Unmarshal(responseRefreshTokenBody, &responseRefreshTokenMap)
		if err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		access_token, ok := responseRefreshTokenMap["access_token"].(string)
		if !ok {
			t.Fatalf("access_token not found or not a string")
		}

		refresh_token, ok := responseRefreshTokenMap["refresh_token"].(string)
		if !ok {
			t.Fatalf("refresh_token not found or not a string")
		}

		access_token_claims, err := utils.VerifyJWT(access_token, config.JWTSecret)
		assert.NoError(t, err, "Failed to verify JWT token")

		assert.Equal(t, "johndoe@email.com", access_token_claims.Email, "Email claim mismatch")
		assert.Equal(t, "JohnDoe", access_token_claims.Username, "Username claim mismatch")
		assert.Equal(t, 1, access_token_claims.UserID, "UserID claim mismatch")

		refresh_token_claims, err := utils.VerifyJWT(refresh_token, config.JWTSecret)
		assert.NoError(t, err, "Failed to verify JWT token")
		assert.Equal(t, 1, refresh_token_claims.UserID, "UserID claim mismatch")
	})
}