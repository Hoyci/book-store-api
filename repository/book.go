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

func (s *BookRepository) Create(ctx context.Context, book types.CreateBookPayload) (int, error) {
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
func (s *BookRepository) UpdateByID(ctx context.Context, id int, newBook types.UpdateBookPayload) (*types.Book, error) {
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
		return nil, &UpdateError{
			Entity: "book",
			ID:     int(id),
			Err:    err,
		}
	}

	return updatedBook, nil
}

func (s *BookRepository) DeleteByID(ctx context.Context, id int) (int, error) {
	var id_returned int
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
