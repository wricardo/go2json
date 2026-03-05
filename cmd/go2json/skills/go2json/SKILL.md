---
name: go2json
description: >
  Go code analysis CLI that extracts structs, methods, functions, interfaces, type definitions,
  and their relationships from Go source code. Outputs structured data in LLM-friendly, JSON, or
  grep-optimized formats. Use when you need to: understand a Go codebase's structure, list all types
  in a package, find struct fields and method signatures, trace type dependencies, generate context
  for code generation, or build a map of a Go project before making changes. Triggers on: "analyze
  this Go code", "list all structs", "what types are in this package", "describe this type",
  "show me the struct fields", "map the codebase", or any Go code introspection task.
---

# go2json — Go Code Analysis for AI Agents

go2json parses Go source code via AST analysis and outputs structured information about structs,
functions, methods, interfaces, type definitions, constants, and variables. It is the fastest way
to build a mental model of a Go codebase without reading every file.

## Install

```bash
go install github.com/wricardo/go2json/cmd/go2json@latest
```

## Commands

### `parse` — Extract everything from a package or directory tree

```bash
go2json parse [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--path` | `-f` | `.` | Directory or file to parse |
| `--recursive` | `-r` | `false` | Recurse into subdirectories |
| `--format` | | `json` | Output format: `json`, `llm`, `grepindex` |
| `--omit-nulls` | | `false` | Drop null/empty fields from JSON output |
| `--ignore-rule` | | | Glob pattern to skip files (repeatable) |
| `--plain-structs` | | `true` | Include structs without methods |
| `--fields-plain-structs` | | `true` | Include fields of plain structs |
| `--structs-with-method` | | `true` | Include structs that have methods |
| `--fields-structs-with-method` | | `true` | Include fields of method-bearing structs |
| `--methods` | | `true` | Include method signatures and bodies |
| `--functions` | | `true` | Include top-level functions |
| `--comments` | | `false` | Include doc comments |
| `--tags` | | `true` | Include struct field tags |

### `describe` — Show a type and its dependency tree

```bash
go2json describe --type <TypeName> [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--type` | `-t` | (required) | Type name to describe |
| `--path` | `-f` | `.` | Directory to parse |
| `--depth` | `-n` | `1` | How many levels of referenced types to expand |
| `--format` | | `llm` | Output format: `llm`, `json`, `grepindex` |
| `--omit-nulls` | | `false` | Drop null/empty fields from JSON output |

## Output Formats

### `llm` — Go-syntax summary (best for AI agents)

Compact, readable Go-like syntax. Ideal for stuffing into LLM context windows — minimal tokens,
maximum information density. Shows type definitions, field types, and method signatures in
familiar Go notation.

```bash
go2json parse --path ./internal/models --format llm
```

Output:
```go
// directory: /project/internal/models
package models
type User struct{
ID        string
Name      string
Email     string
CreatedAt time.Time
Orders    []Order
}
func (u *User) FullName() string
func (u *User) IsActive() bool
type Order struct{
ID     string
UserID string
Total  float64
Items  []OrderItem
}
func NewUser(name, email string) *User
func FindUserByEmail(db *sql.DB, email string) (*User, error)
```

### `json` — Full structured data (best for programmatic use)

Complete AST information as JSON. Every struct, field, method, and function includes full type
details, parameter lists, return types, doc comments, tags, and body source code.

```bash
go2json parse --path . --format json --omit-nulls
```

The JSON output includes `TypeDetails` on every field and parameter with:
- `TypeName`, `Type`, `Package`, `PackageName` — full type path info
- `IsPointer`, `IsSlice`, `IsMap`, `IsBuiltin`, `IsExternal` — type classification flags
- `TypeReferences` — referenced types for further exploration

### `grepindex` — One-line-per-entity index (best for searching/filtering)

Flat, grep-friendly format. Each line is a single entity (struct, field, method, function) with
full context breadcrumbs. Pipe through `grep` to find anything instantly.

```bash
go2json parse --path . -r --format grepindex
```

Output:
```
directory> /project
package> models directory: /project/internal/models
struct> User has_methods package: models directory: /project/internal/models
field> ID (string) struct: User has_methods package: models
field> Email (string) struct: User has_methods package: models
method> FullName() string receiver: (*User) package: models
function> NewUser(name, email string) *User package: models
```

