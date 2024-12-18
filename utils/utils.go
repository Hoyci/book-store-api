package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(payload)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, err any) {
	WriteJSON(
		w,
		status,
		map[string]any{
			"error": err,
		},
	)
}

func StringPtr(s string) *string {
	return &s
}

func IntPtr(i int) *int {
	return &i
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func IsNil(value any) bool {
	switch v := value.(type) {
	case *string:
		return v == nil
	case *int:
		return v == nil
	case *[]string:
		return v == nil
	default:
		return value == nil
	}
}
