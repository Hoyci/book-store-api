package api

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/controller"
	"github.com/hoyci/book-store-api/repository"
	"github.com/hoyci/book-store-api/service"
)

type APIServer struct {
	addr   string
	db     *sql.DB
	router *mux.Router
}

func NewApiServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr:   addr,
		db:     db,
		router: nil,
	}
}

func (s *APIServer) SetupRouter(
	healthcheckService service.HealthcheckServiceInterface,
	bookService service.BookServiceInterface,
) *mux.Router {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	healthcheckController := controller.NewHealthcheckController(healthcheckService)
	subrouter.HandleFunc("/healthcheck", healthcheckController.HandleHealthcheck).Methods("GET")

	bookController := controller.NewBookController(bookService)
	subrouter.HandleFunc("/book", bookController.HandleCreateBook).Methods(http.MethodPost)
	subrouter.HandleFunc("/book/{id}", bookController.HandleGetBookByID).Methods(http.MethodGet)
	subrouter.HandleFunc("/book/{id}", bookController.HandleDeleteBookByID).Methods(http.MethodDelete)

	s.router = router

	return router
}

func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.router == nil {
		healthcheckService := service.NewHealthcheckService(config.Envs)

		bookRepository := repository.NewBookRepository(s.db)
		bookService := service.NewBookService(bookRepository)

		s.SetupRouter(healthcheckService, bookService)
	}
	s.router.ServeHTTP(w, r)
}
