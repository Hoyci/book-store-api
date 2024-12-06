package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hoyci/book-store-api/types"
	"github.com/lib/pq"
)

type BookRepository struct {
	db *sql.DB
}

func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{db: db}
}

func (s *BookRepository) Create(ctx context.Context, book types.CreateBookPayload) (int64, error) {
	var id int64
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
		return 0, &InsertError{
			Entity: book.Name,
			Err:    err,
		}
	}

	return id, nil
}

func (s *BookRepository) GetByID(ctx context.Context, id int) (*types.Book, error) {
	book := &types.Book{}

	err := s.db.QueryRowContext(ctx, "SELECT id, name, description, author, genres, release_year, number_of_pages, image_url, created_at FROM books WHERE id = $1 and deleted_at = null", id).
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
			return nil, &ResourceNotFoundError{ID: id}
		}
		return nil, &InternalDatabaseError{
			Message: fmt.Sprintf("failed to fetch book by ID %d", id),
			Err:     err,
		}
	}

	return book, nil
}

// Verificar se o body está correto no controller
// Verificar se há valores para serem atualizados no controller
func (s *BookRepository) UpdateByID(ctx context.Context, id int64, updates types.BookUpdatePayload) error {
	query := "UPDATE books SET "
	args := []any{}
	counter := 1

	fields := []struct {
		name  string
		value any
	}{
		{"name", updates.Name},
		{"description", updates.Description},
		{"author", updates.Author},
		{"genres", updates.Genres},
		{"release_year", updates.ReleaseYear},
		{"number_of_pages", updates.NumberOfPages},
		{"image_url", updates.ImageUrl},
	}

	for _, field := range fields {
		if !isNil(field.value) {
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
		return fmt.Errorf("no fields to update for book with ID %d", id)
	}

	query = query[:len(query)-2] + fmt.Sprintf(" WHERE id = $%d", counter)
	args = append(args, id)

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return &UpdateError{
			Entity: "book",
			ID:     int(id),
			Err:    err,
		}
	}

	return nil
}

func (s *BookRepository) DeleteByID(ctx context.Context, id int) (int64, error) {
	var id_returned int64
	err := s.db.QueryRowContext(
		ctx,
		"UPDATE books SET deleted_at = $2 WHERE id = $1 RETURNING id",
		id,
		time.Now(),
	).Scan(&id_returned)
	if err != nil {
		return 0, &DeleteError{
			Entity: "book",
			ID:     id,
			Err:    err,
		}
	}

	return id_returned, nil
}

func isNil(value any) bool {
	switch v := value.(type) {
	case *string:
		return v == nil
	case *int32:
		return v == nil
	case *[]string:
		return v == nil
	default:
		return value == nil
	}
}
