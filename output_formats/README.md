# go2json Output Formats

This directory contains examples of go2json output formats.

## Files

- **format_llm.txt** - Go-syntax format optimized for LLMs (compact, same-type field grouping)
- **format_json.json** - Machine-readable JSON (for automation and integration)

## Generate These Files

```bash
go2json parse --path structparser.go --format llm > format_llm.txt
go2json parse --path structparser.go --format json > format_json.json
```

## Format Characteristics

### llm (Go-syntax for LLMs)
- Compact Go syntax with same-type field grouping
- Methods nested inside structs with `*` for pointer receivers
- No indentation, no alignment padding — minimal tokens
- Comments off by default, enable with `--comments`

### json (Machine-Readable)
- Fully structured
- Complete metadata
- Best for programmatic access
- Default output format

### grepindex (Grep-Friendly)
- Line-oriented index format
- Suitable for grep/awk pipelines
