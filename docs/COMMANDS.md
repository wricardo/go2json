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
- `--format` - Output format: `json`, `llm`, `grepindex` (default: "json")
- `--plain-structs` - Print plain structs (default: true)
- `--fields-plain-structs` - Print fields of plain structs (default: true)
- `--structs-with-method` - Print structs with methods (default: true)
- `--fields-structs-with-method` - Print fields of structs with methods (default: true)
- `--methods` - Print methods (default: true)
- `--functions` - Print functions (default: true)
- `--comments` - Print comments (default: false)
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

- `json` - Full structured JSON (default). Includes type details, bodies, params, returns, tags.
- `llm` - Compact Go-syntax format optimized for LLM context windows.
- `grepindex` - One-line-per-entity flat index, pipe through grep to search.

### describe

Describe a type and its dependency tree.

```bash
go2json describe --type <TypeName> [options]
```

**Alias:** `d`

**Options:**
- `--type`, `-t` - Type name to describe (required)
- `--path`, `-f` - Path to directory to parse (default: ".")
- `--depth`, `-n` - Max depth of type dependency traversal (default: 1)
- `--format` - Output format: `llm`, `json`, `grepindex` (default: "llm")
- `--omit-nulls` - Omit null and empty values from JSON output (default: false)

**Examples:**

Describe a type with default depth:
```bash
go2json describe --type UserService
```

Describe with deeper dependency traversal:
```bash
go2json describe --type Order --depth 3 --path ./internal/models
```

Describe with JSON output:
```bash
go2json describe --type Config --format json --omit-nulls
```

### install-skill

Install the go2json Claude Code skill to `~/.claude/skills/go2json/`.

```bash
go2json install-skill
```

This copies the embedded SKILL.md file to your user-level Claude Code skills directory, making the go2json skill available in all Claude Code sessions.

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

Describe a type and its full dependency graph:
```bash
go2json describe --type Handler --depth 2 --path ./internal/api
```
