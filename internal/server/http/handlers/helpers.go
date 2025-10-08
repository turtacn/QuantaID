package handlers

import (
	"encoding/json"
	"github.com/turtacn/QuantaID/pkg/types"
	"net/http"
)

// WriteJSON is a helper function that marshals a given data structure to JSON,
// sets the appropriate content-type header, and writes it to the http.ResponseWriter
// with the specified status code.
//
// Parameters:
//   - w: The http.ResponseWriter to write the response to.
//   - status: The HTTP status code to set for the response.
//   - data: The data to be encoded as JSON. Can be nil.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// WriteJSONError is a helper function that formats a standardized application error
// into a JSON response. It uses the HTTP status from the error object if available,
// otherwise it falls back to a default status.
//
// Parameters:
//   - w: The http.ResponseWriter to write the error response to.
//   - err: The application error to be encoded.
//   - defaultStatus: The HTTP status code to use if the error does not specify one.
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
