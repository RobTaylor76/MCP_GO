package mcp

import "encoding/json"

// JSONRPC represents the base JSON-RPC 2.0 message structure
type JSONRPC struct {
	Version string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
}

// Request represents a JSON-RPC request message
type Request struct {
	JSONRPC
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC response message
type Response struct {
	JSONRPC
	Result *json.RawMessage `json:"result,omitempty"`
	Error  *ErrorResponse   `json:"error,omitempty"`
}

// ErrorResponse represents a JSON-RPC error object
type ErrorResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Notification represents a JSON-RPC notification message
type Notification struct {
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Session represents an MCP session
type Session struct {
	ID              string
	Capabilities    map[string]interface{}
	MessageChannels []chan interface{}
	// Add other session-specific data as needed
}

// CancellationParams represents the parameters for a cancellation notification
type CancellationParams struct {
	RequestID string `json:"requestId"`
	Reason    string `json:"reason,omitempty"`
}

// CancellationNotification represents a cancellation notification
type CancellationNotification struct {
	Version string             `json:"jsonrpc"`
	Method  string             `json:"method"`
	Params  CancellationParams `json:"params"`
}
