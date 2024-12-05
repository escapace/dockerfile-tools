package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

// extractStageNames traverses the AST to find all nodes matching the shape:
// { "Value": "FROM", "Next": { "Value": string, "Next": { "Value": "AS", "Next": { "Value": STAGE_NAME } } } }
func extractStageNames(ast *parser.Node) []string {
	var stageNames []string

	// Traverse all top-level children of the AST
	for _, child := range ast.Children {
		if child.Value == "FROM" {
			// Check for the "Next" field (base image or alias chain)
			next := child.Next
			if next != nil {
				// Traverse the "Next" chain to find "AS" and its alias
				for current := next; current != nil; current = current.Next {
					if current.Value == "AS" && current.Next != nil {
						stageNames = append(stageNames, current.Next.Value)
						break
					}
				}
			}
		}
	}

	return stageNames
}

// ListStages extracts and lists the build stages from a Dockerfile
func ListStages(dockerfilePath string) {
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

	// Extract stage names
	stageNames := extractStageNames(result.AST)

	// Convert stage names to JSON and print to stdout
	stageNamesJSON, err := json.MarshalIndent(stageNames, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting stage names to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(stageNamesJSON))
}
