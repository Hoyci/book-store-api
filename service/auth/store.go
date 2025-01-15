package auth

import (
	"context"
	"database/sql"

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
		return nil, err
	}

	return token, nil
}

func (s *AuthStore) UpsertRefreshToken(ctx context.Context, payload types.UpdateRefreshTokenPayload) error {
	token := &types.RefreshToken{}

	err := s.db.QueryRowContext(
		ctx,
		`INSERT INTO refresh_tokens (user_id, jti, expires_at) 
         VALUES ($1, $2, $3) 
         ON CONFLICT (user_id) 
         DO UPDATE SET 
             jti = EXCLUDED.jti, 
             expires_at = EXCLUDED.expires_at 
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
		return err
	}

	return nil
}
