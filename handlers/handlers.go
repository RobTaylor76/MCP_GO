package handlers

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// JSONResponse sends a JSON response with the given status code and data
func JSONResponse(w http.ResponseWriter, statusCode int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
} 