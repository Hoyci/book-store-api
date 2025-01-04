package user

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

	t.Run("database unexpected error", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO users").
			WithArgs(user.Username, user.Email, user.PasswordHash).
			WillReturnError(fmt.Errorf("database connection error"))

		id, err := store.Create(context.Background(), user)

		assert.Error(t, err)
		assert.Zero(t, id)
		assert.Contains(t, err.Error(), "unexpected error")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully create user", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO users").
			WithArgs(user.Username, user.Email, user.PasswordHash).
			WillReturnRows(
				sqlmock.NewRows([]string{
					"id", "username", "email", "created_at", "updated_at", "deleted_at",
				}).AddRow(
					1,
					"JohnDoe",
					"johndoe@email.com",
					time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
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
	expectedCreatedAt := time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("database did not find any row", func(t *testing.T) {
		mock.ExpectQuery("SELECT *").
			WithArgs(1).
			WillReturnError(sql.ErrNoRows)

		user, err := store.GetByID(context.Background(), 1)

		assert.Nil(t, user)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "no row found with id: '1'")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database unexpected error", func(t *testing.T) {
		mock.ExpectQuery("SELECT *").
			WithArgs(1).
			WillReturnError(fmt.Errorf("database connection error"))

		user, err := store.GetByID(context.Background(), 1)

		assert.Error(t, err)
		assert.Zero(t, user)
		assert.Contains(t, err.Error(), "unexpected error getting user with id: '1'")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully get user by ID", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\$1 AND deleted_at IS null").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "created_at", "updated_at", "deleted_at"}).
				AddRow(1, "johndoe", "johndoe@email.com", expectedCreatedAt, nil, nil))

		expectedID := 1

		user, err := store.GetByID(context.Background(), expectedID)

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

func TestUpdateBook(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := NewUserStore(db)

	t.Run("no fields to update", func(t *testing.T) {
		emptyUpdates := types.UpdateUserPayload{}
		result, err := store.UpdateByID(context.Background(), 1, emptyUpdates)

		assert.Error(t, err, "no fields to update for user with ID %d", 1)
		assert.Nil(t, result)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database did not find any row", func(t *testing.T) {
		updates := types.UpdateUserPayload{
			Username: utils.StringPtr("Updated Username"),
			Email:    utils.StringPtr("Updated Email"),
		}

		mock.ExpectQuery("UPDATE users SET").
			WithArgs(
				"Updated Username",
				"Updated Email",
				999,
			).
			WillReturnError(sql.ErrNoRows)

		id, err := store.UpdateByID(context.Background(), 999, updates)

		assert.Nil(t, id)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no row found with id: '999'")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database unexpected error", func(t *testing.T) {
		updates := types.UpdateUserPayload{
			Username: utils.StringPtr("Updated Username"),
			Email:    utils.StringPtr("Updated Email"),
		}

		mock.ExpectQuery("UPDATE users SET").
			WithArgs(
				"Updated Username",
				"Updated Email",
				1,
			).
			WillReturnError(fmt.Errorf("database connection error"))

		id, err := store.UpdateByID(context.Background(), 1, updates)

		assert.Error(t, err)
		assert.Zero(t, id)
		assert.Contains(t, err.Error(), "unexpected error updating user with id: '1'")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully update user", func(t *testing.T) {
		updates := types.UpdateUserPayload{
			Username: utils.StringPtr("Updated Username"),
			Email:    utils.StringPtr("Updated Email"),
		}
		mock.ExpectQuery("UPDATE users SET").
			WithArgs(
				"Updated Username",
				"Updated Email",
				1,
			).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "username", "email", "created_at", "updated_at", "deleted_at",
			}).AddRow(
				1,
				"Updated Username",
				"Updated Email",
				time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC),
			))

		result, err := store.UpdateByID(context.Background(), 1, updates)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "Updated Username", result.Username)
		assert.Equal(t, "Updated Email", result.Email)

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

	t.Run("database did not find any row", func(t *testing.T) {
		mock.ExpectQuery("UPDATE users SET deleted_at").
			WithArgs(1, sqlmock.AnyArg()).
			WillReturnError(sql.ErrNoRows)

		id, err := store.DeleteByID(context.Background(), 1)

		assert.Equal(t, id, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no row found with id: '1'")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database unexpected error", func(t *testing.T) {
		mock.ExpectQuery("UPDATE users SET deleted_at").
			WithArgs(1, sqlmock.AnyArg()).
			WillReturnError(fmt.Errorf("database connection error"))

		id, err := store.DeleteByID(context.Background(), 1)

		assert.Error(t, err)
		assert.Zero(t, id)
		assert.Contains(t, err.Error(), "unexpected error deleting user with id: '1'")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully delete user by ID", func(t *testing.T) {
		mock.ExpectQuery("UPDATE users SET deleted_at").
			WithArgs(1, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		id, err := store.DeleteByID(context.Background(), 1)
		expectedID := 1

		assert.NoError(t, err)
		assert.Equal(t, expectedID, id)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})
}
