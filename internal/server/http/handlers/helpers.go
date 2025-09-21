package handlers

import (
	"encoding/json"
	"github.com/turtacn/QuantaID/pkg/types"
	"net/http"
)

// WriteJSON writes a JSON response with a given status code.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// WriteJSONError writes a standardized JSON error response.
func WriteJSONError(w http.ResponseWriter, err *types.Error, defaultStatus int) {
	status := err.HttpStatus
	if status == 0 {
		status = defaultStatus
	}

	errorResponse := struct {
		Error *types.Error `json:"error"`
	}{
		Error: err,
	}

	WriteJSON(w, status, errorResponse)
}

//Personal.AI order the ending
