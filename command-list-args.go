package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

type ArgDeclaration struct {
	Name         string  `json:"name"`
	DefaultValue *string `json:"defaultValue"`
	Line         int     `json:"line"`
}

func parseArgDeclaration(spec string, line int) ArgDeclaration {
	name, defaultValue, hasDefault := strings.Cut(spec, "=")
	declaration := ArgDeclaration{
		Name: strings.TrimSpace(name),
		Line: line,
	}
	if hasDefault {
		declaration.DefaultValue = &defaultValue
	}

	return declaration
}

func extractArgDeclarations(ast *parser.Node) []ArgDeclaration {
	declarations := []ArgDeclaration{}

	for _, child := range ast.Children {
		if child.Value != "ARG" {
			continue
		}

		for current := child.Next; current != nil; current = current.Next {
			declarations = append(declarations, parseArgDeclaration(current.Value, child.StartLine))
		}
	}

	return declarations
}

func ListArgs(dockerfilePath string) {
	absPath, err := filepath.Abs(dockerfilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	file, err := os.Open(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening Dockerfile at '%s': %v\n", absPath, err)
		os.Exit(1)
	}
	defer file.Close()

	result, err := parser.Parse(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing Dockerfile: %v\n", err)
		os.Exit(1)
	}

	declarationsJSON, err := json.MarshalIndent(extractArgDeclarations(result.AST), "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting ARG declarations to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(declarationsJSON))
}
