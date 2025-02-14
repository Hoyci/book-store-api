package types

type HealthcheckResponse struct {
	Status     string            `json:"status"`
	SystemInfo map[string]string `json:"system_info"`
}
