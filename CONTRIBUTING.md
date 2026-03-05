# Contributing to go2json

Thank you for your interest in contributing! Here's how to get started.

## Setting Up Your Development Environment

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd go2json
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build and test:
   ```bash
   go build -o ./bin/go2json ./cmd/go2json/
   go test ./...
   ```

## Development Workflow

### Before Making Changes

- Create a new branch for your work: `git checkout -b feature/your-feature-name`
- Keep changes focused on a single feature or fix

### Code Standards

- Run `goimports -w .` to fix imports before committing
- Ensure all tests pass: `go test ./...`
- Write tests for new functionality
- Follow Go style guidelines and conventions

### Commit Messages

Write clear, descriptive commit messages:
- Start with a verb: "Add", "Fix", "Update", "Refactor"
- Keep the summary under 70 characters
- Include context about why the change was made

Example:
```
Add support for parsing interface types

This enables go2json to analyze interface definitions and their methods,
which is useful for code documentation and API analysis.
```

## Submitting Changes

1. Push your branch to your fork
2. Create a pull request with a clear description of your changes
3. Reference any related issues

### PR Guidelines

- Link to related issues
- Include test coverage for new features
- Update documentation if needed
- Keep PRs focused and reasonably sized

## Reporting Issues

When reporting bugs, include:
- Go version (`go version`)
- Steps to reproduce
- Expected vs. actual behavior
- Code sample or file that demonstrates the issue

## Questions?

Feel free to open an issue to ask questions or discuss ideas before investing significant effort.
