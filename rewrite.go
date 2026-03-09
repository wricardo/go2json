package go2json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"

	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/imports"
)

// FileChange describes a set of code fragments to write into a target file.
type FileChange struct {
	PackageName string
	File        string
	Fragments   []CodeFragment
}

// CodeFragment is a piece of Go source to insert or replace in a file.
type CodeFragment struct {
	Content   string
	Overwrite bool
}

// GoList represents the output of `go list -json` for a package.
type GoList struct {
	Dir         string       `json:"Dir"`
	ImportPath  string       `json:"ImportPath"`
	Name        string       `json:"Name"`
	Target      string       `json:"Target"`
	Root        string       `json:"Root"`
	Module      GoListModule `json:"Module"`
	Match       []string     `json:"Match"`
	Stale       bool         `json:"Stale"`
	StaleReason string       `json:"StaleReason"`
	GoFiles     []string     `json:"GoFiles"`
	Imports     []string     `json:"Imports"`
	Deps        []string     `json:"Deps"`
}

// GoListModule represents the module section of `go list -json` output.
type GoListModule struct {
	Path      string `json:"Path"`
	Main      bool   `json:"Main"`
	Dir       string `json:"Dir"`
	GoMod     string `json:"GoMod"`
	GoVersion string `json:"GoVersion"`
}

// ApplyFileChanges applies a set of code changes to multiple files.
// It creates directories and files as needed, then inserts or updates code fragments
// in each target file. Changes are grouped by file path before being applied.
func ApplyFileChanges(changes []FileChange) error {
	implementationsMap := make(map[string][]CodeFragment)
	for _, change := range changes {
		implementationsMap[change.File] = change.Fragments
		if err := os.MkdirAll(filepath.Dir(change.File), 0755); err != nil {
			return fmt.Errorf("Failed to create directory: %v", err)
		}
		if _, err := os.Stat(change.File); os.IsNotExist(err) {
			if f, err := os.Create(change.File); err != nil {
				return fmt.Errorf("Failed to create file: %v", err)
			} else {
				f.Write([]byte("package " + change.PackageName + "\n"))
				defer f.Close()
			}
		}
	}

	return InsertCodeFragments(implementationsMap)
}

// InsertCodeFragments inserts or updates code fragments in Go files.
// It parses each file, applies code changes using AST manipulation, and writes the
// modified files back. If Overwrite is true, existing declarations are replaced.
func InsertCodeFragments(implementationsMap map[string][]CodeFragment) error {
	for file, fragments := range implementationsMap {
		fset := token.NewFileSet()
		if _, err := os.Stat(file); os.IsNotExist(err) {
			if f, err := os.Create(file); err != nil {
				return fmt.Errorf("Failed to create file: %v", err)
			} else {
				f.Write([]byte("package main\n"))
				defer f.Close()
			}
		}
		node, err := parser.ParseFile(fset, file, nil, parser.AllErrors|parser.ParseComments)
		if err != nil {
			fmt.Printf("Failed to parse file: %v\n", err)
			continue
		}

		for _, fragment := range fragments {
			decls, err := parseDeclarationsFromCodeFragment(fragment)
			if err != nil {
				fmt.Printf("Failed to parse change: %v\n", err)
				continue
			}
			for _, decl := range decls {
				upsertDeclaration(node, decl, fragment.Overwrite)
			}
		}

		err = writeChangesToFile(file, fset, node)
		if err != nil {
			fmt.Printf("Failed to write modified file: %v\n", err)
			return err
		}
	}
	return nil
}

// FormatCodeAndFixImports applies gofmt and goimports to the given file.
func FormatCodeAndFixImports(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	processedContent, err := imports.Process(filePath, content, nil)
	if err != nil {
		return err
	}

	formattedContent, err := format.Source(processedContent)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, formattedContent, 0644)
}

// ToSnakeCase converts a camelCase or PascalCase string to snake_case.
func ToSnakeCase(s string) string {
	var result []rune
	for i, c := range s {
		if i > 0 && 'A' <= c && c <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, c)
	}
	return strings.ToLower(string(result))
}

