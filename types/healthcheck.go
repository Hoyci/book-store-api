package types

type HealthCheckResponse struct {
	Status     string            `json:"status"`
	SystemInfo map[string]string `json:"system_info"`
}
