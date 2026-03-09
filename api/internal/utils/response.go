package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type jsonSuccessResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type jsonErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// SendJSONResponse writes a JSON response to the given http.ResponseWriter.
//
// The response written is a typed JSON envelope with status, message, and data.
func SendJSONResponse(w http.ResponseWriter, status string, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	response := jsonSuccessResponse{
		Status:  status,
		Message: message,
	}

	if data != nil {
		encodedData, err := json.Marshal(data)
		if err != nil {
			log.Printf("Error marshaling response data: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(jsonErrorResponse{
				Status: "error",
				Error:  "failed to encode response data",
			})
			return
		}
		response.Data = json.RawMessage(encodedData)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// SendErrorResponse writes an error response to the given http.ResponseWriter.
func SendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := jsonErrorResponse{
		Status: "error",
		Error:  message,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}
