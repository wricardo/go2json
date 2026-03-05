# go2json Output Formats Comparison

This directory contains examples of all 4 output formats from go2json when parsing `structparser.go`.

## Files

- **format_llm.txt** - Default LLM-optimized format (human-readable for AI processing)
- **format_text_short.txt** - Compact text format (quick overview)
- **format_text_long.txt** - Detailed text format (comprehensive, includes tags and comments)
- **format_json.json** - Machine-readable JSON (for automation and integration)

## Generate These Files

```bash
./go2json parse --path structparser.go --format llm > format_llm.txt
./go2json parse --path structparser.go --format text_short > format_text_short.txt
./go2json parse --path structparser.go --format text_long > format_text_long.txt
./go2json parse --path structparser.go --format json > format_json.json
```

## Quick Size Comparison

```bash
wc -l output_formats/format_*.txt output_formats/format_json.json
```

## Format Characteristics

### llm (Human-Readable for LLM)
- Most concise
- Optimized for language model consumption
- Shows type information inline
- Good for prompts and documentation

### text_short (Compact Text)
- Medium verbosity
- One field per line
- Easy to scan
- No extra metadata

### text_long (Detailed Text)
- Most verbose
- Includes struct tags (json, db, etc.)
- Shows field comments
- Comprehensive documentation

### json (Machine-Readable)
- Fully structured
- Complete metadata
- Best for programmatic access
- Suitable for APIs and CI/CD

## Usage Tips

1. Use **llm** for sending to AI models or creating documentation
2. Use **text_short** for quick terminal viewing
3. Use **text_long** for detailed code review
4. Use **json** for scripting, filtering, and automation

Example filtering JSON:
```bash
cat format_json.json | jq '.modules[0].packages[0].structs[] | {name, fields: [.fields[] | .name]}'
```
