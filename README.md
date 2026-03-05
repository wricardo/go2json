# go2json

A Go code parser that outputs structured JSON (or compact Go-syntax) for every struct, function, method, interface, variable, and constant in a package.

The JSON output is machine-readable, so you can pipe it into scripts or AI tools. For example, an LLM can iterate over every function in a package to generate tests, find which handlers match a route pattern, or build a dependency graph — without parsing Go itself.

## Install

```bash
go install github.com/wricardo/go2json/cmd/go2json@latest
```

Or build from source:

```bash
go build -o go2json ./cmd/go2json
```

## Commands

### parse

Parse Go source files and extract structural information.

```bash
# Parse current directory (JSON output, default)
go2json parse --path .

# Parse recursively
go2json parse --path . --recursive

# Compact Go-syntax output (optimized for LLMs)
go2json parse --path . --format llm

# With comments enabled
go2json parse --path . --format llm --comments

# Ignore test files
go2json parse --path . --recursive --ignore-rule "*_test.go"
```

### describe

Show a type and all the types it depends on, expanded to a configurable depth.

```bash
# Describe a type with default depth (1 level of dependencies)
go2json describe --type UserService

# Go deeper into the dependency tree
go2json describe --type Order --depth 3 --path ./internal/models

# JSON output for programmatic use
go2json describe --type Config --format json --omit-nulls
```

### install-skill

Install the go2json [Claude Code](https://claude.ai/code) skill to `~/.claude/skills/go2json/`. This makes the skill available in all Claude Code sessions, giving the AI agent full knowledge of go2json's commands, flags, and usage patterns.

```bash
go2json install-skill
```

## Output Formats

- `json` (default) — structured JSON with full type details, bodies, params, returns, and tags
- `llm` — compact Go-syntax format with same-type field grouping, no indentation, minimal tokens
- `grepindex` — one-line-per-entity flat index, pipe through `grep` to search

### LLM Format Example

```
// directory: ./mypackage
package mypackage
type Config struct{
Host,Port,Path string
Timeout int
Debug,Verbose bool
*Validate() error
*Apply(ctx context.Context) error
}
func NewConfig(host string,port int) *Config
var DefaultConfig Config
const MaxRetries = 3
```

Fields with the same type are grouped (`Host,Port,Path string`). Methods are nested inside their struct with `*` for pointer receivers. No alignment padding, no indentation — minimal tokens.

### grepindex Format Example

```
directory> /project
package> models directory: /project/internal/models
struct> User has_methods package: models
field> ID (string) struct: User has_methods package: models
method> FullName() string receiver: (*User) package: models
function> NewUser(name, email string) *User package: models
```

Pipe through grep:

```bash
# All structs in the project
go2json parse -r --format grepindex | grep "^struct>"

# All methods on a specific type
go2json parse -r --format grepindex | grep "receiver: (\*User)"
```

## Flags

```
--path, -f                     path to parse (default: ".")
--recursive, -r                recurse into subdirectories
--format                       json, llm, or grepindex (default: "json")
--comments                     include doc comments (default: false)
--plain-structs                include structs without methods (default: true)
--structs-with-method          include structs with methods (default: true)
--fields-plain-structs         include fields of plain structs (default: true)
--fields-structs-with-method   include fields of structs with methods (default: true)
--methods                      include methods (default: true)
--functions                    include functions (default: true)
--tags                         include struct tags (default: true)
--omit-nulls                   omit null/empty values from JSON (default: false)
--ignore-rule                  glob pattern to ignore files (repeatable)
```

## Library Usage

```go
import g2j "github.com/wricardo/go2json"

parsed, err := g2j.ParseDirectoryRecursive("./mypackage")
if err != nil {
    log.Fatal(err)
}
fmt.Println(g2j.PrettyPrint(parsed, "llm", nil, true, true, true, true, true, true, true, false, false))
```

## Requirements

- Go 1.24.0+
