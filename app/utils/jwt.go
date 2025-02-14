package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/types"
)

func GenerateTestToken(userID int, username, email string) string {
	claims := types.CustomClaims{
		ID:       "mocked-id",
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(1 * time.Hour)},
		},
	}
	token, err := CreateJWTFromClaims(claims, config.Envs.JWTSecret)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate test token: %v", err))
	}
	return token
}

func CreateJWTFromClaims(claims types.CustomClaims, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return signedToken, nil
}

func CreateJWT(userID int, username string, email string, secretKey string, expTimeInSeconds int64, uuidGen types.UUIDGenerator) (string, error) {
	jti := uuidGen.New()

	claims := types.CustomClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expTimeInSeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "book-store-api",
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return signedToken, nil
}

func VerifyJWT(tokenString, secretKey string) (*types.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &types.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if claims, ok := token.Claims.(*types.CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
