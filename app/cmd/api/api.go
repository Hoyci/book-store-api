package api

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/config"
	_ "github.com/hoyci/book-store-api/docs"
	"github.com/hoyci/book-store-api/service/auth"
	"github.com/hoyci/book-store-api/service/book"
	"github.com/hoyci/book-store-api/service/healthcheck"
	"github.com/hoyci/book-store-api/service/user"
	"github.com/hoyci/book-store-api/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
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
	registry := prometheus.NewRegistry()
	metricsMiddleware := utils.New(registry, nil)

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	router.Use(utils.LoggingMiddleware)

	subrouter := router.PathPrefix("/api/v1").Subrouter()

	subrouter.HandleFunc(
		"/metrics",
		metricsMiddleware.WrapHandler(
			"metrics",
			promhttp.HandlerFor(
				registry,
				promhttp.HandlerOpts{},
			),
		),
	)

	subrouter.HandleFunc(
		"/swagger.json",
		metricsMiddleware.WrapHandler(
			"swagget_json",
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					http.ServeFile(w, r, "docs/swagger.json")
				},
			),
		),
	)

	subrouter.PathPrefix("/swagger/").Handler(
		metricsMiddleware.WrapHandler(
			"swagget_ui",
			httpSwagger.Handler(
				httpSwagger.URL("http://localhost:8080/api/v1/swagger.json"),
				httpSwagger.DeepLinking(true),
				httpSwagger.DocExpansion("none"),
				httpSwagger.DomID("swagger-ui"),
			),
		),
	).Methods(http.MethodGet)

	subrouter.HandleFunc(
		"/healthcheck",
		metricsMiddleware.WrapHandler("healthcheck", http.HandlerFunc(healthCheckHandler.HandleHealthCheck)),
	).Methods(http.MethodGet)

	subrouter.HandleFunc(
		"/auth",
		metricsMiddleware.WrapHandler("auth", http.HandlerFunc(authHandler.HandleUserLogin)),
	).Methods(http.MethodPost)
	subrouter.HandleFunc(
		"/auth/refresh",
		metricsMiddleware.WrapHandler("auth/refresh", http.HandlerFunc(authHandler.HandleRefreshToken)),
	).Methods(http.MethodPost)

	subrouter.HandleFunc(
		"/users",
		metricsMiddleware.WrapHandler("create_user", http.HandlerFunc(userHandler.HandleCreateUser)),
	).Methods(http.MethodPost)
	subrouter.Handle(
		"/users",
		metricsMiddleware.WrapHandler(
			"get_user",
			utils.AuthMiddleware(http.HandlerFunc(userHandler.HandleGetUserByID)),
		),
	).Methods(http.MethodGet)
	subrouter.Handle(
		"/users",
		metricsMiddleware.WrapHandler(
			"update_user",
			utils.AuthMiddleware(http.HandlerFunc(userHandler.HandleUpdateUserByID)),
		),
	).Methods(http.MethodPut)
	subrouter.Handle(
		"/users",
		metricsMiddleware.WrapHandler(
			"delete_user",
			utils.AuthMiddleware(http.HandlerFunc(userHandler.HandleDeleteUserByID)),
		),
	).Methods(http.MethodDelete)

	subrouter.Handle(
		"/books",
		metricsMiddleware.WrapHandler(
			"create_book",
			utils.AuthMiddleware(http.HandlerFunc(bookHandler.HandleCreateBook)),
		),
	).Methods(http.MethodPost)
	subrouter.Handle(
		"/books/{id}",
		metricsMiddleware.WrapHandler(
			"get_book_by_id",
			utils.AuthMiddleware(http.HandlerFunc(bookHandler.HandleGetBookByID)),
		),
	).Methods(http.MethodGet)
	subrouter.Handle(
		"/books",
		metricsMiddleware.WrapHandler(
			"get_books",
			utils.AuthMiddleware(http.HandlerFunc(bookHandler.HandleGetBooks)),
		),
	).Methods(http.MethodGet)
	subrouter.Handle(
		"/books/{id}",
		metricsMiddleware.WrapHandler(
			"update_book_by_id",
			utils.AuthMiddleware(http.HandlerFunc(bookHandler.HandleUpdateBookByID)),
		),
	).Methods(http.MethodPut)
	subrouter.Handle(
		"/books/{id}",
		metricsMiddleware.WrapHandler(
			"delete_book_by_id",
			utils.AuthMiddleware(http.HandlerFunc(bookHandler.HandleDeleteBookByID)),
		),
	).Methods(http.MethodDelete)

	s.Router = router

	return router
}

func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
