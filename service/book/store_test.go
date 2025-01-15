package book

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreateBook(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := NewBookStore(db)
	book := types.CreateBookPayload{
		Name:          "Go Programming",
		Description:   "A book about Go programming",
		Author:        "John Doe",
		Genres:        []string{"Programming"},
		ReleaseYear:   2024,
		NumberOfPages: 300,
		ImageUrl:      "http://example.com/go.jpg",
	}

	t.Run("database unexpected error", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO books").
			WithArgs(book.Name, book.Description, book.Author, pq.Array(book.Genres), book.ReleaseYear, book.NumberOfPages, book.ImageUrl).
			WillReturnError(fmt.Errorf("database connection error"))

		id, err := store.Create(context.Background(), book)

		assert.Error(t, err)
		assert.Zero(t, id)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully create book", func(t *testing.T) {
		mock.ExpectQuery("INSERT INTO books").
			WithArgs(book.Name, book.Description, book.Author, pq.Array(book.Genres), book.ReleaseYear, book.NumberOfPages, book.ImageUrl).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		id, err := store.Create(context.Background(), book)

		assert.NoError(t, err)
		assert.Equal(t, int(1), id)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})
}

func TestGetBookByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := NewBookStore(db)
	expectedCreatedAt := time.Date(0001, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("database did not find any row", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM books WHERE id = \\$1 AND deleted_at IS null").
			WithArgs(1).
			WillReturnError(sql.ErrNoRows)

		book, err := store.GetByID(context.Background(), 1)

		assert.Nil(t, book)
		assert.Error(t, err)
		assert.Equal(t, err, sql.ErrNoRows)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database unexpected error", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM books WHERE id = \\$1 AND deleted_at IS null").
			WithArgs(1).
			WillReturnError(fmt.Errorf("database connection error"))

		book, err := store.GetByID(context.Background(), 1)

		assert.Error(t, err)
		assert.Zero(t, book)
		assert.NotEqual(t, err, sql.ErrNoRows)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully get book by ID", func(t *testing.T) {
		mock.ExpectQuery("SELECT \\* FROM books WHERE id = \\$1 AND deleted_at IS null").
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "author", "genres", "release_year", "number_of_pages", "image_url", "created_at", "updated_at", "deleted_at"}).
				AddRow(1, "Go Programming", "A book about Go programming", "John Doe", pq.Array([]string{"Programming"}), 2024, 300, "http://example.com/go.jpg", expectedCreatedAt, nil, nil))

		expectedID := 1

		book, err := store.GetByID(context.Background(), expectedID)

		assert.NoError(t, err)
		assert.NotNil(t, book)
		assert.Equal(t, expectedID, book.ID)
		assert.Equal(t, "Go Programming", book.Name)
		assert.Equal(t, expectedCreatedAt, book.CreatedAt)

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

	store := NewBookStore(db)

	t.Run("no fields to update", func(t *testing.T) {
		emptyUpdates := types.UpdateBookPayload{}
		result, err := store.UpdateByID(context.Background(), 1, emptyUpdates)

		assert.Error(t, err, "no fields to update for book with ID %d", 1)
		assert.Nil(t, result)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database did not find any row", func(t *testing.T) {
		updates := types.UpdateBookPayload{
			Name:        utils.StringPtr("Updated Book Name"),
			Description: utils.StringPtr("Updated Description"),
			Genres:      &[]string{"Genre1", "Genre2"},
			ReleaseYear: utils.IntPtr(2025),
		}

		mock.ExpectQuery("UPDATE books SET").
			WithArgs(
				"Updated Book Name",
				"Updated Description",
				pq.Array([]string{"Genre1", "Genre2"}),
				2025,
				1,
			).
			WillReturnError(sql.ErrNoRows)

		id, err := store.UpdateByID(context.Background(), 1, updates)

		assert.Nil(t, id)
		assert.Error(t, err)
		assert.Equal(t, err, sql.ErrNoRows)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database unexpected error", func(t *testing.T) {
		updates := types.UpdateBookPayload{
			Name:        utils.StringPtr("Updated Book Name"),
			Description: utils.StringPtr("Updated Description"),
			Genres:      &[]string{"Genre1", "Genre2"},
			ReleaseYear: utils.IntPtr(2025),
		}

		mock.ExpectQuery("UPDATE books SET").
			WithArgs(
				"Updated Book Name",
				"Updated Description",
				pq.Array([]string{"Genre1", "Genre2"}),
				2025,
				1,
			).
			WillReturnError(fmt.Errorf("database connection error"))

		id, err := store.UpdateByID(context.Background(), 1, updates)

		assert.Error(t, err)
		assert.Zero(t, id)
		assert.NotEqual(t, err, sql.ErrNoRows)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully update book", func(t *testing.T) {
		updates := types.UpdateBookPayload{
			Name:        utils.StringPtr("Updated Book Name"),
			Description: utils.StringPtr("Updated Description"),
			Genres:      &[]string{"Genre1", "Genre2"},
			ReleaseYear: utils.IntPtr(2025),
		}

		mock.ExpectQuery("UPDATE books SET").
			WithArgs(
				"Updated Book Name",
				"Updated Description",
				pq.Array([]string{"Genre1", "Genre2"}),
				2025,
				1,
			).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "name", "description", "author", "genres", "release_year", "number_of_pages", "image_url", "created_at", "updated_at",
			}).AddRow(
				1,
				"Updated Book Name",
				"Updated Description",
				"Author Name",
				pq.Array([]string{"Genre1", "Genre2"}),
				2025,
				300,
				"http://example.com/image.jpg",
				time.Now(),
				time.Now(),
			))

		result, err := store.UpdateByID(context.Background(), 1, updates)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "Updated Book Name", result.Name)
		assert.Equal(t, "Updated Description", result.Description)
		assert.Equal(t, []string{"Genre1", "Genre2"}, result.Genres)

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

	store := NewBookStore(db)

	t.Run("database did not find any row", func(t *testing.T) {
		mock.ExpectQuery("UPDATE books SET deleted_at").
			WithArgs(1, sqlmock.AnyArg()).
			WillReturnError(sql.ErrNoRows)

		id, err := store.DeleteByID(context.Background(), 1)

		assert.Equal(t, id, 0)
		assert.Error(t, err)
		assert.Equal(t, err, sql.ErrNoRows)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database unexpected error", func(t *testing.T) {
		mock.ExpectQuery("UPDATE books SET deleted_at").
			WithArgs(1, sqlmock.AnyArg()).
			WillReturnError(fmt.Errorf("database connection error"))

		id, err := store.DeleteByID(context.Background(), 1)

		assert.Error(t, err)
		assert.Zero(t, id)
		assert.NotEqual(t, err, sql.ErrNoRows)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully delete book by ID", func(t *testing.T) {
		mock.ExpectQuery("UPDATE books SET deleted_at").
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
