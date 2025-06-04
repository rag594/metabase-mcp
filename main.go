package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MetabaseQuery represents a Metabase query structure
type MetabaseQuery struct {
	Type       string        `json:"type"`
	Database   int           `json:"database"`
	Native     NativeQuery   `json:"native"`
	Parameters []interface{} `json:"parameters"`
}

// NativeQuery represents the native query part of a Metabase query
type NativeQuery struct {
	Query        string                 `json:"query"`
	TemplateTags map[string]interface{} `json:"template-tags"`
}

// MetabaseResponse represents the complete response from Metabase API
type MetabaseResponse struct {
	Data                 MetabaseData `json:"data"`
	Cached               bool         `json:"cached"`
	DatabaseID           int          `json:"database_id"`
	StartedAt            string       `json:"started_at"`
	JSONQuery            JSONQuery    `json:"json_query"`
	AverageExecutionTime *float64     `json:"average_execution_time"`
	Status               string       `json:"status"`
	Context              string       `json:"context"`
	RowCount             int          `json:"row_count"`
	RunningTime          int          `json:"running_time"`
}

// MetabaseData represents the data section of the response
type MetabaseData struct {
	Rows            [][]interface{} `json:"rows"`
	Cols            []Column        `json:"cols"`
	NativeForm      NativeForm      `json:"native_form"`
	ResultsTimezone string          `json:"results_timezone"`
	ResultsMetadata ResultsMetadata `json:"results_metadata"`
	Insights        *interface{}    `json:"insights"`
}

// Column represents a column definition in the response
type Column struct {
	DisplayName   string        `json:"display_name"`
	Source        string        `json:"source"`
	FieldRef      []interface{} `json:"field_ref"`
	Name          string        `json:"name"`
	BaseType      string        `json:"base_type"`
	EffectiveType string        `json:"effective_type"`
}

// NativeForm represents the native form of the executed query
type NativeForm struct {
	Query  string      `json:"query"`
	Params interface{} `json:"params"`
}

// ResultsMetadata contains metadata about the query results
type ResultsMetadata struct {
	Columns []MetadataColumn `json:"columns"`
}

// MetadataColumn represents detailed column metadata
type MetadataColumn struct {
	DisplayName   string        `json:"display_name"`
	FieldRef      []interface{} `json:"field_ref"`
	Name          string        `json:"name"`
	BaseType      string        `json:"base_type"`
	EffectiveType string        `json:"effective_type"`
	SemanticType  *string       `json:"semantic_type"`
	Fingerprint   *Fingerprint  `json:"fingerprint"`
}

// Fingerprint represents column fingerprint data
type Fingerprint struct {
	Global GlobalFingerprint          `json:"global"`
	Type   map[string]TypeFingerprint `json:"type"`
}

// GlobalFingerprint represents global fingerprint statistics
type GlobalFingerprint struct {
	DistinctCount int     `json:"distinct-count"`
	NilPercent    float64 `json:"nil%"`
}

// TypeFingerprint represents type-specific fingerprint data
type TypeFingerprint struct {
	PercentJSON   float64 `json:"percent-json"`
	PercentURL    float64 `json:"percent-url"`
	PercentEmail  float64 `json:"percent-email"`
	PercentState  float64 `json:"percent-state"`
	AverageLength float64 `json:"average-length"`
}

// JSONQuery represents the JSON query that was executed
type JSONQuery struct {
	Type       string                 `json:"type"`
	Database   int                    `json:"database"`
	Native     NativeQuery            `json:"native"`
	Middleware map[string]interface{} `json:"middleware"`
}

func main() {
	fmt.Println("Metabase MCP Server starting...")

	// Get database ID from environment variable
	var databaseID int
	dbEnv := os.Getenv("METABASE_DATABASE_ID")
	if dbEnv == "" {
		log.Fatalln("Database ID not set or invalid")
	}

	if parsedDB, err := strconv.Atoi(dbEnv); err == nil {
		databaseID = parsedDB
	}

	// Get authentication cookies from environment variable
	cookies := os.Getenv("METABASE_COOKIES")
	if cookies == "" {
		log.Fatalln("METABASE_COOKIES not set")
	}

	// Get Metabase URL from environment variable
	metabaseHost := os.Getenv("METABASE_HOST")
	if metabaseHost == "" {
		log.Fatalln("METABASE_HOST is not set")
	}

	// Create a new MCP server
	s := server.NewMCPServer(
		"metabase-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	// Add API invocation tool
	apiTool := mcp.NewTool(
		"metabase-tool",
		mcp.WithDescription("Metabase mcp can access dashboards, execute queries"),
		mcp.WithString(
			"query",
			mcp.Required(),
			mcp.Description("The query to execute against the the db"),
		),
	)

	// Add API tool handler
	s.AddTool(apiTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Convert arguments to map[string]interface{}
		arguments, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments format"), nil
		}

		// Extract query (required)
		query, ok := arguments["query"].(string)
		if !ok || query == "" {
			return mcp.NewToolResultError("query is required and must be a string"), nil
		}

		// Create MetabaseQuery struct with the provided query
		metabaseQuery := MetabaseQuery{
			Type:     "native",
			Database: databaseID,
			Native: NativeQuery{
				Query:        query,
				TemplateTags: make(map[string]interface{}),
			},
			Parameters: make([]interface{}, 0),
		}

		// Convert the query struct to JSON
		queryJSON, err := json.Marshal(metabaseQuery)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create query JSON: %v", err)), nil
		}

		// Extract timeout (optional, defaults to 120 seconds)
		timeout := 120 * time.Second

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: timeout,
		}

		// Create request with the query JSON as body
		reqBody := strings.NewReader(string(queryJSON))
		metabaseURL := fmt.Sprintf("%s/api/dataset", metabaseHost)
		fmt.Println(metabaseURL)
		req, err := http.NewRequestWithContext(ctx, "POST", metabaseURL, reqBody)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create request: %v", err)), nil
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookies)

		// Make the request
		resp, err := client.Do(req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("request failed: %v", err)), nil
		}
		defer resp.Body.Close()

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to read response: %v", err)), nil
		}

		// Try to parse the response into the MetabaseResponse struct
		var metabaseResp MetabaseResponse
		if err := json.Unmarshal(respBody, &metabaseResp); err == nil {
			// Successfully parsed as MetabaseResponse, format nicely
			formattedResponse := map[string]interface{}{
				"status":       metabaseResp.Status,
				"row_count":    metabaseResp.RowCount,
				"running_time": metabaseResp.RunningTime,
				"database_id":  metabaseResp.DatabaseID,
				"cached":       metabaseResp.Cached,
				"rows":         metabaseResp.Data.Rows,
				"columns":      metabaseResp.Data.Cols,
				"query_sent":   metabaseQuery,
			}

			responseJSON, err := json.MarshalIndent(formattedResponse, "", "  ")
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
			}
			return mcp.NewToolResultText(string(responseJSON)), nil
		}

		// Fallback: if parsing as MetabaseResponse fails, return raw response
		response := map[string]interface{}{
			"status_code": resp.StatusCode,
			"status":      resp.Status,
			"body":        string(respBody),
			"query_sent":  metabaseQuery,
		}

		responseJSON, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to format response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(responseJSON)), nil
	})

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
