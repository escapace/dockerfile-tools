package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

// GenerateAST generates a JSON AST from a Dockerfile
func AST(dockerfilePath string) {
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

	// Convert the AST to JSON
	astJSON, err := json.MarshalIndent(result.AST, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting AST to JSON: %v\n", err)
		os.Exit(1)
	}

	// Print the JSON representation of the AST to stdout
	fmt.Println(string(astJSON))
}
