package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func writeTempDockerfile(t *testing.T, contents string) string {
	t.Helper()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "Dockerfile")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write Dockerfile: %v", err)
	}

	return path
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	os.Stdout = writer
	defer func() {
		os.Stdout = originalStdout
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	return string(output)
}

func parseDockerfileAST(t *testing.T, contents string) *parser.Node {
	t.Helper()

	path := writeTempDockerfile(t, contents)
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open Dockerfile: %v", err)
	}
	defer file.Close()

	result, err := parser.Parse(file)
	if err != nil {
		t.Fatalf("parse Dockerfile: %v", err)
	}

	return result.AST
}

func runMainProcess(t *testing.T, args ...string) (string, int) {
	t.Helper()

	cmdArgs := append([]string{"-test.run=TestHelperProcess", "--"}, args...)
	cmd := exec.Command(os.Args[0], cmdArgs...)
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")

	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), 0
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return string(output), exitErr.ExitCode()
	}

	t.Fatalf("run helper process: %v", err)
	return "", 0
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	separatorIndex := -1
	for index, arg := range os.Args {
		if arg == "--" {
			separatorIndex = index
			break
		}
	}
	if separatorIndex == -1 {
		os.Exit(2)
	}

	os.Args = append([]string{os.Args[0]}, os.Args[separatorIndex+1:]...)
	main()
	os.Exit(0)
}

func TestParseArgsDefaultsAndOverrides(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		got := parseArgs(nil)

		if got["BUILDOS"] != runtime.GOOS {
			t.Fatalf("BUILDOS = %q, want %q", got["BUILDOS"], runtime.GOOS)
		}
		if got["BUILDARCH"] != runtime.GOARCH {
			t.Fatalf("BUILDARCH = %q, want %q", got["BUILDARCH"], runtime.GOARCH)
		}
		if got["BUILDPLATFORM"] != runtime.GOOS+"/"+runtime.GOARCH {
			t.Fatalf("BUILDPLATFORM = %q, want %q", got["BUILDPLATFORM"], runtime.GOOS+"/"+runtime.GOARCH)
		}
	})

	t.Run("overrides and comma-delimited pairs", func(t *testing.T) {
		got := parseArgs([]string{
			"BUILDOS=linux,BUILDARCH=arm64",
			"BUILDPLATFORM=linux/arm64,GO_VERSION=1.26.2",
		})

		want := map[string]string{
			"BUILDOS":       "linux",
			"BUILDARCH":     "arm64",
			"BUILDPLATFORM": "linux/arm64",
			"GO_VERSION":    "1.26.2",
		}

		for key, wantValue := range want {
			if got[key] != wantValue {
				t.Fatalf("%s = %q, want %q", key, got[key], wantValue)
			}
		}
	})
}

func TestParseMountOptionsParsesQuotedValuesAndSubstitutesArgs(t *testing.T) {
	options := parseMountOptions(
		`--mount=type=cache,target="/root/.cache/${BUILDPLATFORM}",id="go-$BUILDPLATFORM",sharing=locked`,
		map[string]string{"BUILDPLATFORM": "linux/arm64"},
	)

	want := map[string]string{
		"type":    "cache",
		"target":  "/root/.cache/linux/arm64",
		"id":      "go-linux/arm64",
		"sharing": "locked",
	}

	for key, wantValue := range want {
		gotValue, ok := options.Get(key)
		if !ok {
			t.Fatalf("missing option %q", key)
		}
		if gotValue != wantValue {
			t.Fatalf("option %s = %q, want %q", key, gotValue, wantValue)
		}
	}
}

func TestExtractStageNamesReturnsAliasedStagesOnly(t *testing.T) {
	ast := parseDockerfileAST(t, `FROM golang:1.26 AS build
FROM build AS test
FROM scratch
`)

	got := extractStageNames(ast)
	want := []string{"build", "test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractStageNames() = %#v, want %#v", got, want)
	}
}

func TestASTPrintsJSON(t *testing.T) {
	dockerfile := writeTempDockerfile(t, `# syntax=docker/dockerfile:1
ARG GO_VERSION=1.26
FROM golang:${GO_VERSION} AS build
RUN go build ./...
`)

	output := captureStdout(t, func() {
		AST(dockerfile)
	})

	var ast struct {
		Value    string `json:"Value"`
		Children []struct {
			Value    string `json:"Value"`
			Original string `json:"Original"`
		} `json:"Children"`
	}
	if err := json.Unmarshal([]byte(output), &ast); err != nil {
		t.Fatalf("unmarshal AST JSON: %v\noutput: %s", err, output)
	}

	if ast.Value != "" {
		t.Fatalf("root Value = %q, want empty string", ast.Value)
	}
	if len(ast.Children) != 3 {
		t.Fatalf("len(Children) = %d, want 3", len(ast.Children))
	}
	if ast.Children[1].Value != "FROM" {
		t.Fatalf("second child Value = %q, want %q", ast.Children[1].Value, "FROM")
	}
	if ast.Children[1].Original != "FROM golang:${GO_VERSION} AS build" {
		t.Fatalf("second child Original = %q", ast.Children[1].Original)
	}
}

