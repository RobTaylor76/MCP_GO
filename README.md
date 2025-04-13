# Go MCP Web Server

This project implements a web server with Model Context Protocol (MCP) support based on the [2025-03-26 specification](https://modelcontextprotocol.io/specification/2025-03-26). The server provides both standard HTTP endpoints and MCP protocol functionality.

## Project Structure 

## Features

### Web Server
- Basic HTTP endpoints (`/` and `/health`)
- JSON response utilities
- Standard HTTP handler structure

### MCP Implementation
1. **Base Protocol Support**
   - JSON-RPC 2.0 message handling
   - Session management
   - Server-Sent Events (SSE) support
   - Capability negotiation

2. **Tool Support**
   - Tool listing endpoint (`tools/list`)
   - Tool execution endpoint (`tools/call`)
   - Sample weather tool implementation
   ```json
   {
     "name": "get_weather",
     "description": "Get current weather information for a location",
     "inputSchema": {
       "type": "object",
       "properties": {
         "location": {
           "type": "string",
           "description": "City name or zip code"
         }
       },
       "required": ["location"]
     }
   }
   ```

3. **Real-time Features**
   - SSE-based streaming support
   - Client connection management
   - Keepalive mechanism
   - Notification broadcasting

## Getting Started

1. Initialize the Go module:
```bash
go mod init github.com/rob/go-web-server
go mod tidy
```

2. Run the server:
```bash
go run main.go
```

The server will start on port 8080.

## API Usage

### Standard Web Endpoints

1. Home endpoint:
```bash
curl http://localhost:8080/
```

2. Health check:
```bash
curl http://localhost:8080/health
```

### MCP Protocol Endpoints

1. Initialize an MCP session:
```bash
curl -v -X POST http://localhost:8080/mcp \
  -H "X-API-Key: sample-key" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize"}'
```

2. List available tools:
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Mcp-Session-Id: ${SESSION_KEY}" \
  -H "X-API-Key: sample-key" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
```

3. Call the weather tool:
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Mcp-Session-Id: ${SESSION_KEY}" \
  -H "X-API-Key: sample-key" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_weather","arguments":{"location":"New York"}}}'
```

4. Establish SSE connection:
```bash
curl -N http://localhost:8080/mcp \
  -H "Mcp-Session-Id: <session-id>"
```

## Testing SSE Functionality

To test the Server-Sent Events (SSE) functionality:

1. Initialize a session:
```bash
curl -v -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize"}'
```

2. Establish an SSE connection using the session ID from step 1:
```bash
curl -N http://localhost:8080/mcp \
  -H "Mcp-Session-Id: ${SESSION_KEY}" \
  -H "X-API-Key: test"
```

You should observe:
- An initial "connected" event
- Keepalive messages every 30 seconds
- Any notifications sent via `broadcastToSession`

3. Test notifications by calling the weather tool in another terminal:
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Mcp-Session-Id: ${SESSION_KEY}" \
  -H "X-API-Key: test" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_weather","arguments":{"location":"New York"}}}'
```

The SSE connection will display these events in real-time.

### Expected SSE Output

## Implementation Details

### Session Management
- Secure session IDs using UUID v4
- Session-scoped message channels for SSE
- Automatic session cleanup on client disconnect

### SSE Implementation
- Keepalive messages every 30 seconds
- Proper header configuration
- Graceful connection handling
- Support for server-to-client notifications

### Security Features
- Origin validation support
- Session validation
- Tool input validation
- Error handling for invalid requests

## Development History

1. Initial project setup with basic web server
2. Added MCP protocol base implementation
3. Implemented tool support with weather example
4. Added SSE support for real-time communication
5. Added documentation and examples

## Future Improvements

1. Add real weather API integration
2. Implement rate limiting
3. Add comprehensive input validation
4. Add proper error handling for API failures
5. Implement proper logging and monitoring
6. Add authentication and authorization
7. Add more tools and capabilities

## References

- [MCP Specification](https://modelcontextprotocol.io/specification/2025-03-26)
- [MCP Tools Specification](https://modelcontextprotocol.io/specification/2025-03-26/server/tools)

## Cursor Integration

To use this MCP server as a tool in Cursor:

1. Start the MCP server:
```bash
go run main.go
```

2. In Cursor, open the Command Palette (Cmd/Ctrl + Shift + P)

3. Select "Tools: Add New Tool"

4. Choose "Add MCP Tool"

5. Enter the path to your `cursor-tool.json` file

6. The weather tool will now be available in Cursor's tool palette

### Using the Weather Tool in Cursor

1. Open the Command Palette
2. Type "Get Weather"
3. Enter a location when prompted
4. The weather information will be displayed in your current context

### Troubleshooting

If the tool is not appearing in Cursor:
1. Ensure the MCP server is running on port 8080
2. Check that the `cursor-tool.json` file is properly formatted
3. Verify that Cursor has the correct permissions to access the local server
4. Try restarting Cursor after adding the tool

### Security Note

When using this tool in Cursor:
- The server runs locally on port 8080
- All communication is unencrypted (HTTP)
- For production use, consider adding proper authentication and HTTPS support 

## Backward Compatibility

The server supports both the new Streamable HTTP transport and the legacy HTTP+SSE transport:

1. New Streamable HTTP transport (recommended):
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize"}'
```

2. Legacy HTTP+SSE transport:
```bash
# First, connect to the SSE endpoint
curl -N http://localhost:8080/sse \
  -H "X-API-Key: test"

# Then use the provided endpoint for subsequent requests
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize"}'
```

The legacy transport will:
1. Return an `endpoint` event with the new MCP endpoint URL
2. Maintain an SSE connection for server-to-client messages
3. Support all the same functionality as the new transport 