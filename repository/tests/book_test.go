package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hoyci/book-store-api/repository"
	"github.com/hoyci/book-store-api/types"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreateBook(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := repository.NewBookStore(db)
	book := types.Book{
		Name:          "Go Programming",
		Description:   "A book about Go programming",
		Author:        "John Doe",
		Genres:        []string{"Programming"},
		ReleaseYear:   2024,
		NumberOfPages: 300,
		ImageUrl:      "http://example.com/go.jpg",
	}

	mock.ExpectExec("INSERT INTO books").
		WithArgs(book.Name, book.Description, book.Author, pq.Array(book.Genres), book.ReleaseYear, book.NumberOfPages, book.ImageUrl).
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := store.Create(context.Background(), book)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGetBookByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := repository.NewBookStore(db)
	expectedCreatedAt := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT id, name, description, author, genres").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "author", "genres", "release_year", "number_of_pages", "image_url", "created_at"}).
			AddRow(1, "Go Programming", "A book about Go programming", "John Doe", pq.Array([]string{"Programming"}), 2024, 300, "http://example.com/go.jpg", expectedCreatedAt))

	book, err := store.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}
	if book == nil {
		t.Fatal("Expected book to be not nil, but got nil")
	}

	expectedID := int64(1)

	assert.Equal(t, expectedID, book.ID)
	assert.Equal(t, "Go Programming", book.Name)
	assert.Equal(t, expectedCreatedAt, book.CreatedAt)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
