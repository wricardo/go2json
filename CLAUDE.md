# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go2json is a Go library and CLI tool for analyzing and parsing Go code:
- **AST-based code analysis**: Extracts structs, methods, functions, and type information from Go source
- **Multiple output formats**: Supports `llm`, `text_short`, `text_long`, and `json` formats
- **AI integration**: Uses instructor-go with OpenAI/Anthropic APIs for structured code analysis
- **Stack trace parsing**: Analyzes Go function signatures and receiver types

## Essential Commands

### Build and Install
```bash
go install ./cmd/go2json/     # Install go2json binary
go build -o ./bin/go2json ./cmd/go2json/  # Build to bin/
goimports -w .                # Fix imports before building
```

### Running the CLI
```bash
# Parse a directory (non-recursive)
go2json parse --path ./mypackage

# Parse recursively with JSON output
go2json parse --path . --recursive --format json

# Parse with ignore patterns
go2json parse --recursive --ignore-rule "*_test.go" --ignore-rule "vendor/*"

# Run from source
go run ./cmd/go2json parse --path .
```

### Testing
```bash
go test ./...              # Run all tests
go test -v ./...           # Verbose output
go test ./structparser/...  # Test specific package
go test -run TestName ./...  # Run specific test
go test -v ./... -count=1  # Disable test caching
```

## Architecture Overview

### Core Packages

1. **`structparser.go`**: Parses Go source files to extract struct definitions, fields, methods, and functions. Core AST analysis logic.

2. **`codesurgeon.go`**: Provides code manipulation utilities for analyzing function signatures, receiver types, and call chains. Used by stack trace analysis.

3. **`ai/ai.go`**: Unified interface for OpenAI and Anthropic APIs using instructor-go for structured, type-safe outputs.

4. **`log2/log2.go`**: Logging configuration using zerolog.

5. **`cmd/go2json/main.go`**: CLI entry point using urfave/cli/v2 framework with the `parse` command.

### Data Flow

1. **Parse Input**: CLI entry point in `cmd/go2json/main.go` accepts path and flags
2. **AST Analysis**: `ParseDirectoryRecursive()` or `ParseDirectoryWithFilter()` process Go files
3. **Extract Info**: `structparser.go` extracts structs, methods, functions, comments, and tags
4. **Format Output**: `PrettyPrint()` formats results based on requested format (llm/json/text)

### Key Dependencies

- `golang.org/x/tools`: Go AST and type analysis
- `github.com/instructor-ai/instructor-go`: Type-safe AI responses
- `github.com/urfave/cli/v2`: CLI framework with command/flag support
- `github.com/rs/zerolog`: Structured logging

## Development Workflow

1. **Environment Setup** (optional, only for AI features):
   ```bash
   cp .env.example .env
   # Add: OPENAI_API_KEY=sk-...
   ```

2. **Before Committing**:
   ```bash
   goimports -w .
   go test ./...
   ```

3. **Debugging a Specific Test**:
   ```bash
   go test -v -run TestStructParserBasic ./structparser/ -count=1
   ```

## Key Concepts

### ParsedInfo Structure

The main output type containing:
- **Structs**: Struct definitions with fields and receiver methods
- **Functions**: Top-level functions
- **Methods**: Methods associated with receiver types
- **Comments**: Associated documentation
- **Tags**: Struct field tags (json, db, etc.)

### Output Formats

- `llm`: Human-readable format optimized for LLM analysis
- `text_short`: Compact text output
- `text_long`: Detailed text with full signatures
- `json`: Machine-readable JSON (useful for automation)

## Important Notes

- Go 1.24.0+ required (see go.mod)
- No external services required for basic parsing (AI features optional)
- Stack trace analysis functions (from deleted neo4j2 package) are in `codesurgeon.go`

## Common Patterns

### Parsing from Code
```go
import g2j "github.com/wricardo/go2json"

// Parse a directory
parsed, err := g2j.ParseDirectoryRecursive("./mypackage")

// Pretty print with formatting
output := g2j.PrettyPrint(parsed, "json", nil, true, true, true, true, true, true, true, true, false)
```

### Stack Trace Parsing (from codesurgeon.go)
```go
// Parse function signatures from stack traces
function, receiver := g2j.ParseReceiver("(*MyStruct).DoSomething")
// Returns: function="DoSomething", receiver="(*MyStruct)"
```

## Command Reference

For detailed CLI flag documentation, see [docs/COMMANDS.md](docs/COMMANDS.md)