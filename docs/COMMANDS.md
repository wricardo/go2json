# go2json CLI Commands Reference

This document provides a comprehensive reference for all commands available in the go2json CLI tool.

## Commands

### parse

Parse Go code to extract structural information.

```bash
go2json parse [options]
```

**Alias:** `p`

**Options:**
- `--path`, `-f` - Path to file or directory to parse (default: ".")
- `--recursive`, `-r` - Recursively parse directories
- `--format` - Output format: `llm`, `text_short`, `text_long`, `json` (default: "llm")
- `--plain-structs` - Print plain structs (default: true)
- `--fields-plain-structs` - Print fields of plain structs (default: true)
- `--structs-with-method` - Print structs with methods (default: true)
- `--fields-structs-with-method` - Print fields of structs with methods (default: true)
- `--methods` - Print methods (default: true)
- `--functions` - Print functions (default: true)
- `--comments` - Print comments (default: true)
- `--tags` - Print struct tags (default: true)
- `--ignore-rule` - Ignore files or directories matching the pattern (repeatable)
- `--omit-nulls` - Omit null and empty values from JSON output (default: false)

**Examples:**

Parse current directory:
```bash
go2json parse
```

Parse with recursive directory traversal:
```bash
go2json parse --path . --recursive
```

Parse specific file with JSON output:
```bash
go2json parse --path main.go --format json
```

Parse with ignore rules:
```bash
go2json parse --recursive --ignore-rule "*_test.go" --ignore-rule "vendor/*"
```

Parse and omit null values:
```bash
go2json parse --path . --format json --omit-nulls
```

**Output Formats:**

- `llm` - Human-readable format optimized for LLM analysis
- `text_short` - Compact text format
- `text_long` - Detailed text format
- `json` - JSON format for programmatic processing

## Examples

Parse the current directory recursively and output as JSON:
```bash
go2json parse --recursive --format json
```

Parse a specific package:
```bash
go2json parse --path ./internal/parser --recursive --format llm
```

Parse and filter out test files:
```bash
go2json parse --recursive --ignore-rule "*_test.go"
```
