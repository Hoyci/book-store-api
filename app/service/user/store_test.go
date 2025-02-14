package user

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := NewUserStore(db)
	user := types.CreateUserDatabasePayload{
		Username:     "JohnDoe",
		Email:        "johndoe@email.com",
		PasswordHash: "2345678",
	}

	t.Run("database connection error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id, username, email, created_at, updated_at, deleted_at")).
			WithArgs(user.Username, user.Email, user.PasswordHash).
			WillReturnError(sql.ErrConnDone)

		id, err := store.Create(context.Background(), user)

		assert.Error(t, err)
		assert.Zero(t, id)
		assert.Equal(t, err, sql.ErrConnDone)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully create user", func(t *testing.T) {
		mockedDate := time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC)
		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id, username, email, created_at, updated_at, deleted_at")).
			WithArgs(user.Username, user.Email, user.PasswordHash).
			WillReturnRows(
				sqlmock.NewRows([]string{
					"id", "username", "email", "created_at", "updated_at", "deleted_at",
				}).AddRow(
					1,
					"JohnDoe",
					"johndoe@email.com",
					mockedDate,
					nil,
					nil,
				))

		newUser, err := store.Create(context.Background(), user)

		assert.NoError(t, err)
		assert.Equal(t, 1, newUser.ID)
		assert.Equal(t, user.Username, newUser.Username)
		assert.Equal(t, user.Email, newUser.Email)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})
}

func TestGetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := NewUserStore(db)

	ctx := utils.SetClaimsToContext(context.Background(), &types.CustomClaims{
		ID:               "ID-CRAZY",
		UserID:           1,
		Username:         "johndoe",
		Email:            "johndoe@email.com",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24))},
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		ctx = utils.SetClaimsToContext(ctx, &types.CustomClaims{
			ID:               "ID-CRAZY",
			UserID:           1,
			Username:         "johndoe",
			Email:            "johndoe@email.com",
			RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24))},
		})

		user, err := store.GetByID(ctx, 1)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, user)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database did not find any row", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, created_at, updated_at, deleted_at  FROM users WHERE id = $1 AND deleted_at IS null")).
			WithArgs(1).
			WillReturnError(sql.ErrNoRows)

		user, err := store.GetByID(ctx, 1)

		assert.Nil(t, user)
		assert.Error(t, err)
		assert.Equal(t, err, sql.ErrNoRows)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database connection error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, created_at, updated_at, deleted_at  FROM users WHERE id = $1 AND deleted_at IS null")).
			WithArgs(1).
			WillReturnError(sql.ErrConnDone)

		user, err := store.GetByID(ctx, 1)

		assert.Error(t, err)
		assert.Zero(t, user)
		assert.Equal(t, err, sql.ErrConnDone)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully get user by ID", func(t *testing.T) {
		expectedCreatedAt := time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, created_at, updated_at, deleted_at  FROM users WHERE id = $1 AND deleted_at IS null")).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "created_at", "updated_at", "deleted_at"}).
				AddRow(1, "johndoe", "johndoe@email.com", expectedCreatedAt, nil, nil))

		user, err := store.GetByID(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "johndoe", user.Username)
		assert.Equal(t, "johndoe@email.com", user.Email)
		assert.Equal(t, expectedCreatedAt, user.CreatedAt)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})
}

func TestGetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := NewUserStore(db)

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		user, err := store.GetByEmail(ctx, "johndoe@email.com")

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, user)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database did not find any row", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, created_at, updated_at, deleted_at FROM users WHERE email = $1 AND deleted_at IS null")).
			WithArgs("johndoe@email.com").
			WillReturnError(sql.ErrNoRows)

		user, err := store.GetByEmail(context.Background(), "johndoe@email.com")

		assert.Nil(t, user)
		assert.Error(t, err)
		assert.Equal(t, err, sql.ErrNoRows)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database connection error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, username, email, created_at, updated_at, deleted_at FROM users WHERE email = $1 AND deleted_at IS null")).
			WithArgs("johndoe@email.com").
			WillReturnError(sql.ErrConnDone)

		user, err := store.GetByEmail(context.Background(), "johndoe@email.com")

		assert.Error(t, err)
		assert.Zero(t, user)
		assert.Equal(t, err, sql.ErrConnDone)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully get user by ID", func(t *testing.T) {
		expectedCreatedAt := time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC)

		mock.ExpectQuery("SELECT id, username, email, created_at, updated_at, deleted_at FROM users WHERE email = \\$1 AND deleted_at IS null").
			WithArgs("johndoe@email.com").
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "created_at", "updated_at", "deleted_at"}).
				AddRow(1, "johndoe", "johndoe@email.com", expectedCreatedAt, nil, nil))

		expectedID := 1

		user, err := store.GetByEmail(context.Background(), "johndoe@email.com")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedID, user.ID)
		assert.Equal(t, "johndoe", user.Username)
		assert.Equal(t, "johndoe@email.com", user.Email)
		assert.Equal(t, expectedCreatedAt, user.CreatedAt)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})
}

func TestUpdateByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := NewUserStore(db)

	ctx := utils.SetClaimsToContext(context.Background(), &types.CustomClaims{
		ID:               "ID-CRAZY",
		UserID:           1,
		Username:         "johndoe",
		Email:            "johndoe@email.com",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24))},
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		ctx = utils.SetClaimsToContext(ctx, &types.CustomClaims{
			ID:               "ID-CRAZY",
			UserID:           1,
			Username:         "johndoe",
			Email:            "johndoe@email.com",
			RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24))},
		})

		updatedUser, err := store.UpdateByID(ctx, 1, types.UpdateUserPayload{
			Username: "Updated Username",
			Email:    "Updated Email",
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, updatedUser)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("database did not find any row", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`
			UPDATE users SET 
			username = $2, 
			email = $3,
			updated_at = $4
			WHERE id = $1
			RETURNING 
				id, 
				username, 
				email, 
				created_at, 
				deleted_at,
				updated_at;
			`)).
			WithArgs(
				1,
				"Updated Username",
				"Updated Email",
				time.Now(),
			).
			WillReturnError(sql.ErrNoRows)

		id, err := store.UpdateByID(ctx, 1, types.UpdateUserPayload{
			Username: "Updated Username",
			Email:    "Updated Email",
		})

		assert.Nil(t, id)
		assert.Error(t, err)
		assert.Equal(t, err, sql.ErrNoRows)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database connection error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`
			UPDATE users SET 
			username = $2, 
			email = $3,
			updated_at = $4
			WHERE id = $1
			RETURNING 
				id, 
				username, 
				email, 
				created_at, 
				deleted_at,
				updated_at;
			`)).
			WithArgs(
				1,
				"Updated Username",
				"Updated Email",
				time.Now(),
			).
			WillReturnError(sql.ErrConnDone)

		id, err := store.UpdateByID(ctx, 1, types.UpdateUserPayload{
			Username: "Updated Username",
			Email:    "Updated Email",
		})

		assert.Error(t, err)
		assert.Zero(t, id)
		assert.Equal(t, err, sql.ErrConnDone)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully update user", func(t *testing.T) {
		mockedDate := time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC)
		mock.ExpectQuery(regexp.QuoteMeta(`
			UPDATE users SET 
			username = $2, 
			email = $3,
			updated_at = $4
			WHERE id = $1
			RETURNING 
				id, 
				username, 
				email, 
				created_at, 
				deleted_at,
				updated_at;
			`)).
			WithArgs(
				1,
				"Updated Username",
				"Updated Email",
				time.Now(),
			).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "username", "email", "created_at", "deleted_at", "updated_at",
			}).AddRow(
				1,
				"Updated Username",
				"Updated Email",
				mockedDate,
				nil,
				&mockedDate,
			))

		updatedUser, err := store.UpdateByID(ctx, 1, types.UpdateUserPayload{
			Username: "Updated Username",
			Email:    "Updated Email",
		})

		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, 1, updatedUser.ID)
		assert.Equal(t, "Updated Username", updatedUser.Username)
		assert.Equal(t, "Updated Email", updatedUser.Email)
		assert.Equal(t, mockedDate, updatedUser.CreatedAt)
		assert.Nil(t, updatedUser.DeletedAt)
		assert.Equal(t, &mockedDate, updatedUser.UpdatedAt)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestDeleteByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := NewUserStore(db)

	ctx := utils.SetClaimsToContext(context.Background(), &types.CustomClaims{
		ID:               "ID-CRAZY",
		UserID:           1,
		Username:         "johndoe",
		Email:            "johndoe@email.com",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24))},
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		ctx = utils.SetClaimsToContext(ctx, &types.CustomClaims{
			ID:               "ID-CRAZY",
			UserID:           1,
			Username:         "johndoe",
			Email:            "johndoe@email.com",
			RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24))},
		})

		err := store.DeleteByID(ctx, 1)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("database did not find any row", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET deleted_at = $2 WHERE id = $1")).
			WithArgs(1, sqlmock.AnyArg()).
			WillReturnError(ErrUserNotFound)

		err := store.DeleteByID(ctx, 1)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "user not found")
		assert.ErrorIs(t, err, ErrUserNotFound)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database connection error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET deleted_at = $2 WHERE id = $1")).
			WithArgs(1, sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		err := store.DeleteByID(ctx, 1)

		assert.Error(t, err)
		assert.Equal(t, err, sql.ErrConnDone)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully delete user by ID", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET deleted_at = $2 WHERE id = $1")).
			WithArgs(1, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.DeleteByID(ctx, 1)

		assert.NoError(t, err)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})
}
