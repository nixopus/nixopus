package types

// Response represents the standard API response format
// @Description Standard API response structure
type Response struct {
	// Status of the response ("success" or "error")
	// @Example success
	Status string `json:"status"`

	// Optional message providing additional information
	// @Example Operation completed successfully
	Message string `json:"message,omitempty"`

	// Error message in case of error
	// @Example Invalid input parameters
	Error string `json:"error,omitempty"`

	// Actual response data
	Data interface{} `json:"data,omitempty"`
}
