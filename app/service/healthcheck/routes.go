package healthcheck

import (
	"net/http"

	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
)

type HealthCheckHandler struct {
	cfg config.Config
}

func NewHealthCheckHandler(cfg config.Config) *HealthCheckHandler {
	return &HealthCheckHandler{
		cfg: cfg,
	}
}

func (h *HealthCheckHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, &types.HealthcheckResponse{
		Status: "available",
		SystemInfo: map[string]string{
			"environment": h.cfg.Environment,
		},
	})
}
