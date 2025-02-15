package book

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
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

	ctx := utils.SetClaimsToContext(context.Background(), &types.CustomClaims{
		ID:               "ID-CRAZY",
		UserID:           1,
		Username:         "johndoe",
		Email:            "johndoe@email.com",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24))},
	})

	t.Run("missing userID in context", func(t *testing.T) {
		ctx := context.Background()

		id, err := store.Create(ctx, book)

		assert.Error(t, err)
		assert.Equal(t, "failed to retrieve userID from context", err.Error())
		assert.Zero(t, id)
	})

	t.Run("fail to begin transaction", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(fmt.Errorf("failed to begin transaction"))

		id, err := store.Create(ctx, book)

		assert.Error(t, err)
		assert.Equal(t, "failed to begin transaction", err.Error())
		assert.Zero(t, id)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database unexpected error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO books").
			WithArgs(book.Name, book.Description, book.Author, pq.Array(book.Genres), book.ReleaseYear, book.NumberOfPages, book.ImageUrl).
			WillReturnError(fmt.Errorf("database connection error"))
		mock.ExpectRollback()

		id, err := store.Create(ctx, book)

		assert.Error(t, err)
		assert.Zero(t, id)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("fail to insert into users_books", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO books").
			WithArgs(book.Name, book.Description, book.Author, pq.Array(book.Genres), book.ReleaseYear, book.NumberOfPages, book.ImageUrl).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec("INSERT INTO users_books").
			WithArgs(1, 1).
			WillReturnError(fmt.Errorf("failed to insert into users_books"))
		mock.ExpectRollback()

		id, err := store.Create(ctx, book)

		assert.Error(t, err)
		assert.Equal(t, "failed to insert into users_books", err.Error())
		assert.Zero(t, id)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		ctx = utils.SetClaimsToContext(ctx, &types.CustomClaims{
			ID:               "ID-CRAZY",
			UserID:           1,
			Username:         "johndoe",
			Email:            "johndoe@email.com",
			RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24))},
		})

		id, err := store.Create(ctx, book)

		assert.Error(t, err)
		assert.Zero(t, id)
		assert.True(t, errors.Is(err, context.Canceled))
	})

	t.Run("rollback on intermediate failure", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO books").
			WithArgs(book.Name, book.Description, book.Author, pq.Array(book.Genres), book.ReleaseYear, book.NumberOfPages, book.ImageUrl).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec("INSERT INTO users_books").
			WithArgs(1, 1).
			WillReturnError(fmt.Errorf("database error"))
		mock.ExpectRollback()

		id, err := store.Create(ctx, book)

		assert.Error(t, err)
		assert.Zero(t, id)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully create book", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO books").
			WithArgs(book.Name, book.Description, book.Author, pq.Array(book.Genres), book.ReleaseYear, book.NumberOfPages, book.ImageUrl).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec("INSERT INTO users_books").
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		id, err := store.Create(ctx, book)

		assert.NoError(t, err)
		assert.Equal(t, 1, id)

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
	expectedCreatedAt := time.Now()

	ctx := utils.SetClaimsToContext(context.Background(), &types.CustomClaims{
		ID:               "ID-CRAZY",
		UserID:           1,
		Username:         "johndoe",
		Email:            "johndoe@email.com",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24))},
	})

	t.Run("missing userID in context", func(t *testing.T) {
		ctx := context.Background()

		id, err := store.GetByID(ctx, 1)

		assert.Error(t, err)
		assert.Equal(t, "failed to retrieve userID from context", err.Error())
		assert.Zero(t, id)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		cancel()

		books, err := store.GetByID(ctx, 1)

		assert.Error(t, err)
		assert.Nil(t, books)
		assert.True(t, errors.Is(err, context.Canceled))

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database did not find any row", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
		b.* 
		FROM books b
		INNER JOIN users_books ub ON ub.book_id = b.id
		WHERE b.id = $1
		AND ub.user_id = $2
		AND b.deleted_at IS NULL;
		`)).
			WithArgs(1, 1).
			WillReturnError(sql.ErrNoRows)

		book, err := store.GetByID(ctx, 1)

		assert.Nil(t, book)
		assert.Error(t, err)
		assert.Equal(t, err, sql.ErrNoRows)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database connection error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT 
				b.* 
				FROM books b
				INNER JOIN users_books ub ON ub.book_id = b.id
				WHERE b.id = $1
				AND ub.user_id = $2
				AND b.deleted_at IS NULL;
			`)).
			WithArgs(1, 1).
			WillReturnError(sql.ErrConnDone)

		book, err := store.GetByID(ctx, 1)

		assert.Error(t, err)
		assert.Zero(t, book)
		assert.True(t, errors.Is(err, sql.ErrConnDone))

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully get book by ID", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT 
				b.* 
				FROM books b
				INNER JOIN users_books ub ON ub.book_id = b.id
				WHERE b.id = $1
				AND ub.user_id = $2
				AND b.deleted_at IS NULL;
			`)).
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "author", "genres", "release_year", "number_of_pages", "image_url", "created_at", "updated_at", "deleted_at"}).
				AddRow(1, "Go Programming", "A book about Go programming", "John Doe", pq.Array([]string{"Programming"}), 2024, 300, "http://example.com/go.jpg", expectedCreatedAt, nil, nil))

		expectedID := 1

		book, err := store.GetByID(ctx, expectedID)

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

func TestGetManyBooks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := NewBookStore(db)
	expectedCreatedAt := time.Now()

	ctx := utils.SetClaimsToContext(context.Background(), &types.CustomClaims{
		ID:               "ID-CRAZY",
		UserID:           1,
		Username:         "johndoe",
		Email:            "johndoe@email.com",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24))},
	})

	t.Run("missing userID in context", func(t *testing.T) {
		ctx := context.Background()

		id, err := store.GetMany(ctx)

		assert.Error(t, err)
		assert.Equal(t, "failed to retrieve userID from context", err.Error())
		assert.Zero(t, id)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		cancel()

		books, err := store.GetMany(ctx)

		assert.Error(t, err)
		assert.Nil(t, books)
		assert.True(t, errors.Is(err, context.Canceled))

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database connection error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT b.*
				FROM books b
				INNER JOIN users_books ub  ON
				ub.book_id = b.id
				WHERE ub.user_id = $1 
				AND b.deleted_at IS NULL;
			`)).
			WithArgs(1).
			WillReturnError(sql.ErrConnDone)

		book, err := store.GetMany(ctx)

		assert.Error(t, err)
		assert.Zero(t, book)
		assert.True(t, errors.Is(err, sql.ErrConnDone))

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("empty result set (no rows)", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT b.*
				FROM books b
				INNER JOIN users_books ub  ON
				ub.book_id = b.id
				WHERE ub.user_id = $1 
				AND b.deleted_at IS NULL;
			`)).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "name", "description", "author", "genres", "release_year", "number_of_pages", "image_url", "created_at", "updated_at", "deleted_at",
			}))

		books, err := store.GetMany(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, books)
		assert.Equal(t, 0, len(books))

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully get user books", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT b.*
			FROM books b
			INNER JOIN users_books ub ON ub.book_id = b.id
			WHERE ub.user_id = $1 
			AND b.deleted_at IS NULL;
		`)).
			WithArgs(1).
			WillReturnRows(
				sqlmock.NewRows([]string{
					"id", "name", "description", "author", "genres", "release_year", "number_of_pages", "image_url", "created_at", "updated_at", "deleted_at",
				}).
					AddRow(1, "Go Programming", "A book about Go programming", "John Doe", pq.Array([]string{"Programming"}), 2024, 300, "http://example.com/go.jpg", expectedCreatedAt, nil, nil).
					AddRow(2, "Clean Code", "A book about writing clean code", "Robert C. Martin", pq.Array([]string{"Programming", "Best Practices"}), 2008, 464, "http://example.com/clean-code.jpg", expectedCreatedAt, nil, nil),
			)

		books, err := store.GetMany(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, books)
		assert.Equal(t, 2, len(books))

		assert.Equal(t, 1, books[0].ID)
		assert.Equal(t, "Go Programming", books[0].Name)
		assert.Equal(t, "John Doe", books[0].Author)

		assert.Equal(t, 2, books[1].ID)
		assert.Equal(t, "Clean Code", books[1].Name)
		assert.Equal(t, "Robert C. Martin", books[1].Author)

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

	store := NewBookStore(db)

	ctx := utils.SetClaimsToContext(context.Background(), &types.CustomClaims{
		ID:               "ID-CRAZY",
		UserID:           1,
		Username:         "johndoe",
		Email:            "johndoe@email.com",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24))},
	})

	t.Run("missing userID in context", func(t *testing.T) {
		ctx := context.Background()

		book, err := store.UpdateByID(ctx, 1, types.UpdateBookPayload{})

		assert.Error(t, err)
		assert.Equal(t, "failed to retrieve userID from context", err.Error())
		assert.Nil(t, book)
	})

	t.Run("database did not find any row", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`
			UPDATE books SET 
			name = $3, 
			description = $4,
			author = $5,
			genres = $6,
			release_year = $7,
			number_of_pages = $8,
			image_url = $9,
			updated_at = $10
			WHERE id IN (
				SELECT b.id
				FROM books b
				INNER JOIN users_books ub ON ub.book_id = b.id
				WHERE b.id = $1 AND ub.user_id = $2
			)
			RETURNING 
				id, 
				name, 
				description, 
				author, 
				genres, 
				release_year, 
				number_of_pages, 
				image_url, 
				created_at, 
				deleted_at,
				updated_at;
			`)).
			WithArgs(
				1, // bookID
				1, // userID
				"Updated Book Name",
				"Updated Description",
				"John Doe",
				pq.Array([]string{"Genre1", "Genre2"}),
				2025,
				199,
				"http://google.com/somerandomimage.jpg",
				sqlmock.AnyArg(),
			).
			WillReturnError(sql.ErrNoRows)

		book, err := store.UpdateByID(ctx, 1, types.UpdateBookPayload{
			Name:          "Updated Book Name",
			Description:   "Updated Description",
			Author:        "John Doe",
			Genres:        []string{"Genre1", "Genre2"},
			ReleaseYear:   2025,
			NumberOfPages: 199,
			ImageUrl:      "http://google.com/somerandomimage.jpg",
		})

		assert.Nil(t, book)
		assert.Error(t, err)
		assert.Equal(t, err, sql.ErrNoRows)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully update book", func(t *testing.T) {
		mockDate := time.Date(0001, 01, 01, 0, 0, 0, 0, time.UTC)

		mock.ExpectQuery(regexp.QuoteMeta(`
			UPDATE books SET 
			name = $3, 
			description = $4,
			author = $5,
			genres = $6,
			release_year = $7,
			number_of_pages = $8,
			image_url = $9,
			updated_at = $10
			WHERE id IN (
				SELECT b.id
				FROM books b
				INNER JOIN users_books ub ON ub.book_id = b.id
				WHERE b.id = $1 AND ub.user_id = $2
			)
			RETURNING 
				id, 
				name, 
				description, 
				author, 
				genres, 
				release_year, 
				number_of_pages, 
				image_url, 
				created_at, 
				deleted_at,
				updated_at;
			`)).
			WithArgs(
				1, 1,
				"Updated Book Name",
				"Updated Description",
				"John Doe",
				pq.Array([]string{"Genre1", "Genre2"}),
				2025,
				199,
				"http://google.com/somerandomimage.jpg",
				sqlmock.AnyArg(),
			).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "name", "description", "author", "genres", "release_year",
				"number_of_pages", "image_url", "created_at", "deleted_at", "updated_at",
			}).AddRow(
				1,
				"Updated Book Name",
				"Updated Description",
				"Author Name",
				pq.Array([]string{"Genre1", "Genre2"}),
				2025,
				300,
				"http://example.com/image.jpg",
				mockDate,
				&mockDate,
				&mockDate,
			))

		updatedBook, err := store.UpdateByID(ctx, 1, types.UpdateBookPayload{
			Name:          "Updated Book Name",
			Description:   "Updated Description",
			Author:        "John Doe",
			Genres:        []string{"Genre1", "Genre2"},
			ReleaseYear:   2025,
			NumberOfPages: 199,
			ImageUrl:      "http://google.com/somerandomimage.jpg",
		})

		assert.NoError(t, err)
		assert.NotNil(t, updatedBook)
		assert.Equal(t, 1, updatedBook.ID)
		assert.Equal(t, "Updated Book Name", updatedBook.Name)
		assert.Equal(t, "Updated Description", updatedBook.Description)
		assert.Equal(t, []string{"Genre1", "Genre2"}, updatedBook.Genres)
		assert.Equal(t, 2025, updatedBook.ReleaseYear)
		assert.Equal(t, 300, updatedBook.NumberOfPages)
		assert.Equal(t, "http://example.com/image.jpg", updatedBook.ImageUrl)
		assert.Equal(t, mockDate, updatedBook.CreatedAt)
		assert.Equal(t, &mockDate, updatedBook.DeletedAt)
		assert.Equal(t, &mockDate, updatedBook.UpdatedAt)

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

	ctx := utils.SetClaimsToContext(context.Background(), &types.CustomClaims{
		ID:               "ID-CRAZY",
		UserID:           1,
		Username:         "johndoe",
		Email:            "johndoe@email.com",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24))},
	})

	t.Run("missing userID in context", func(t *testing.T) {
		ctx := context.Background()

		err := store.DeleteByID(ctx, 1)

		assert.Error(t, err)
		assert.Equal(t, "failed to retrieve userID from context", err.Error())
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
		mock.ExpectExec(regexp.QuoteMeta(`
			UPDATE books
			SET deleted_at = $3
			WHERE id IN (
				SELECT b.id
				FROM books b
				INNER JOIN users_books ub ON ub.book_id = b.id
				WHERE b.id = $1
				AND ub.user_id = $2
			)
			RETURNING id;
		`)).
			WithArgs(1, 1, sqlmock.AnyArg()).
			WillReturnError(ErrBookNotFound)

		err := store.DeleteByID(ctx, 1)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "book not found")
		assert.ErrorIs(t, err, ErrBookNotFound)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("database connection error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`
			UPDATE books
			SET deleted_at = $3
			WHERE id IN (
				SELECT b.id
				FROM books b
				INNER JOIN users_books ub ON ub.book_id = b.id
				WHERE b.id = $1
				AND ub.user_id = $2
			)
			RETURNING id;
		`)).
			WithArgs(1, 1, sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		err := store.DeleteByID(ctx, 1)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, sql.ErrConnDone))

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})

	t.Run("successfully delete book by ID", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`
			UPDATE books
			SET deleted_at = $3
			WHERE id IN (
				SELECT b.id
				FROM books b
				INNER JOIN users_books ub ON ub.book_id = b.id
				WHERE b.id = $1
				AND ub.user_id = $2
			)
			RETURNING id;
		`)).
			WithArgs(1, 1, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.DeleteByID(ctx, 1)

		assert.NoError(t, err)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet expectations: %v", err)
		}
	})
}
