# go2json

go2json is:
- a library for parsing and analyzing Go code
- a command line tool to parse and inspect Go code
- a stack trace analysis library for Go function parsing

## Installation

```bash
go build ./cmd/go2json
```

## Usage

### Parse a directory

```bash
./go2json parse --path . --recursive --format json
```

### Parse a single file

```bash
./go2json parse --path main.go --format llm
```

### Available Formats

- `llm` - Human-readable format optimized for LLM analysis (default)
- `json` - JSON format for programmatic processing
- `text_short` - Compact text format
- `text_long` - Detailed text format

### Help

```bash
./go2json parse --help
```

## Setup

### Environment Variables (Optional)

For AI features, create a `.env` file:

```
OPENAI_API_KEY=sk-...
```

## Dependencies

- Go 1.24.0+
- `golang.org/x/tools` - For Go AST analysis
- `github.com/urfave/cli/v2` - For CLI framework

## Examples

See the `examples/` directory for code generation and analysis examples:

- `example_01/` - Basic code generation
- `example_02/` - Request-based code analysis
- `example_03/` - File change application
- `example_04/` - Template-based code generation

Run examples:

```bash
cd examples/example_01
go run main.go
```

## Architecture

The tool consists of:

1. **AST Parser** (`structparser.go`) - Parses Go code into structured data
2. **Pretty Printer** (`parser_pretty_print.go`) - Formats parsed data in various output formats
3. **Code Surgeon** (`codesurgeon.go`) - Utilities for code generation and manipulation
4. **CLI** (`cmd/go2json/`) - Command-line interface

## Features

- **Recursive Parsing** - Analyzes entire directory trees
- **Selective Output** - Choose what to include (structs, functions, methods, comments, etc.)
- **Multiple Formats** - JSON, text, LLM-optimized formats
- **Ignore Patterns** - Skip test files, vendor directories, etc.
- **Code Generation** - Insert or modify code fragments in existing files
- **Template Support** - Generate code from Go templates
