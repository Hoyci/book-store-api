package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
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

func (s *UserStore) GetByID(ctx context.Context) (*types.UserResponse, error) {
	claimsCtx, ok := utils.GetClaimsFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to retrieve userID from context")
	}
	userID := claimsCtx.UserID

	user := &types.UserResponse{}
	err := s.db.QueryRowContext(ctx, "SELECT id, username, email, created_at, updated_at, deleted_at  FROM users WHERE id = $1 AND deleted_at IS null", userID).
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

func (s *UserStore) GetByEmail(ctx context.Context) (*types.UserResponse, error) {
	claimsCtx, ok := utils.GetClaimsFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to retrieve userEmail from context")
	}
	userEmail := claimsCtx.Email

	user := &types.UserResponse{}
	err := s.db.QueryRowContext(ctx, "SELECT id, username, email, created_at, updated_at, deleted_at FROM users WHERE email = $1 AND deleted_at IS null", userEmail).
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

func (s *UserStore) UpdateByID(ctx context.Context, newUser types.UpdateUserPayload) (*types.UserResponse, error) {
	claimsCtx, ok := utils.GetClaimsFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to retrieve userID from context")
	}
	userID := claimsCtx.UserID

	query := `
			UPDATE users SET 
			username = $2, 
			email = $3,
			updated_at = $4
			WHERE id = $1
			RETURNING 
				id, 
				username, 
				email, 
				created_at, 
				deleted_at,
				updated_at;
			`

	updatedUser := &types.UserResponse{}
	err := s.db.QueryRowContext(
		ctx,
		query,
		userID,
		newUser.Username,
		newUser.Email,
		time.Now(),
	).Scan(
		&updatedUser.ID,
		&updatedUser.Username,
		&updatedUser.Email,
		&updatedUser.CreatedAt,
		&updatedUser.DeletedAt,
		&updatedUser.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

var ErrUserNotFound = errors.New("user not found")

func (s *UserStore) DeleteByID(ctx context.Context) error {
	claimsCtx, ok := utils.GetClaimsFromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to retrieve userID from context")
	}
	userID := claimsCtx.UserID

	result, err := s.db.ExecContext(
		ctx,
		"UPDATE users SET deleted_at = $2 WHERE id = $1",
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
		return fmt.Errorf("%w: %d", ErrUserNotFound, userID)
	}

	return nil
}
