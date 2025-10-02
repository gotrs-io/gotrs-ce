package main

import (
	"fmt"
	"github.com/gotrs-io/gotrs-ce/internal/api"
)

func main() {
	// Test table markdown
	tableMarkdown := `| Method | Time (10 modules) | Speed |
|--------|------------------|-------|
| Manual YAML Writing | 150 minutes | Baseline |
| Schema Discovery | 0.33 seconds | **9000x faster** |`

	result := api.RenderMarkdown(tableMarkdown)
	fmt.Println("Input:")
	fmt.Println(tableMarkdown)
	fmt.Println("\nOutput:")
	fmt.Println(result)
}