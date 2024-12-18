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
	var id int
	err := s.db.QueryRowContext(
		ctx,
		"INSERT INTO books (name, description, author, genres, release_year, number_of_pages, image_url) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		book.Name,
		book.Description,
		book.Author,
		pq.Array(book.Genres),
		book.ReleaseYear,
		book.NumberOfPages,
		book.ImageUrl,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("unexpected error inserting book: %w", err)
	}

	return id, nil
}

func (s *BookStore) GetByID(ctx context.Context, id int) (*types.Book, error) {
	book := &types.Book{}

	err := s.db.QueryRowContext(ctx, "SELECT * FROM books WHERE id = $1 AND deleted_at IS null", id).
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
			&book.UpdatedAt,
			&book.DeletedAt,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no row found with id: '%d'", id)
		}
		return nil, fmt.Errorf("unexpected error getting book with id: '%d': %v", id, err)
	}

	return book, nil
}

func (s *BookStore) UpdateByID(ctx context.Context, id int, newBook types.UpdateBookPayload) (*types.Book, error) {
	query := fmt.Sprintf("UPDATE books SET updated_at = '%s', ", time.Now().Format("2006-01-02 15:04:05"))
	args := []any{}
	counter := 1

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

	if len(args) == 0 {
		return nil, fmt.Errorf("no fields to update for book with ID %d", id)
	}

	query = query[:len(query)-2] + fmt.Sprintf(" WHERE id = $%d RETURNING id, name, description, author, genres, release_year, number_of_pages, image_url, created_at, updated_at", counter)
	args = append(args, id)

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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no row found with id: '%d'", id)
		}
		return nil, fmt.Errorf("unexpected error updating book with id: '%d': %v", id, err)
	}

	return updatedBook, nil
}

func (s *BookStore) DeleteByID(ctx context.Context, id int) (int, error) {
	var returnedID int
	err := s.db.QueryRowContext(
		ctx,
		"UPDATE books SET deleted_at = $2 WHERE id = $1 RETURNING id",
		id,
		time.Now(),
	).Scan(&returnedID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no row found with id: '%d'", id)
		}
		return 0, fmt.Errorf("unexpected error deleting book with id: '%d': %v", id, err)
	}

	return returnedID, nil
}
