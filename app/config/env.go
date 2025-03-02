package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	Environment            string
	DatabaseURL            string
	JWTSecret              string
	TracerURL              string
	JWTExpirationInSeconds int64
}

var Envs = initConfig()

func initConfig() Config {
	godotenv.Load()

	return Config{
		Port:                   getEnv("PORT", "8182"),
		Environment:            getEnv("ENV", "development"),
		DatabaseURL:            getEnv("DATABASE_URL", "postgresql://user:password@localhost:5432/postgres?sslmode=disable"),
		TracerURL:              getEnv("TRACER_URL", "localhost:4317"),
		JWTSecret:              getEnv("SECRET_KEY", "ABRACADABARA"),
		JWTExpirationInSeconds: getEnvAsInt("JWT_EXP", 3600*24*7),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}

		return i
	}

	return fallback
}
