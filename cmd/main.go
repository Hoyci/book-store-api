package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hoyci/book-store-api/cmd/api"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/db"
	"github.com/hoyci/book-store-api/service/book"
	"github.com/hoyci/book-store-api/service/healthcheck"
)

func main() {
	db := db.NewPGStorage()
	path := fmt.Sprintf("127.0.0.1:%s", config.Envs.Port)

	apiServer := api.NewApiServer(path, db)

	healthCheckHandler := healthcheck.NewHealthCheckHandler(config.Envs)

	bookStore := book.NewBookStore(db)
	bookHandler := book.NewBookHandler(bookStore)

	apiServer.SetupRouter(healthCheckHandler, bookHandler)

	log.Println("Listening on:", path)
	http.ListenAndServe(path, apiServer.Router)
}
