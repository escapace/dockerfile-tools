package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/samber/lo"
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

func parseMountOptions(flag string, argMap map[string]string) map[string]string {
	// Match key-value pairs, considering quoted values
	regex := regexp.MustCompile(`([^=,\s]+)=((?:"[^"]*")|(?:[^",]+))`)
	matches := regex.FindAllStringSubmatch(flag, -1)

	options := map[string]string{}
	for _, match := range matches {
		key := match[1]
		if key == "--mount" {
			continue // Skip the --mount key
		}

		value := strings.Trim(match[2], `"`) // Remove quotes if present

		// Replace $ARG variables with provided values
		for argKey, argValue := range argMap {
			value = strings.ReplaceAll(value, fmt.Sprintf("$%s", argKey), argValue)
		}

		options[key] = value
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
	data := map[string]map[string]string{}

	// Traverse the AST to find RUN instructions with --mount=type=cache
	for _, child := range result.AST.Children {
		if child.Value == "RUN" && child.Flags != nil {
			for _, flag := range child.Flags {
				if strings.Contains(flag, "--mount=type=cache") {
					// Extract mount options and parse them
					options := parseMountOptions(flag, argMap)
					if target, ok := options["target"]; ok {
						// Convert the target to kebab-case for the key
						key := lo.KebabCase(target)
						data[key] = options
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
