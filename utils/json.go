package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hoyci/book-store-api/types"
	"github.com/sirupsen/logrus"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(payload)
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func WriteError[T types.ErrorResponse](w http.ResponseWriter, status int, err error, context string, clientError T) {
	Log.WithFields(logrus.Fields{
		"error":   err.Error(),
		"context": context,
	}).Error(err.Error())

	WriteJSON(w, status, clientError)
}
