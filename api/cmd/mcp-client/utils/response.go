package utils

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PrintToolResponse prints the tool response in a formatted way
func PrintToolResponse(res *mcp.CallToolResult) {
	if res.IsError {
		fmt.Printf("Tool returned an error:\n")
		if len(res.Content) > 0 {
			for _, c := range res.Content {
				if textContent, ok := c.(*mcp.TextContent); ok {
					fmt.Printf("  Error: %s\n", textContent.Text)
				}
			}
		}
		return
	}

	fmt.Println("Tool call successful:")
	for _, c := range res.Content {
		switch content := c.(type) {
		case *mcp.TextContent:
			fmt.Printf("  Text: %s\n", content.Text)
		case *mcp.ImageContent:
			fmt.Printf("  Image: %d bytes (mime: %s)\n", len(content.Data), content.MIMEType)
		case *mcp.AudioContent:
			fmt.Printf("  Audio: %d bytes (mime: %s)\n", len(content.Data), content.MIMEType)
		default:
			// Try to marshal unknown content types
			jsonBytes, err := json.MarshalIndent(content, "  ", "  ")
			if err == nil {
				fmt.Printf("  Content: %s\n", string(jsonBytes))
			} else {
				fmt.Printf("  Content: %+v\n", content)
			}
		}
	}

	// Also check structured content if available
	if res.StructuredContent != nil {
		fmt.Println("\nStructured Content:")
		jsonBytes, err := json.MarshalIndent(res.StructuredContent, "  ", "  ")
		if err == nil {
			fmt.Printf("%s\n", string(jsonBytes))
		} else {
			fmt.Printf("%+v\n", res.StructuredContent)
		}
	}
}
