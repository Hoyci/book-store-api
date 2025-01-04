package api

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/service/book"
	"github.com/hoyci/book-store-api/service/healthcheck"
	"github.com/hoyci/book-store-api/service/user"
)

type APIServer struct {
	addr   string
	db     *sql.DB
	Router *mux.Router
	Config config.Config
}

func NewApiServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr:   addr,
		db:     db,
		Router: nil,
		Config: config.Envs,
	}
}

func (s *APIServer) SetupRouter(
	healthCheckHandler *healthcheck.HealthCheckHandler,
	bookHandler *book.BookHandler,
	userHandler *user.UserHandler,
) *mux.Router {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	subrouter.HandleFunc("/healthcheck", healthCheckHandler.HandleHealthCheck).Methods(http.MethodGet)

	subrouter.HandleFunc("/book", bookHandler.HandleCreateBook).Methods(http.MethodPost)
	subrouter.HandleFunc("/book/{id}", bookHandler.HandleGetBookByID).Methods(http.MethodGet)
	subrouter.HandleFunc("/book/{id}", bookHandler.HandleUpdateBookByID).Methods(http.MethodPut)
	subrouter.HandleFunc("/book/{id}", bookHandler.HandleDeleteBookByID).Methods(http.MethodDelete)

	subrouter.HandleFunc("/user", userHandler.HandleCreateUser).Methods(http.MethodPost)
	subrouter.HandleFunc("/user/{id}", userHandler.HandleGetUserByID).Methods(http.MethodGet)
	subrouter.HandleFunc("/user/{id}", userHandler.HandleUpdateUserByID).Methods(http.MethodPut)
	subrouter.HandleFunc("/user/{id}", userHandler.HandleDeleteUserByID).Methods(http.MethodDelete)

	s.Router = router

	return router
}

func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
