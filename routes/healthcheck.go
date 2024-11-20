package routes

import (
	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/handler"
	"github.com/hoyci/book-store-api/service"
)

func RegisterHealthcheckRoutes(router *mux.Router) {
	healthcheckService := service.NewHealthcheckService(config.Envs)
	healthcheckHandler := handler.NewHealthcheckHandler(healthcheckService)

	router.HandleFunc("/healthcheck", healthcheckHandler.HandleHealthcheck).Methods("GET")
}
