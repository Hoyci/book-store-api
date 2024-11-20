package handler

import (
	"net/http"

	"github.com/hoyci/book-store-api/service"
	"github.com/hoyci/book-store-api/utils"
)

type HealthCheckHandler struct {
	Service *service.HealthcheckService
}

func NewHealthcheckHandler(service *service.HealthcheckService) *HealthCheckHandler {
	return &HealthCheckHandler{
		Service: service,
	}
}

func (h *HealthCheckHandler) HandleHealthcheck(w http.ResponseWriter, r *http.Request) {
	response := h.Service.CheckHealth()

	utils.WriteJSON(w, http.StatusOK, response)
}
