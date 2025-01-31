package types

import (
	"context"
	"time"
)

type UserStore interface {
	Create(ctx context.Context, user CreateUserDatabasePayload) (*UserResponse, error)
	GetByID(ctx context.Context) (*UserResponse, error)
	GetByEmail(ctx context.Context) (*UserResponse, error)
	UpdateByID(ctx context.Context, user UpdateUserPayload) (*UserResponse, error)
	DeleteByID(ctx context.Context) error
}

type User struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"passwordHash"`
	CreatedAt    time.Time  `json:"createdAt"`
	DeletedAt    *time.Time `json:"deletedAt"`
	UpdatedAt    *time.Time `json:"updatedAt"`
}

type UserResponse struct {
	ID        int        `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"createdAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

type CreateUserRequestPayload struct {
	Username        string `json:"username" validate:"required,min=5"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8"`
}

type CreateUserByIDResponse struct {
	ID int `json:"id"`
}

type CreateUserDatabasePayload struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwor_hash"`
}

type UpdateUserPayload struct {
	Username string `json:"username" validate:"required,min=5"`
	Email    string `json:"email" validate:"required,email"`
}

type DeleteUserByIDResponse struct {
	ID int `json:"id"`
}
