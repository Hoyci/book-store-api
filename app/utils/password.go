package utils

import (
	"context"

	"go.opentelemetry.io/otel"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(ctx context.Context, password string) (string, error) {
	tracer := otel.Tracer("utils")
	_, span := tracer.Start(ctx, "Utils.HashPassword")
	defer span.End()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPassword(ctx context.Context, hashedPassword, password string) error {
	tracer := otel.Tracer("utils")
	_, span := tracer.Start(ctx, "Utils.CheckPassword")
	defer span.End()

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err
}
