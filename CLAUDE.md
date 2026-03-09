# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go2json is a Go library and CLI tool for analyzing and parsing Go code:
- **AST-based code analysis**: Extracts structs, methods, functions, interfaces, type definitions, and type information from Go source
- **Multiple output formats**: Supports `json`, `llm`, and `grepindex` formats
- **Type dependency traversal**: BFS-based `describe` command to explore type graphs

## Essential Commands

### Build and Install
```bash
go install ./cmd/go2json/     # Install go2json binary
go build -o ./bin/go2json ./cmd/go2json/  # Build to bin/
goimports -w .                # Fix imports before building
```

### Running the CLI
```bash
# Parse a directory (non-recursive, JSON output by default)
go2json parse --path ./mypackage

# Parse recursively with LLM format
go2json parse --path . --recursive --format llm

# Parse with ignore patterns
go2json parse --recursive --ignore-rule "*_test.go" --ignore-rule "vendor/*"

# Describe a type and its dependencies
go2json describe --type MyStruct --path . --depth 2

# Run from source
go run ./cmd/go2json parse --path .
```

### Testing
```bash
go test ./...              # Run all tests
go test -v ./...           # Verbose output
go test -run TestName ./...  # Run specific test
go test -v ./... -count=1  # Disable test caching
```

## Architecture Overview

### Core Packages

1. **`types.go`**: All data type definitions — `ParsedInfo`, `Module`, `Package`, `Struct`, `Method`, `Field`, `TypeDetails`, etc.

2. **`parser.go`**: Public Parse* API — `ParseDirectory`, `ParseDirectoryRecursive`, `ParseDirectoryWithFilter`, `ParseString`.

3. **`extract.go`**: AST extraction helpers — `extractParsedInfo`, `extractStructs`, `extractInterfaces`, `getFullType`, `parseFunctionDecl`, and all supporting unexported functions.

4. **`typeindex.go`**: BFS-based type dependency traversal — `BuildTypeIndex`, `DescribeType`.

5. **`format.go`**: Output formatting — `PrettyPrint` and all formatters for json, llm, and grepindex modes.

6. **`rewrite.go`**: Code manipulation utilities — `ApplyFileChanges`, `InsertCodeFragments`, `FindFunction`, `FormatCodeAndFixImports`, `EnsureGoFileExists`, `FormatWithGoImports`.

7. **`template.go`**: Go template rendering — `RenderTemplate`, `RenderTemplateNoError`, `MustRenderTemplate`.

8. **`cmd/go2json/main.go`**: CLI entry point using urfave/cli/v2 with `parse` and `describe` commands.

### Data Flow

1. **Parse Input**: CLI entry point accepts path and flags
2. **AST Analysis**: `ParseDirectoryRecursive()` or `ParseDirectoryWithFilter()` process Go files
3. **Extract Info**: `extract.go` extracts structs, methods, functions, interfaces, type defs, comments, and tags
4. **Format Output**: `PrettyPrint()` formats results based on requested format (json/llm/grepindex)

### Key Dependencies

- `golang.org/x/tools`: Go AST and type analysis
- `github.com/urfave/cli/v2`: CLI framework with command/flag support
- `github.com/Knetic/govaluate`: Expression evaluation for ignore rules

## Key Concepts

### ParsedInfo Structure

The main output type containing:
- **Structs**: Struct definitions with fields and receiver methods
- **Functions**: Top-level functions
- **Interfaces**: Interface definitions with methods
- **TypeDefs**: Named type declarations (not struct/interface)
- **Variables/Constants**: Package-level declarations
- **Comments**: Associated documentation
- **Tags**: Struct field tags (json, db, etc.)

### Output Formats

- `json` (default): Machine-readable JSON
- `llm`: Compact Go-syntax format with same-type field grouping, methods nested in structs
- `grepindex`: Line-oriented index format for grep/awk pipelines

## Common Patterns

### Parsing from Code
```go
import "github.com/wricardo/go2json"

parsed, err := go2json.ParseDirectoryRecursive("./mypackage")
output := go2json.PrettyPrint(parsed, "llm", nil, true, true, true, true, true, true, true, false, false)
```

### Describe a Type
```go
parsed, _ := go2json.ParseDirectoryRecursive("./mypackage")
result, _ := go2json.DescribeType("MyStruct", parsed, 2)
fmt.Println(go2json.PrettyPrint(result, "llm", nil, true, true, true, true, true, true, true, false, false))
```

## Important Notes

- Go 1.24.0+ required (see go.mod)
- No external services required for parsing
- Pure Go library - no dependencies on databases, AI services, or networking tools

## Command Reference

For detailed CLI flag documentation, see [docs/COMMANDS.md](docs/COMMANDS.md)
