package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// Tool represents an MCP tool definition
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ToolCallRequest represents a tool call request parameters
type ToolCallRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ContentItem represents a single content item in a tool result
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError"`
}

// WeatherArgs represents the input arguments for the get_weather tool
type WeatherArgs struct {
	Location string `json:"location"`
}

// weatherTool is our sample weather tool implementation
var weatherTool = Tool{
	Name:        "get_weather",
	Description: "Get current weather information for a location",
	InputSchema: json.RawMessage(`{
		"type": "object",
		"properties": {
			"location": {
				"type": "string",
				"description": "City name or zip code"
			}
		},
		"required": ["location"]
	}`),
}

// getAvailableTools returns the list of available tools
func (s *Server) getAvailableTools() []Tool {
	return []Tool{weatherTool}
}

// handleToolsList handles the tools/list request
func (s *Server) handleToolsList(req *Request) Response {
	return Response{
		JSONRPC: JSONRPC{
			Version: "2.0",
			ID:      req.ID,
		},
		Result: s.marshalJSON(map[string]interface{}{
			"tools": s.getAvailableTools(),
		}),
	}
}

// handleToolCall handles the tools/call request
func (s *Server) handleToolCall(req *Request) Response {
	var toolReq ToolCallRequest
	if err := json.Unmarshal(req.Params, &toolReq); err != nil {
		return Response{
			JSONRPC: JSONRPC{
				Version: "2.0",
				ID:      req.ID,
			},
			Error: &ErrorResponse{
				Code:    -32602,
				Message: "Invalid tool call parameters",
			},
		}
	}

	// Handle get_weather tool
	if toolReq.Name == "get_weather" {
		//	return s.handleWeatherTool(req.Context(), req.ID, toolReq.Arguments)
		return s.handleWeatherTool(context.Background(), req.ID, toolReq.Arguments)
	}

	return Response{
		JSONRPC: JSONRPC{
			Version: "2.0",
			ID:      req.ID,
		},
		Error: &ErrorResponse{
			Code:    -32602,
			Message: fmt.Sprintf("Unknown tool: %s", toolReq.Name),
		},
	}
}

// handleWeatherTool handles the get_weather tool execution
func (s *Server) handleWeatherTool(ctx context.Context, reqID any, args json.RawMessage) Response {
	var weatherArgs WeatherArgs
	if err := json.Unmarshal(args, &weatherArgs); err != nil {
		return Response{
			JSONRPC: JSONRPC{
				Version: "2.0",
				ID:      reqID,
			},
			Result: s.marshalJSON(ToolResult{
				Content: []ContentItem{
					{
						Type: "text",
						Text: "Failed to parse weather arguments",
					},
				},
				IsError: true,
			}),
		}
	}

	// Simulate a long-running operation that can be cancelled
	// select {
	// case <-ctx.Done():
	// Request was cancelled
	//	return Response{
	//		JSONRPC: JSONRPC{
	//			Version: "2.0",
	//			ID:      reqID,
	//		},
	//		Result: s.marshalJSON(ToolResult{
	//			Content: []ContentItem{
	//				{
	//					Type: "text",
	//					Text: "Weather request cancelled",
	//				},
	//			},
	//			IsError: true,
	//		}),
	//	}
	// case <-time.After(2 * time.Second): // Simulate API delay
	// Return mock weather data
	result := ToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Current weather in %s:\nTemperature: 72°F\nConditions: Partly cloudy", weatherArgs.Location),
			},
		},
		IsError: false,
	}

	return Response{
		JSONRPC: JSONRPC{
			Version: "2.0",
			ID:      reqID,
		},
		Result: s.marshalJSON(result),
	}

}
