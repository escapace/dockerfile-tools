# dockerfile-tools

A command-line application designed to assist with analyzing and extracting
information from Dockerfiles. It provides three main functionalities: generating a
JSON Abstract Syntax Tree (AST) from a Dockerfile, listing the build stages defined
within a Dockerfile, and listing cache mounts in RUN instructions.

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

Lists the build stages of a specified Dockerfile.

#### Options

- `--dockerfile string`: Path to the Dockerfile.
- `--help`: Display help information for the `list-stages` command.

#### Example

```bash
dockerfile-tools list-stages --dockerfile path/to/Dockerfile
```

### list-cache-mounts

Extracts --mount=type=cache flags from RUN instructions in a Dockerfile and outputs a JSON object.
The target paths are used as the values, while the keys are derived by replacing / with spaces and
converting the result to kebab-case. The output is compatible with the `cache-map` option of
[buildkit-cache-dance](https://github.com/reproducible-containers/buildkit-cache-dance), making it
easier to define cache mappings for reproducible builds.

#### Options

- `--dockerfile string`: Path to the Dockerfile.
- `--arg`: comma-delimited ARG key-value pairs
- `--help`: Display help information for the `list-cache-mounts` command.

#### Example

```bash
dockerfile-tools list-cache-mounts --dockerfile path/to/Dockerfile \
--arg BUILDPLATFORM=linux/amd64 --arg BUILDOS=linux,BUILDARCH=amd64
```

## Help

For general help, run the application without any arguments:

```bash
dockerfile-tools
```

This will display a list of available commands and their descriptions.
