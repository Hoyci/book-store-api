package utils

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hoyci/book-store-api/config"
	"github.com/sirupsen/logrus"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		Log.WithFields(logrus.Fields{
			"method":      r.Method,
			"url":         r.URL.String(),
			"status_code": rw.statusCode,
			"duration_ms": duration.Milliseconds(),
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		}).Info("API Request")
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			WriteError(
				w,
				http.StatusUnauthorized,
				fmt.Errorf("user did not send an authorization header"),
				"AuthMiddleware",
				"Missing authorization header",
			)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			WriteError(
				w,
				http.StatusUnauthorized,
				fmt.Errorf("user sent an authorization header out of format"),
				"AuthMiddleware",
				"Invalid authorization header format",
			)
			return
		}

		token := parts[1]

		claims, err := VerifyJWT(token, config.Envs.JWTSecret)
		if err != nil {
			WriteError(
				w,
				http.StatusUnauthorized,
				fmt.Errorf("user sent an invalid or expired authorization header"),
				"AuthMiddleware",
				"Invalid or expired token",
			)
			return
		}

		ctx := r.Context()
		ctx = SetClaimsToContext(ctx, claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
