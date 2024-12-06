package db

import (
	"database/sql"
	"log"

	"github.com/hoyci/book-store-api/config"
)

func NewPGStorage() *sql.DB {
	db, err := sql.Open("postgres", config.Envs.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	return db
}