// FindFunction uses Comby to locate a function or method in a directory.
// Returns the file path and nil error if found, empty string and nil if not found.
func FindFunction(directory, receiver, functionName string) (foundFilePath string, err error) {
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			found, err := findFunctionInFile(path, receiver, functionName)
			if err != nil {
				return err
			}
			if found {
				foundFilePath = path
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil && err != filepath.SkipDir {
		return "", err
	}
	return foundFilePath, nil
}

func findFunctionInFile(filePath, receiver, functionName string) (bool, error) {
	type matchRewrite struct {
		Match    string
		Rewrite  string
		Rule     string
		Continue bool
	}

	mrs := []matchRewrite{}

	if receiver != "" {
		mrs = append(mrs,
			matchRewrite{
				Match:   fmt.Sprintf("func (:[receiver]%s) %s(:[c])", receiver, functionName),
				Rewrite: "nop",
			},
			matchRewrite{
				Match:   fmt.Sprintf("func (:[receiver]*%s) %s(:[c])", receiver, functionName),
				Rewrite: "nop",
			},
		)
	} else {
		mrs = append(mrs,
			matchRewrite{
				Match:   fmt.Sprintf("func %s(:[c])", functionName),
				Rewrite: "nop",
			},
		)
	}

	for _, mr := range mrs {
		args := []string{mr.Match, mr.Rewrite, filePath, "-json-lines", "-match-newline-at-toplevel", "-matcher", ".go"}
		if mr.Rule != "" {
			args = append(args, "-rule", mr.Rule)
		}
		cmd := exec.Command("comby", args...)

		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		if err != nil {
			return false, fmt.Errorf("failed to run comby: %w, output: %s", err, string(output))
		}

		if outputStr != "" {
			expectedOutput := &struct {
				RewrittenSource string `json:"rewritten_source"`
			}{}
			if err := json.Unmarshal(output, expectedOutput); err != nil {
				return false, fmt.Errorf("failed to unmarshal comby output: %w", err)
			}
			if expectedOutput.RewrittenSource == "" {
				continue
			}
			return true, nil
		}
	}

	return false, nil
}

func parseDeclarationsFromCodeFragment(f CodeFragment) ([]ast.Decl, error) {
	code := strings.TrimSpace(f.Content)
	if !strings.HasPrefix(code, "package") {
		code = "package main\n\n" + code
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", code, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return file.Decls, nil
}

func upsertDeclaration(file *ast.File, newDecl ast.Decl, overwrite bool) {
	shouldAppend := true

	astutil.Apply(file, func(c *astutil.Cursor) bool {
		decl, ok := c.Node().(ast.Decl)
		if !ok {
			return true
		}

		switch existing := decl.(type) {
		case *ast.GenDecl:
			if existing.Tok == token.IMPORT {
				if newGenDecl, ok := newDecl.(*ast.GenDecl); ok && newGenDecl.Tok == token.IMPORT {
					existing.Specs = append(existing.Specs, newGenDecl.Specs...)
					shouldAppend = false
					return false
				}
				return true
			}

			if ts, ok := existing.Specs[0].(*ast.TypeSpec); ok {
				if ts.Name.Name == getDeclName(newDecl) {
					if overwrite {
						c.Replace(newDecl)
					}
					shouldAppend = false
					return false
				}
			}

		case *ast.FuncDecl:
			newFunc, ok := newDecl.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if existing.Name.Name == newFunc.Name.Name {
				existingRecv := getReceiverType(existing)
				newRecv := getReceiverType(newFunc)
				if existingRecv == newRecv {
					if overwrite {
						c.Replace(newDecl)
					}
					shouldAppend = false
					return false
				}
			}
		}

		return true
	}, nil)

	if shouldAppend {
		if newGenDecl, ok := newDecl.(*ast.GenDecl); ok && newGenDecl.Tok == token.IMPORT {
			for _, decl := range file.Decls {
				if existingImport, ok := decl.(*ast.GenDecl); ok && existingImport.Tok == token.IMPORT {
					existingImport.Specs = append(existingImport.Specs, newGenDecl.Specs...)
					return
				}
			}
		}
		file.Decls = append(file.Decls, newDecl)
	}
}

func getReceiverType(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		if starExpr, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				return ident.Name
			}
		} else if ident, ok := funcDecl.Recv.List[0].Type.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

func getDeclName(decl ast.Decl) string {
	if fd, ok := decl.(*ast.FuncDecl); ok {
		return fd.Name.Name
	}
	if gd, ok := decl.(*ast.GenDecl); ok {
		if ts, ok := gd.Specs[0].(*ast.TypeSpec); ok {
			return ts.Name.Name
		}
	}
	return ""
}

func writeChangesToFile(filePath string, fset *token.FileSet, node ast.Node) error {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return err
	}

	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, formattedCode, 0644)
}

func renderModifiedNode(fset *token.FileSet, node ast.Node) (string, error) {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// EnsureGoFileExists creates an empty Go file with the given package declaration if it
// does not already exist.
func EnsureGoFileExists(filename string, packageName string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if f, err := os.Create(filename); err != nil {
			return fmt.Errorf("Failed to create file: %v", err)
		} else {
			f.Write([]byte("package " + packageName + "\n"))
			defer f.Close()
		}
	}
	return nil
}

// FormatWithGoImports runs goimports on the given file, updating it in place.
func FormatWithGoImports(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filename)
	}

	cmd := exec.Command("goimports", "-w", filename)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run goimports: %v, stderr: %s", err, stderr.String())
	}

	return nil
}
