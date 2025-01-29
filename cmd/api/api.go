package api

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/service/auth"
	"github.com/hoyci/book-store-api/service/book"
	"github.com/hoyci/book-store-api/service/healthcheck"
	"github.com/hoyci/book-store-api/service/user"
	"github.com/hoyci/book-store-api/utils"
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
	authHandler *auth.AuthHandler,
) *mux.Router {
	utils.InitLogger()
	router := mux.NewRouter()

	router.Use(utils.LoggingMiddleware)

	subrouter := router.PathPrefix("/api/v1").Subrouter()

	subrouter.HandleFunc("/healthcheck", healthCheckHandler.HandleHealthCheck).Methods(http.MethodGet)

	subrouter.HandleFunc("/auth", authHandler.HandleUserLogin).Methods(http.MethodPost)
	subrouter.HandleFunc("/auth/refresh", authHandler.HandleRefreshToken).Methods(http.MethodPost)

	subrouter.Handle("/books", utils.AuthMiddleware(http.HandlerFunc(bookHandler.HandleCreateBook))).Methods(http.MethodPost)
	subrouter.Handle("/books/{id}", utils.AuthMiddleware(http.HandlerFunc(bookHandler.HandleGetBookByID))).Methods(http.MethodGet)
	subrouter.Handle("/books", utils.AuthMiddleware(http.HandlerFunc(bookHandler.HandleGetBooks))).Methods(http.MethodGet)
	subrouter.Handle("/books/{id}", utils.AuthMiddleware(http.HandlerFunc(bookHandler.HandleUpdateBookByID))).Methods(http.MethodPut)
	subrouter.Handle("/books/{id}", utils.AuthMiddleware(http.HandlerFunc(bookHandler.HandleDeleteBookByID))).Methods(http.MethodDelete)

	subrouter.HandleFunc("/users", userHandler.HandleCreateUser).Methods(http.MethodPost)
	subrouter.HandleFunc("/users/{id}", userHandler.HandleGetUserByID).Methods(http.MethodGet)
	subrouter.HandleFunc("/users/{id}", userHandler.HandleUpdateUserByID).Methods(http.MethodPut)
	subrouter.HandleFunc("/users/{id}", userHandler.HandleDeleteUserByID).Methods(http.MethodDelete)

	s.Router = router

	return router
}

func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
