package auth

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hoyci/book-store-api/types"
)

type AuthStore struct {
	db *sql.DB
}

func NewAuthStore(db *sql.DB) *AuthStore {
	return &AuthStore{db: db}
}

func (s *AuthStore) GetRefreshTokenByUserID(ctx context.Context, userID int) (*types.RefreshToken, error) {
	token := &types.RefreshToken{}

	err := s.db.QueryRowContext(ctx, "SELECT * FROM refresh_tokens WHERE user_id = $1", userID).
		Scan(
			&token.ID,
			&token.UserID,
			&token.Jti,
			&token.CreatedAt,
			&token.ExpiresAt,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no row found with id: '%d'", userID)
		}
		return nil, fmt.Errorf("unexpected error getting refresh_token with user_id: '%d'", userID)
	}

	return token, nil
}

func (s *AuthStore) UpdateRefreshTokenByUserID(ctx context.Context, payload types.UpdateRefreshTokenPayload) error {
	token := &types.RefreshToken{}

	err := s.db.QueryRowContext(
		ctx,
		`UPDATE refresh_tokens 
         SET jti = $2, expires_at = $3 
         WHERE user_id = $1 
         RETURNING id, user_id, jti, created_at, expires_at`,
		payload.UserID,
		payload.Jti,
		payload.ExpiresAt,
	).Scan(
		&token.ID,
		&token.UserID,
		&token.Jti,
		&token.CreatedAt,
		&token.ExpiresAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no row found with user_id: '%d'", payload.UserID)
		}
		return fmt.Errorf("unexpected error updating refresh_token with user_id: '%d': %v", payload.UserID, err)
	}

	return nil
}
