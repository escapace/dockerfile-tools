# dockerfile-tools

A command line application designed to assist with analyzing and extracting
information from Dockerfiles. It provides two main functionalities: generating a
JSON Abstract Syntax Tree (AST) from a Dockerfile and listing the build stages
defined within a Dockerfile.

## Usage

The application provides the following commands:

### ast

Generates a JSON AST from a specified Dockerfile.

#### Options

- `--dockerfile string`: Path to the Dockerfile.
- `--help`: Display help information for the `ast` command.

#### Example

```bash
dockerfile-tools ast --dockerfile path/to/Dockerfile
```

### list-stages

Lists the build stages from a specified Dockerfile.

#### Options

- `--dockerfile string`: Path to the Dockerfile.
- `--help`: Display help information for the `list-stages` command.

#### Example

```bash
dockerfile-tools list-stages --dockerfile path/to/Dockerfile
```

## Help

For general help, run the application without any arguments:

```bash
dockerfile-tools
```

This will display a list of available commands and their descriptions.
