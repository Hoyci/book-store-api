package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hoyci/book-store-api/types"
	"github.com/lib/pq"
)

type BookStore struct {
	db *sql.DB
}

func NewBookStore(db *sql.DB) *BookStore {
	return &BookStore{db: db}
}

func (s *BookStore) Create(ctx context.Context, book types.Book) (int64, error) {
	result, err := s.db.ExecContext(
		ctx,
		"INSERT INTO books (name, description, author, genres, release_year, number_of_pages, image_url) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		book.Name,
		book.Description,
		book.Author,
		pq.Array(book.Genres),
		book.ReleaseYear,
		book.NumberOfPages,
		book.ImageUrl,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert book: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch last inserted id: %w", err)
	}

	return id, nil
}

func (s *BookStore) GetByID(ctx context.Context, id int) (*types.Book, error) {
	book := &types.Book{}

	err := s.db.QueryRowContext(ctx, "SELECT id, name, description, author, genres, release_year, number_of_pages, image_url, created_at FROM books WHERE id = $1", id).
		Scan(
			&book.ID,
			&book.Name,
			&book.Description,
			&book.Author,
			pq.Array(&book.Genres),
			&book.ReleaseYear,
			&book.NumberOfPages,
			&book.ImageUrl,
			&book.CreatedAt,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("book with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch book by id %d: %w", id, err)
	}

	return book, nil
}
