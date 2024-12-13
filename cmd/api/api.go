package api

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/service/book"
	"github.com/hoyci/book-store-api/service/healthcheck"
)

type APIServer struct {
	addr   string
	db     *sql.DB
	Router *mux.Router
	config config.Config
}

func NewApiServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr:   addr,
		db:     db,
		Router: nil,
		config: config.Envs,
	}
}

func (s *APIServer) SetupRouter(
	healthCheckHandler *healthcheck.HealthCheckHandler,
	bookHandler *book.BookHandler,
) *mux.Router {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	subrouter.HandleFunc("/healthcheck", healthCheckHandler.HandleHealthCheck).Methods(http.MethodGet)

	subrouter.HandleFunc("/book", bookHandler.HandleCreateBook).Methods(http.MethodPost)
	subrouter.HandleFunc("/book/{id}", bookHandler.HandleGetBookByID).Methods(http.MethodGet)
	subrouter.HandleFunc("/book/{id}", bookHandler.HandleUpdateBookByID).Methods(http.MethodPut)
	subrouter.HandleFunc("/book/{id}", bookHandler.HandleDeleteBookByID).Methods(http.MethodDelete)

	s.Router = router

	return router
}

func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
