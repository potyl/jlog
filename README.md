# jlog

A simple CLI tool for parsing JSON logs and extracting messages.

## Overview

`jlog` reads logs line by line (from stdin or a file) and extracts the `.message` field from JSON-formatted log entries. Non-JSON lines are passed through unchanged.

## Installation

```bash
go install github.com/potyl/jlog@latest
```

Or build from source:

```bash
go build -o jlog
```

## Usage

Read from stdin:

```bash
echo '{"level":"info","message":"Hello, World!","timestamp":"2024-01-01"}' | jlog
# Output: Hello, World!
```

Read from a file:

```bash
jlog application.log
```

## Behavior

- **JSON lines**: Extracts and prints the `.message` field
- **Non-JSON lines**: Prints the line as-is
- **JSON without `.message`**: Prints the original line

## Examples

```bash
# Mix of JSON and non-JSON logs
cat <<EOF | jlog
{"message":"User logged in","user":"john"}
Plain text log entry
{"level":"error","message":"Connection failed"}
EOF

# Output:
# User logged in
# Plain text log entry
# Connection failed
```

## Requirements

- Go 1.21 or later

## License

MIT — see [LICENSE](LICENSE) for details.
