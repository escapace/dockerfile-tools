package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/samber/lo"
	"github.com/wk8/go-ordered-map/v2"
)

func parseArgs(args []string) map[string]string {
	argMap := map[string]string{}

	argMap["BUILDOS"] = runtime.GOOS
	argMap["BUILDARCH"] = runtime.GOARCH

	for _, arg := range args {
		pairs := strings.Split(arg, ",")
		for _, pair := range pairs {
			if eqIdx := strings.Index(pair, "="); eqIdx != -1 {
				key := pair[:eqIdx]
				value := pair[eqIdx+1:]
				argMap[key] = value
			}
		}
	}

	if _, exists := argMap["BUILDPLATFORM"]; !exists {
		argMap["BUILDPLATFORM"] = runtime.GOOS + "/" + runtime.GOARCH
	}

	return argMap
}

func splitCommaSeparated(input string) []string {
	parts := []string{}
	var current strings.Builder
	inQuotes := false

	for _, char := range input {
		switch char {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(char)
		case ',':
			if inQuotes {
				current.WriteRune(char)
				continue
			}
			parts = append(parts, current.String())
			current.Reset()
		default:
			current.WriteRune(char)
		}
	}

	parts = append(parts, current.String())
	return parts
}

func substituteArgs(value string, argMap map[string]string) string {
	for argKey, argValue := range argMap {
		value = strings.ReplaceAll(value, fmt.Sprintf("${%s}", argKey), argValue)
		value = strings.ReplaceAll(value, fmt.Sprintf("$%s", argKey), argValue)
	}

	return value
}

func parseMountOptions(flag string, argMap map[string]string) *orderedmap.OrderedMap[string, string] {
	options := orderedmap.New[string, string]()

	mountDefinition, found := strings.CutPrefix(flag, "--mount=")
	if !found {
		return options
	}

	for _, part := range splitCommaSeparated(mountDefinition) {
		key, value, hasValue := strings.Cut(part, "=")
		if !hasValue {
			continue
		}

		options.Set(key, substituteArgs(strings.Trim(value, `"`), argMap))
	}

	return options
}

func ListCacheMounts(dockerfilePath string, args []string) {
	// Parse the provided ARGs
	argMap := parseArgs(args)

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
	data := orderedmap.New[string, *orderedmap.OrderedMap[string, string]]()

	// Traverse the AST to find RUN instructions with cache mounts.
	for _, child := range result.AST.Children {
		if child.Value != "RUN" || child.Flags == nil {
			continue
		}

		for _, flag := range child.Flags {
			if !strings.HasPrefix(flag, "--mount=") {
				continue
			}

			options := parseMountOptions(flag, argMap)
			mountType, exists := options.Get("type")
			if !exists || mountType != "cache" {
				continue
			}

			var key string

			if id, exists := options.Get("id"); exists {
				// Use the kebab-case of the "id" value as the key
				key = lo.KebabCase(id)
			} else if target, exists := options.Get("target"); exists {
				// Use the kebab-case of the "target" value as the fallback key
				key = lo.KebabCase(target)
			} else {
				// Skip if neither "id" nor "target" is present
				continue
			}

			data.Set(".cache-"+key, options)
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
