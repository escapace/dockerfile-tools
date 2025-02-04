package main

import (
	"fmt"
	"os"
)
import flag "github.com/spf13/pflag"

func printHelp() {
	fmt.Println("dockerfile-tools <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  ast                Generate a JSON AST from a Dockerfile")
	fmt.Println("  list-stages        List the build stages of a Dockerfile")
	fmt.Println("  list-cache-mounts  List the cache mounts of a Dockerfile")
	fmt.Println("Use \"dockerfile-tools <command> --help\" for more information about a command.")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "ast":
		astCmd := flag.NewFlagSet("ast", flag.ExitOnError)
		dockerfile := astCmd.String("dockerfile", "", "path to Dockerfile")
		astHelp := astCmd.Bool("help", false, "display help")
		astCmd.Parse(os.Args[2:])

		if *astHelp {
			fmt.Println("dockerfile-tools ast [options]")
			fmt.Println("")
			fmt.Println("  --dockerfile string")
			fmt.Println("        path to Dockerfile")
			fmt.Println("  --help")
			fmt.Println("        display help")
			os.Exit(0)
		}

		if *dockerfile == "" {
			fmt.Println("Please provide a path to the Dockerfile using --dockerfile")
			os.Exit(1)
		}

		// Call the function from ast-json.go
		AST(*dockerfile)

	case "list-stages":
		listStagesCmd := flag.NewFlagSet("list-stages", flag.ExitOnError)
		dockerfile := listStagesCmd.String("dockerfile", "", "path to Dockerfile")
		listStagesHelp := listStagesCmd.Bool("help", false, "display help")
		listStagesCmd.Parse(os.Args[2:])

		if *listStagesHelp {
			fmt.Println("dockerfile-tools list-stages [options]")
			fmt.Println("")
			fmt.Println("  --dockerfile string")
			fmt.Println("        path to Dockerfile")
			fmt.Println("  --help")
			fmt.Println("        display help")
			os.Exit(0)
		}

		if *dockerfile == "" {
			fmt.Println("Please provide a path to the Dockerfile using --dockerfile")
			os.Exit(1)
		}

		// Call the function from list-stages.go
		ListStages(*dockerfile)

	case "list-cache-mounts":
		listCacheMountsCmd := flag.NewFlagSet("list-cache-mounts", flag.ExitOnError)
		dockerfile := listCacheMountsCmd.String("dockerfile", "", "path to Dockerfile")
    args := listCacheMountsCmd.StringSlice("arg", nil, "comma-delimited ARG key-value pairs (can be provided multiple times)")
		listCacheMountsHelp := listCacheMountsCmd.Bool("help", false, "display help")
		listCacheMountsCmd.Parse(os.Args[2:])

		if *listCacheMountsHelp {
			fmt.Println("dockerfile-tools list-cache-mounts [options]")
			fmt.Println("")
			fmt.Println("  --dockerfile string")
			fmt.Println("        path to Dockerfile")
			fmt.Println("  --arg string")
			fmt.Println("        comma-delimited ARG key-value pairs (can be provided multiple times)")
			fmt.Println("  --help")
			fmt.Println("        display help")
			os.Exit(0)
		}


		if *dockerfile == "" {
			fmt.Println("Please provide a path to the Dockerfile using --dockerfile")
			os.Exit(1)
		}

		// Call the function
		ListCacheMounts(*dockerfile, *args)

	default:
		fmt.Println("Error: expected 'ast' or 'list-stages' subcommands")
		printHelp()
		os.Exit(1)
	}
}