func TestListStagesPrintsJSON(t *testing.T) {
	dockerfile := writeTempDockerfile(t, `FROM golang:1.26 AS build
FROM build AS test
FROM scratch
`)

	output := captureStdout(t, func() {
		ListStages(dockerfile)
	})

	var got []string
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("unmarshal stages JSON: %v\noutput: %s", err, output)
	}

	want := []string{"build", "test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ListStages() output = %#v, want %#v", got, want)
	}
}

func TestListCacheMountsPrintsJSON(t *testing.T) {
	dockerfile := writeTempDockerfile(t, `# syntax=docker/dockerfile:1
FROM golang:1.26 AS build
RUN --mount=target="/root/.cache/${BUILDPLATFORM}",id="go-${BUILDPLATFORM}",type=cache go build ./...
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked go test ./...
RUN --mount=type=bind,target=/src echo not-a-cache-mount
`)

	output := captureStdout(t, func() {
		ListCacheMounts(dockerfile, []string{"BUILDPLATFORM=linux/arm64"})
	})

	var got map[string]map[string]string
	if err := json.Unmarshal([]byte(output), &got); err != nil {
		t.Fatalf("unmarshal cache mounts JSON: %v\noutput: %s", err, output)
	}

	want := map[string]map[string]string{
		".cache-go-linux-arm-64": {
			"type":   "cache",
			"target": "/root/.cache/linux/arm64",
			"id":     "go-linux/arm64",
		},
		".cache-go-pkg-mod": {
			"type":    "cache",
			"target":  "/go/pkg/mod",
			"sharing": "locked",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ListCacheMounts() output = %#v, want %#v", got, want)
	}
}

func TestPrintHelpIncludesAllCommands(t *testing.T) {
	output := captureStdout(t, printHelp)

	for _, command := range []string{"ast", "list-stages", "list-cache-mounts"} {
		if !strings.Contains(output, command) {
			t.Fatalf("help output missing %q\noutput: %s", command, output)
		}
	}
}

func TestMainWithoutArgsPrintsHelp(t *testing.T) {
	output, exitCode := runMainProcess(t)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0\noutput: %s", exitCode, output)
	}
	if !strings.Contains(output, "dockerfile-tools <command> [options]") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestMainInvalidSubcommandExitsOne(t *testing.T) {
	output, exitCode := runMainProcess(t, "unknown")
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1\noutput: %s", exitCode, output)
	}
	if !strings.Contains(output, "Error: expected 'ast', 'list-stages', or 'list-cache-mounts' subcommands") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestMainSubcommandHelp(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		wantOutputPart string
	}{
		{
			name:           "ast",
			args:           []string{"ast", "--help"},
			wantOutputPart: "dockerfile-tools ast [options]",
		},
		{
			name:           "list-stages",
			args:           []string{"list-stages", "--help"},
			wantOutputPart: "dockerfile-tools list-stages [options]",
		},
		{
			name:           "list-cache-mounts",
			args:           []string{"list-cache-mounts", "--help"},
			wantOutputPart: "dockerfile-tools list-cache-mounts [options]",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			output, exitCode := runMainProcess(t, testCase.args...)
			if exitCode != 0 {
				t.Fatalf("exit code = %d, want 0\noutput: %s", exitCode, output)
			}
			if !strings.Contains(output, testCase.wantOutputPart) {
				t.Fatalf("unexpected output: %s", output)
			}
		})
	}
}

func TestMainRequiresDockerfileArgument(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{name: "ast", args: []string{"ast"}},
		{name: "list-stages", args: []string{"list-stages"}},
		{name: "list-cache-mounts", args: []string{"list-cache-mounts"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			output, exitCode := runMainProcess(t, testCase.args...)
			if exitCode != 1 {
				t.Fatalf("exit code = %d, want 1\noutput: %s", exitCode, output)
			}
			if !strings.Contains(output, "Please provide a path to the Dockerfile using --dockerfile") {
				t.Fatalf("unexpected output: %s", output)
			}
		})
	}
}

func TestMainCommandsSucceed(t *testing.T) {
	dockerfile := writeTempDockerfile(t, `# syntax=docker/dockerfile:1
FROM golang:1.26 AS build
RUN --mount=type=cache,target=/go/pkg/mod go test ./...
`)

	testCases := []struct {
		name           string
		args           []string
		wantOutputPart string
	}{
		{
			name:           "ast",
			args:           []string{"ast", "--dockerfile", dockerfile},
			wantOutputPart: `"Value": "FROM"`,
		},
		{
			name:           "list-stages",
			args:           []string{"list-stages", "--dockerfile", dockerfile},
			wantOutputPart: `"build"`,
		},
		{
			name:           "list-cache-mounts",
			args:           []string{"list-cache-mounts", "--dockerfile", dockerfile},
			wantOutputPart: `".cache-go-pkg-mod"`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			output, exitCode := runMainProcess(t, testCase.args...)
			if exitCode != 0 {
				t.Fatalf("exit code = %d, want 0\noutput: %s", exitCode, output)
			}
			if !strings.Contains(output, testCase.wantOutputPart) {
				t.Fatalf("unexpected output: %s", output)
			}
		})
	}
}
