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

func parseMountOptions(flag string, argMap map[string]string) *orderedmap.OrderedMap[string, string] {
	// Match key-value pairs, considering quoted values
	regex := regexp.MustCompile(`([^=,\s]+)=((?:"[^"]*")|(?:[^",]+))`)
	matches := regex.FindAllStringSubmatch(flag, -1)

	options := orderedmap.New[string, string]()
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

		options.Set(key, value)
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

	// Traverse the AST to find RUN instructions with --mount=type=cache
	for _, child := range result.AST.Children {
		if child.Value == "RUN" && child.Flags != nil {
			for _, flag := range child.Flags {
				if strings.Contains(flag, "--mount=type=cache") {
					// Extract mount options and parse them
					options := parseMountOptions(flag, argMap)

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
