package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/types"
)

// SendJSONResponse writes a JSON response to the given http.ResponseWriter.
//
// The response written is a types.Response with the given status, message, and
// data. If the response cannot be encoded, the error is logged.
func SendJSONResponse(w http.ResponseWriter, status string, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response := types.Response{
		Status:  status,
		Message: message,
		Data:    data,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// SendErrorResponse writes an error response to the given http.ResponseWriter.
//
// The response written is a types.Response with the given status code and error
// message. If the response cannot be encoded, the error is logged.
func SendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := types.Response{
		Status: "error",
		Error:  message,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}
