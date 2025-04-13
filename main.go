package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rob/go-web-server/mcp"
)

func main() {
	router := mux.NewRouter()

	// Create and configure MCP server
	mcpServer := mcp.NewServer()

	// MCP endpoint
	router.HandleFunc("/mcp", mcpServer.HandleMCP)

	// Regular web server endpoints
	router.HandleFunc("/", handleHome)
	router.HandleFunc("/health", handleHealth)

	// Start server
	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// handleHome is the handler for the root endpoint
func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "Welcome to the Go Web Server!")
}

// handleHealth is a simple health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Status: OK")
}