Use with grep:
```bash
# Find all structs in the project
go2json parse -r --format grepindex | grep "^struct>"

# Find all methods on a specific type
go2json parse -r --format grepindex | grep "receiver: (\*User)"

# Find all functions that return an error
go2json parse -r --format grepindex | grep "^function>" | grep "error"
```

## Recipes for AI Agents

### 1. Map an entire codebase before making changes

```bash
go2json parse --path . --recursive --format llm --ignore-rule "*_test.go" --ignore-rule "vendor/*"
```

This gives you every type, function, and method in the project in a compact format. Use this
as context before generating code, so you know what already exists and can reuse existing types.

### 2. Understand a specific type and what it depends on

```bash
go2json describe --type UserService --path . --depth 2
```

Shows `UserService` plus all types referenced by its fields and method signatures, expanded
two levels deep. Use this when you need to understand a type's full interface before modifying it.

Depth controls how far to follow type references:
- `--depth 1` (default): The target type + directly referenced types
- `--depth 2`: Also expands types referenced by those types
- `--depth 3+`: Keeps going (use sparingly on large codebases)

### 3. List only function signatures (no structs, no fields)

```bash
go2json parse --path ./internal/service -r --format llm \
  --plain-structs=false --structs-with-method=false
```

Useful when you only care about the function API surface, not data types.

### 4. List only struct definitions (no functions, no methods)

```bash
go2json parse --path . -r --format llm --functions=false --methods=false
```

Quick inventory of all data types in the codebase.

### 5. Get JSON schema of a type for validation or code generation

```bash
go2json describe --type CreateUserInput --format json --omit-nulls
```

Returns full field details including types, tags (json/db/validate), and nested type info.
Use this to generate validation logic, API documentation, or test fixtures.

### 6. Find all types in a package quickly

```bash
go2json parse --path ./internal/models --format grepindex | grep "^struct>\|^type>"
```

### 7. Get struct tags for database/JSON mapping

```bash
go2json parse --path ./internal/models --format json --omit-nulls --tags \
  --functions=false --methods=false
```

Extracts all struct field tags (`json:"name"`, `db:"column"`, `validate:"required"`) —
useful for checking serialization behavior or generating migration scripts.

### 8. Explore a large project incrementally

```bash
# Step 1: Get package-level overview
go2json parse --path . -r --format grepindex | grep "^package>"

# Step 2: Drill into an interesting package
go2json parse --path ./internal/auth --format llm

# Step 3: Deep-dive into a specific type
go2json describe --type AuthMiddleware --path ./internal/auth --depth 2
```

### 9. Pipe to another LLM for analysis (with venu)

```bash
go2json parse --path . -r --format llm --ignore-rule "*_test.go" \
  | venu "List all exported types that look like API request/response DTOs"

go2json describe --type Order --depth 2 --format llm \
  | venu "What fields are missing for a complete audit trail?"
```

### 10. Compare what a package exports vs what tests cover

```bash
# All exported functions
go2json parse --path ./internal/service --format grepindex \
  | grep "^function>" | grep -v "is_exported: false"

# All test functions
go2json parse --path ./internal/service --format grepindex --ignore-rule "!*_test.go" \
  | grep "^function>" | grep "Test"
```

## When to Use Each Format

| Scenario | Format | Why |
|----------|--------|-----|
| Feed context to an AI agent | `llm` | Minimal tokens, maximum info density |
| Code generation / automation | `json` | Full type details, machine-readable |
| Quick searching / filtering | `grepindex` | One line per entity, grep-friendly |
| Understanding a type's API | `llm` + `describe` | Shows the type + its dependencies |
| Debugging type relationships | `json` + `describe --depth 2` | Full nested type info |

## What go2json Extracts

For each Go package, go2json extracts:

- **Structs**: Name, fields (name, type, tags, visibility), doc comments, whether exported
- **Methods**: Receiver type, name, params, returns, signature, body, doc comments
- **Functions**: Name, params, returns, signature, body, doc comments, test/benchmark flags
- **Interfaces**: Name, method signatures, doc comments
- **Type definitions**: `type Foo Bar` aliases and named types
- **Constants and variables**: Name, value, doc comments
- **Imports**: Name (alias) and path
- **Modules**: Module name, relative directory, contained packages

Field-level type details include: whether the type is a pointer, slice, map, builtin, or
external (from another module), plus the full package path for non-builtin types.
