package controller

import (
	"net/http"

	"github.com/hoyci/book-store-api/service"
	"github.com/hoyci/book-store-api/utils"
)

type HealthcheckController struct {
	Service service.HealthcheckServiceInterface
}

func NewHealthcheckController(service service.HealthcheckServiceInterface) *HealthcheckController {
	return &HealthcheckController{
		Service: service,
	}
}

func (h *HealthcheckController) HandleHealthcheck(w http.ResponseWriter, r *http.Request) {
	response := h.Service.HandleHealthcheck(r.Context())

	utils.WriteJSON(w, http.StatusOK, response)
}
