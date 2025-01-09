package types

import (
	"context"
	"time"
)

type AuthStore interface {
	UpdateRefreshTokenByUserID(ctx context.Context, payload UpdateRefreshTokenPayload) (*RefreshToken, error)
	GetRefreshTokenByUserID(ctx context.Context, userID int) (*RefreshToken, error)
}

type UserLoginPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UserLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenPayload struct {
	AccessToken string `json:"access_token" validate:"required"`
}

type RefreshToken struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	Jti       string    `db:"jti"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

type UpdateRefreshTokenPayload struct {
	UserID    int       `db:"user_id"`
	Jti       string    `db:"jti"`
	ExpiresAt time.Time `db:"expires_at"`
}
