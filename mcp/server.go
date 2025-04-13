package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Server represents an MCP server instance
type Server struct {
	sessions     map[string]*Session
	sessionMutex sync.RWMutex
}

// NewServer creates a new MCP server instance
func NewServer() *Server {
	return &Server{
		sessions: make(map[string]*Session),
	}
}

// HandleMCP handles all MCP protocol requests
func (s *Server) HandleMCP(w http.ResponseWriter, r *http.Request) {
	// Validate Origin header for security
	origin := r.Header.Get("Origin")
	if !s.isValidOrigin(origin) {
		http.Error(w, "Invalid Origin", http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodPost:
		s.handlePost(w, r)
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodDelete:
		s.handleDelete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendError(w, &ErrorResponse{
			Code:    -32700,
			Message: "Parse error",
		})
		return
	}

	// Handle initialization request
	if request.Method == "initialize" {
		s.handleInitialize(w, r, &request)
		return
	}

	// Validate session for non-initialize requests
	sessionID := r.Header.Get("Mcp-Session-Id")
	if !s.validateSession(sessionID) {
		http.Error(w, "Invalid session", http.StatusNotFound)
		return
	}

	// Process the request based on method
	s.processRequest(w, r, &request)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create or validate session
	sessionID := r.Header.Get("Mcp-Session-Id")
	if !s.validateSession(sessionID) {
		http.Error(w, "Invalid session", http.StatusNotFound)
		return
	}

	// Keep connection alive and send SSE events
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Create a channel for client disconnection
	notify := r.Context().Done()

	// Create a channel for server messages
	messageChan := make(chan interface{}, 10)

	// Register client's message channel
	s.sessionMutex.Lock()
	session := s.sessions[sessionID]
	if session.MessageChannels == nil {
		session.MessageChannels = make([]chan interface{}, 0)
	}
	session.MessageChannels = append(session.MessageChannels, messageChan)
	s.sessionMutex.Unlock()

	// Cleanup function
	defer func() {
		s.sessionMutex.Lock()
		// Remove the message channel from the session
		for i, ch := range session.MessageChannels {
			if ch == messageChan {
				session.MessageChannels = append(session.MessageChannels[:i], session.MessageChannels[i+1:]...)
				break
			}
		}
		s.sessionMutex.Unlock()
		close(messageChan)
	}()

	// Send initial connection established message
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	// SSE event loop
	for {
		select {
		case <-notify:
			// Client disconnected
			return
		case msg := <-messageChan:
			// Convert message to JSON
			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}

			// Send the event
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-time.After(30 * time.Second):
			// Send keepalive comment
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("Mcp-Session-Id")
	if s.validateSession(sessionID) {
		s.sessionMutex.Lock()
		delete(s.sessions, sessionID)
		s.sessionMutex.Unlock()
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleInitialize(w http.ResponseWriter, r *http.Request, req *Request) {
	// Create new session
	sessionID := uuid.New().String()
	session := &Session{
		ID: sessionID,
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
		},
	}

	s.sessionMutex.Lock()
	s.sessions[sessionID] = session
	s.sessionMutex.Unlock()

	// Send initialization response
	w.Header().Set("Mcp-Session-Id", sessionID)
	s.sendResponse(w, Response{
		JSONRPC: JSONRPC{
			Version: "2.0",
			ID:      req.ID,
		},
		Result: s.marshalJSON(map[string]interface{}{
			"capabilities": session.Capabilities,
		}),
	})
}

func (s *Server) validateSession(sessionID string) bool {
	if sessionID == "" {
		return false
	}
	s.sessionMutex.RLock()
	defer s.sessionMutex.RUnlock()
	_, exists := s.sessions[sessionID]
	return exists
}

func (s *Server) isValidOrigin(origin string) bool {
	// Implement origin validation logic
	// For development, you might want to allow localhost
	return true
}

func (s *Server) processRequest(w http.ResponseWriter, r *http.Request, req *Request) {
	var response Response

	switch req.Method {
	case "tools/list":
		response = s.handleToolsList(req)
	case "tools/call":
		response = s.handleToolCall(req)
	default:
		response = Response{
			JSONRPC: JSONRPC{
				Version: "2.0",
				ID:      req.ID,
			},
			Error: &ErrorResponse{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}

	s.sendResponse(w, response)
}

func (s *Server) sendError(w http.ResponseWriter, err *ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		JSONRPC: JSONRPC{Version: "2.0"},
		Error:   err,
	})
}

func (s *Server) sendResponse(w http.ResponseWriter, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) marshalJSON(v interface{}) *json.RawMessage {
	data, _ := json.Marshal(v)
	raw := json.RawMessage(data)
	return &raw
}

// broadcastToSession sends a message to all connected clients for a session
func (s *Server) broadcastToSession(sessionID string, message interface{}) {
	s.sessionMutex.RLock()
	session, exists := s.sessions[sessionID]
	if !exists {
		s.sessionMutex.RUnlock()
		return
	}

	for _, ch := range session.MessageChannels {
		select {
		case ch <- message:
			// Message sent successfully
		default:
			// Channel is full, skip this client
		}
	}
	s.sessionMutex.RUnlock()
}
