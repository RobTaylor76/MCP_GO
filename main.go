package main

import (
	"crypto/tls"
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

	// MCP endpoint with authentication middleware
	// New MCP endpoint
	router.HandleFunc("/mcp", mcpServer.AuthMiddleware(mcpServer.HandleMCP))

	// Legacy SSE endpoint for backward compatibility
	router.HandleFunc("/sse", mcpServer.AuthMiddleware(mcpServer.HandleLegacySSE))

	// Regular web server endpoints
	router.HandleFunc("/", handleHome)
	router.HandleFunc("/health", handleHealth)

	// Start HTTP server
	go func() {
		log.Printf("HTTP server starting on port 8080...\n")
		if err := http.ListenAndServe(":8080", router); err != nil {
			log.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Start HTTPS server
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	httpsServer := &http.Server{
		Addr:      ":8443",
		Handler:   router,
		TLSConfig: tlsConfig,
	}

	log.Printf("HTTPS server starting on port 8443...\n")
	if err := httpsServer.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
		log.Fatalf("HTTPS server error: %v", err)
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
