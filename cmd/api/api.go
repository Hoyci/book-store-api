package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/routes"
)

type APIServer struct {
	addr string
	// db   *sql.DB
}

func NewApiServer(addr string) *APIServer {
	return &APIServer{
		addr: addr,
		// db:   db,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	routes.Routes(subrouter)

	log.Println("Listening on", s.addr)
	return http.ListenAndServe(s.addr, subrouter)
}
