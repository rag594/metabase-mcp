# Metabase MCP Server

A Model Context Protocol (MCP) server that enables querying Metabase databases through natural language interfaces and AI assistants.

## Overview

This MCP server provides a bridge between AI assistants (like Claude) and Metabase, allowing you to execute SQL queries against your Metabase databases programmatically. It's built in Go and implements the MCP protocol for seamless integration with compatible clients.

## Features

- Execute native SQL queries against Metabase databases
- Structured response formatting with query metadata
- Cookie-based authentication support
- Configurable database selection
- Error handling and timeout management
- Compatible with VS Code MCP extension and other MCP clients

## Prerequisites

- Go 1.19 or later
- Access to a Metabase instance
- Valid Metabase session cookies
- VS Code with MCP extension (for VS Code integration)

## Installation

### 1. Clone and Build

```bash
git clone <repository-url>
cd metabase-mcp
go mod tidy
go build -o metabase-mcp
```

### 2. Get Metabase Authentication Cookies

1. Open your Metabase instance in a web browser
2. Log in to your account
3. Open browser developer tools (F12)
4. Go to the Network tab
5. Make any request to Metabase
6. Copy the `Cookie` header value from the request

The cookies should include at minimum:
- `metabase.SESSION=<session-id>`
- Any additional authentication cookies

### 3. Find Your Database ID

1. In Metabase, go to Admin → Databases
2. Click on your target database
3. Note the database ID from the URL

## Configuration

### VS Code Setup

Create or update `.vscode/mcp.json` in your workspace:

```json
{
  "servers": {
    "metabase-mcp-server": {
      "type": "stdio",
      "command": "/path/to/your/metabase-mcp/metabase-mcp",
      "args": [],
      "env": {
        "METABASE_DATABASE_ID": "1",
        "METABASE_HOST": "https://your-metabase-instance.com",
        "METABASE_COOKIES": "your-complete-cookie-string-here"
      }
    }
  }
}
```

### Environment Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `METABASE_DATABASE_ID` | Target database ID in Metabase | Yes | `1` |
| `METABASE_HOST` | Metabase instance URL | Yes | `https://metabase.example.com` |
| `METABASE_COOKIES` | Authentication cookies | Yes | `metabase.SESSION=abc123;...` |

## Usage

### VS Code Integration

1. Install the MCP extension in VS Code
2. Configure the server in `.vscode/mcp.json` as shown above
3. Restart VS Code
4. Use the MCP tool in your AI assistant conversations

Example queries you can ask:
```
"What's the table size of VKYCRequest?"
"Show me the top 10 users by transaction count"
"Get the latest records from the payments table"
```

### Direct Usage

You can also run the server directly:

```bash
export METABASE_DATABASE_ID=20
export METABASE_HOST=https://your-metabase-instance.com
export METABASE_COOKIES="your-cookie-string"
./metabase-mcp
```

## API Reference

### Tool: metabase-tool

**Description**: Execute SQL queries against the configured Metabase database

**Parameters**:
- `query` (string, required): The SQL query to execute

**Example**:
```json
{
  "name": "metabase-tool",
  "arguments": {
    "query": "SELECT COUNT(*) as total_records FROM users"
  }
}
```

**Response Format**:
```json
{
  "status": "completed",
  "row_count": 1,
  "running_time": 15,
  "database_id": 31,
  "cached": false,
  "rows": [[1000]],
  "columns": [
    {
      "name": "total_records",
      "display_name": "total_records",
      "base_type": "type/BigInteger"
    }
  ],
  "query_sent": {
    "type": "native",
    "database": 31,
    "native": {
      "query": "SELECT COUNT(*) as total_records FROM users"
    }
  }
}
```

## Troubleshooting

### Common Issues

1. **TLS Handshake Timeout**
   - Check your network connectivity to the Metabase host
   - Verify the METABASE_HOST URL is correct
   - Ensure your firewall allows outbound HTTPS connections

2. **404 Not Found / API endpoint does not exist**
   - Your session cookies have expired - refresh them
   - Check if you have access to the specified database ID
   - Verify the METABASE_HOST URL is correct

3. **Authentication Errors**
   - Update your METABASE_COOKIES with fresh session cookies
   - Ensure you're logged into the correct Metabase account
   - Check if your account has query permissions

4. **Database Not Found**
   - Verify the METABASE_DATABASE_ID is correct
   - Check if you have access to that database in Metabase
   - Confirm the database is active and connected

### Debug Mode

To enable debug logging, you can modify the server to include more verbose output or check the Metabase logs for additional error details.

### Cookie Refresh

Session cookies typically expire after some time. To refresh:
1. Clear your browser cache for the Metabase domain
2. Log out and log back into Metabase
3. Extract new cookies using browser developer tools
4. Update your `.vscode/mcp.json` configuration
5. Restart VS Code

## Security Considerations

- Keep your session cookies secure and don't commit them to version control
- Use environment variables or secure configuration management for production
- Regularly rotate session cookies
- Limit database permissions to only what's necessary for your queries
- Consider using Metabase API keys instead of session cookies for production use

## Development

### Project Structure

```
metabase-mcp/
├── main.go              # Main server implementation
├── go.mod               # Go module dependencies
├── go.sum               # Go module checksums
├── metabase-mcp         # Compiled binary
└── .vscode/
    └── mcp.json         # VS Code MCP configuration
```

### Building from Source

```bash
go mod tidy
go build -o metabase-mcp main.go
```

### Dependencies

- `github.com/mark3labs/mcp-go/mcp` - MCP protocol implementation
- `github.com/mark3labs/mcp-go/server` - MCP server framework

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test with your Metabase instance
5. Submit a pull request

## License

[Add your license information here]

## Support

For issues and questions:
1. Check the troubleshooting section above
2. Verify your Metabase instance is accessible
3. Test your SQL queries directly in Metabase first
4. Open an issue with detailed error messages and configuration (excluding sensitive data)