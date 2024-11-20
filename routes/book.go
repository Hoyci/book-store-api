package routes

import (
	"database/sql"

	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/handler"
	"github.com/hoyci/book-store-api/repository"
	"github.com/hoyci/book-store-api/service"
)

func RegisterBookRoutes(router *mux.Router, db *sql.DB) {
	bookRepository := repository.NewBookStore(db)
	bookService := service.NewBookService(bookRepository)
	bookHandler := handler.NewBookHandler(bookService)

	router.HandleFunc("/book/{id}", bookHandler.HandleCreateBook).Methods("POST")
	router.HandleFunc("/book", bookHandler.HandleCreateBook).Methods("POST")
}
