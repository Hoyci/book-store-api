package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
	"github.com/lib/pq"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(ctx context.Context, newUser types.CreateUserDatabasePayload) (*types.User, error) {
	user := &types.User{}

	err := s.db.QueryRowContext(
		ctx,
		"INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id, username, email, created_at, updated_at, deleted_at",
		newUser.Username,
		newUser.Email,
		newUser.PasswordHash,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("unexpected error updating user %w", err)
	}

	return user, nil
}

func (s *UserStore) GetByID(ctx context.Context, id int) (*types.User, error) {
	user := &types.User{}

	err := s.db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = $1 AND deleted_at IS null", id).
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no row found with id: '%d'", id)
		}
		return nil, fmt.Errorf("unexpected error getting user with id: '%d'", id)
	}

	return user, nil
}

func (s *UserStore) UpdateByID(ctx context.Context, id int, newUser types.UpdateUserPayload) (*types.User, error) {
	query := fmt.Sprintf("UPDATE users SET updated_at = '%s', ", time.Now().Format("2006-01-02 15:04:05"))
	args := []any{}
	counter := 1

	fields := []struct {
		name  string
		value any
	}{
		{"username", newUser.Username},
		{"email", newUser.Email},
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
		return nil, fmt.Errorf("no fields to update for user with ID %d", id)
	}

	query = query[:len(query)-2] + fmt.Sprintf(" WHERE id = $%d RETURNING id, username, email, created_at, updated_at, deleted_at", counter)
	args = append(args, id)

	updatedUser := &types.User{}
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&updatedUser.ID,
		&updatedUser.Username,
		&updatedUser.Email,
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
		&updatedUser.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no row found with id: '%d'", id)
		}
		return nil, fmt.Errorf("unexpected error updating user with id: '%d': %v", id, err)
	}

	return updatedUser, nil
}

func (s *UserStore) DeleteByID(ctx context.Context, id int) (int, error) {
	var returnedID int
	err := s.db.QueryRowContext(
		ctx,
		"UPDATE users SET deleted_at = $2 WHERE id = $1 RETURNING id",
		id,
		time.Now(),
	).Scan(&returnedID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no row found with id: '%d'", id)
		}
		return 0, fmt.Errorf("unexpected error deleting user with id: '%d': %v", id, err)
	}

	return returnedID, nil
}
