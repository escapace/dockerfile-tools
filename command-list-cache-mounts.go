package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
  "github.com/iancoleman/strcase"
)

func ListCacheMounts(dockerfilePath string) {
	// Resolve the absolute path to the Dockerfile
	absPath, err := filepath.Abs(dockerfilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	// Open the Dockerfile
	file, err := os.Open(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening Dockerfile at '%s': %v\n", absPath, err)
		os.Exit(1)
	}
	defer file.Close()

	// Parse the Dockerfile
	result, err := parser.Parse(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing Dockerfile: %v\n", err)
		os.Exit(1)
	}

	// Initialize the map to hold cache mount data
	data := map[string]string{}

	// Traverse the AST to find RUN instructions with --mount=type=cache
	for _, child := range result.AST.Children {
		if child.Value == "RUN" && child.Flags != nil {
			for _, flag := range child.Flags {
				if strings.Contains(flag, "--mount=type=cache") {
					// Extract the target path
					targetIdx := strings.Index(flag, "target=")
					if targetIdx != -1 {
						target := flag[targetIdx+len("target="):]
						// Split by comma to isolate the target value
						if commaIdx := strings.Index(target, ","); commaIdx != -1 {
							target = target[:commaIdx]
						}
						// Convert the target to kebab-case and add to the map
						data[strcase.ToKebab(strings.ReplaceAll(target, "/", " "))] = target
					}
				}
			}
		}
	}

	// Convert the result map to JSON
	cacheJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting cache mounts to JSON: %v\n", err)
		os.Exit(1)
	}

	// Output the JSON
	fmt.Println(string(cacheJSON))
}
