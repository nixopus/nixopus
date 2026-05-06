package utils

import (
	"encoding/json"
	"net/http"

	apilog "github.com/nixopus/nixopus/api/internal/log"
)

type jsonSuccessResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// jsonErrorResponse matches the Fuego HTTPError wire format so all error
// responses from middleware and Fuego handlers share a single envelope shape.
type jsonErrorResponse struct {
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
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
			apilog.Printf("Error marshaling response data: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(jsonErrorResponse{
				Title:  http.StatusText(http.StatusInternalServerError),
				Status: http.StatusInternalServerError,
				Detail: "failed to encode response data",
			})
			return
		}
		response.Data = json.RawMessage(encodedData)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		apilog.Printf("Error encoding response: %v", err)
	}
}

// SendErrorResponse writes an error response to the given http.ResponseWriter.
// The JSON shape matches Fuego's HTTPError so clients see one consistent envelope.
func SendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := jsonErrorResponse{
		Title:  http.StatusText(statusCode),
		Status: statusCode,
		Detail: message,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		apilog.Printf("Error encoding error response: %v", err)
	}
}
