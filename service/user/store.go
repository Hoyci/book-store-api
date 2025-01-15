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

func (s *UserStore) Create(ctx context.Context, newUser types.CreateUserDatabasePayload) (*types.UserResponse, error) {
	user := &types.UserResponse{}

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
		return nil, err
	}

	return user, nil
}

func (s *UserStore) GetByID(ctx context.Context, id int) (*types.UserResponse, error) {
	user := &types.UserResponse{}

	err := s.db.QueryRowContext(ctx, "SELECT id, username, email, created_at, updated_at, deleted_at  FROM users WHERE id = $1 AND deleted_at IS null", id).
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*types.UserResponse, error) {
	user := &types.UserResponse{}
	err := s.db.QueryRowContext(ctx, "SELECT id, username, email, created_at, updated_at, deleted_at FROM users WHERE email = $1 AND deleted_at IS null", email).
		Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserStore) UpdateByID(ctx context.Context, id int, newUser types.UpdateUserPayload) (*types.UserResponse, error) {
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

	updatedUser := &types.UserResponse{}
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&updatedUser.ID,
		&updatedUser.Username,
		&updatedUser.Email,
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
		&updatedUser.DeletedAt,
	)
	if err != nil {
		return nil, err
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
		return 0, err
	}

	return returnedID, nil
}
