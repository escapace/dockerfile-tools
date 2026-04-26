# dockerfile-tools

Command-line utility for parsing Dockerfiles into JSON syntax trees, listing `ARG` declarations, listing named build stages, and extracting cache mounts with `ARG` value expansion.

GitHub Action: [escapace/action-dockerfile-tools](https://github.com/escapace/action-dockerfile-tools)

## Usage

Available commands:

### ast

Generates a JSON AST from a specified Dockerfile.

#### Options

- `--dockerfile string`: Path to the Dockerfile.
- `--help`: Display help information for the `ast` command.

#### Example

```bash
dockerfile-tools ast --dockerfile path/to/Dockerfile
```

### list-args

Lists `ARG` declarations in the order they appear in the Dockerfile.
Each declaration includes the argument name, the declared default value when present, and the
source line number. Declarations without a default value are reported with `"defaultValue": null`.
If an explicit empty default is declared, for example `ARG EMPTY=`, the output keeps that as an
empty string.

Example output:

```json
[
  {
    "name": "GO_VERSION",
    "defaultValue": "1.26",
    "line": 1
  },
  {
    "name": "TARGETPLATFORM",
    "defaultValue": null,
    "line": 2
  }
]
```

#### Options

- `--dockerfile string`: Path to the Dockerfile.
- `--help`: Display help information for the `list-args` command.

#### Example

```bash
dockerfile-tools list-args --dockerfile path/to/Dockerfile
```

### list-stages

Lists the aliased build stages of a specified Dockerfile.
Only `FROM ... AS <name>` stages are included in the output.

#### Options

- `--dockerfile string`: Path to the Dockerfile.
- `--help`: Display help information for the `list-stages` command.

#### Example

```bash
dockerfile-tools list-stages --dockerfile path/to/Dockerfile
```

### list-cache-mounts

Extracts `type=cache` mounts from `RUN` instructions in a Dockerfile and outputs a JSON object.
Each key is derived from the mount `id` when present, or from `target` as a fallback, by converting
that value to kebab-case and prefixing it with `.cache-`. Each value is an object containing the
parsed mount options, such as `type`, `target`, `id`, and `sharing`.

The parser recognizes cache mounts regardless of option order. It also expands both `$ARG` and
`${ARG}` placeholders using values passed with `--arg`, plus default values for `BUILDOS`,
`BUILDARCH`, and `BUILDPLATFORM`.

Example output:

```json
{
  ".cache-go-pkg-mod": {
    "type": "cache",
    "target": "/go/pkg/mod",
    "sharing": "locked"
  }
}
```

#### Options

- `--dockerfile string`: Path to the Dockerfile.
- `--arg`: comma-delimited ARG key-value pairs. Can be provided multiple times.
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
