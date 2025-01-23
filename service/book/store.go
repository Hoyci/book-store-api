package book

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
	"github.com/lib/pq"
)

type BookStore struct {
	db *sql.DB
}

func NewBookStore(db *sql.DB) *BookStore {
	return &BookStore{db: db}
}

func (s *BookStore) Create(ctx context.Context, book types.CreateBookPayload) (int, error) {
	var bookID int
	claimsCtx, ok := utils.GetClaimsFromContext(ctx)
	if !ok {
		return 0, fmt.Errorf("failed to retrieve userID from context")
	}
	userID := claimsCtx.UserID

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	err = tx.QueryRowContext(
		ctx,
		`
        INSERT INTO books (name, description, author, genres, release_year, number_of_pages, image_url) 
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
        `,
		book.Name,
		book.Description,
		book.Author,
		pq.Array(book.Genres),
		book.ReleaseYear,
		book.NumberOfPages,
		book.ImageUrl,
	).Scan(&bookID)
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(
		ctx,
		`
        INSERT INTO users_books (user_id, book_id) 
        VALUES ($1, $2)
        `,
		userID,
		bookID,
	)
	if err != nil {
		return 0, err
	}

	return bookID, nil
}

func (s *BookStore) GetByID(ctx context.Context, bookID int) (*types.Book, error) {
	book := &types.Book{}
	claimsCtx, ok := utils.GetClaimsFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to retrieve userID from context")
	}
	userID := claimsCtx.UserID

	err := s.db.QueryRowContext(
		ctx,
		`
		SELECT 
		b.* 
		FROM books b
		INNER JOIN users_books ub ON ub.book_id = b.id
		WHERE b.id = $1
		AND ub.user_id = $2
		AND b.deleted_at IS NULL;
		`,
		bookID,
		userID,
	).Scan(
		&book.ID,
		&book.Name,
		&book.Description,
		&book.Author,
		pq.Array(&book.Genres),
		&book.ReleaseYear,
		&book.NumberOfPages,
		&book.ImageUrl,
		&book.CreatedAt,
		&book.UpdatedAt,
		&book.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return book, nil
}

func (s *BookStore) UpdateByID(ctx context.Context, bookID int, newBook types.UpdateBookPayload) (*types.Book, error) {
	claimsCtx, ok := utils.GetClaimsFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to retrieve userID from context")
	}
	userID := claimsCtx.UserID

	query := "UPDATE books SET updated_at = $1, "
	args := []any{time.Now()}
	counter := 2

	fields := []struct {
		name  string
		value any
	}{
		{"name", newBook.Name},
		{"description", newBook.Description},
		{"author", newBook.Author},
		{"genres", newBook.Genres},
		{"release_year", newBook.ReleaseYear},
		{"number_of_pages", newBook.NumberOfPages},
		{"image_url", newBook.ImageUrl},
	}

	for _, field := range fields {
		if !utils.IsNil(field.value) {
			query += fmt.Sprintf("%s = $%d, ", field.name, counter)
			if ptr, ok := field.value.(*[]string); ok {
				args = append(args, pq.Array(*ptr))
			} else {
				args = append(args, field.value)
			}
			counter++
		}
	}

	if len(args) == 1 {
		return nil, fmt.Errorf("no fields to update for book with ID %d", bookID)
	}

	query = query[:len(query)-2] + fmt.Sprintf(`
		WHERE id IN (
			SELECT b.id
			FROM books b
			INNER JOIN users_books ub ON ub.book_id = b.id
			WHERE b.id = $%d AND ub.user_id = $%d
		) 
		RETURNING id, name, description, author, genres, release_year, number_of_pages, image_url, created_at, updated_at
	`, counter, counter+1)
	args = append(args, bookID, userID)

	updatedBook := &types.Book{}
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&updatedBook.ID,
		&updatedBook.Name,
		&updatedBook.Description,
		&updatedBook.Author,
		pq.Array(&updatedBook.Genres),
		&updatedBook.ReleaseYear,
		&updatedBook.NumberOfPages,
		&updatedBook.ImageUrl,
		&updatedBook.CreatedAt,
		&updatedBook.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return updatedBook, nil
}

func (s *BookStore) DeleteByID(ctx context.Context, bookID int) (int, error) {
	var returnedID int
	claimsCtx, ok := utils.GetClaimsFromContext(ctx)
	if !ok {
		return 0, fmt.Errorf("failed to retrieve userID from context")
	}
	userID := claimsCtx.UserID

	err := s.db.QueryRowContext(
		ctx,
		`
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
		`,
		bookID,
		userID,
		time.Now(),
	).Scan(&returnedID)
	if err != nil {
		return 0, err
	}

	return returnedID, nil
}
