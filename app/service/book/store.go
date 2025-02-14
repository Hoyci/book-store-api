package book

import (
	"context"
	"database/sql"
	"errors"
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

func (s *BookStore) GetMany(ctx context.Context) ([]*types.Book, error) {
	claimsCtx, ok := utils.GetClaimsFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to retrieve userID from context")
	}
	userID := claimsCtx.UserID

	rows, err := s.db.QueryContext(
		ctx,
		`
		SELECT b.*
		FROM books b
		INNER JOIN users_books ub  ON
		ub.book_id = b.id
		WHERE ub.user_id = $1 
		AND b.deleted_at IS NULL;
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := []*types.Book{}

	for rows.Next() {
		book := &types.Book{}
		err := rows.Scan(
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
		books = append(books, book)
	}

	return books, nil
}

func (s *BookStore) UpdateByID(ctx context.Context, bookID int, newBook types.UpdateBookPayload) (*types.Book, error) {
	claimsCtx, ok := utils.GetClaimsFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to retrieve userID from context")
	}
	userID := claimsCtx.UserID

	query := `
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
			`

	updatedBook := &types.Book{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		bookID,
		userID,
		newBook.Name,
		newBook.Description,
		newBook.Author,
		pq.Array(newBook.Genres),
		newBook.ReleaseYear,
		newBook.NumberOfPages,
		newBook.ImageUrl,
		time.Now(),
	).Scan(
		&updatedBook.ID,
		&updatedBook.Name,
		&updatedBook.Description,
		&updatedBook.Author,
		pq.Array(&updatedBook.Genres),
		&updatedBook.ReleaseYear,
		&updatedBook.NumberOfPages,
		&updatedBook.ImageUrl,
		&updatedBook.CreatedAt,
		&updatedBook.DeletedAt,
		&updatedBook.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return updatedBook, nil
}

var ErrBookNotFound = errors.New("book not found")

func (s *BookStore) DeleteByID(ctx context.Context, bookID int) error {
	claimsCtx, ok := utils.GetClaimsFromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to retrieve userID from context")
	}
	userID := claimsCtx.UserID

	result, err := s.db.ExecContext(
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
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: %d", ErrBookNotFound, bookID)
	}

	return nil
}
