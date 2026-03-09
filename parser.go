package go2json

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// ParseFile parses a Go file or directory and returns the parsed information.
//
// Deprecated: Use ParseDirectory instead.
func ParseFile(fileOrDirectory string) (*ParsedInfo, error) {
	return ParseDirectory(fileOrDirectory)
}

// ParseDirectory parses a directory containing Go files and returns the parsed information.
func ParseDirectory(fileOrDirectory string) (*ParsedInfo, error) {
	return ParseDirectoryWithFilter(fileOrDirectory, nil)
}

// ParseString parses Go source code provided as a string and returns the parsed information.
func ParseString(fileContent string) (*ParsedInfo, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", fileContent, parser.ParseComments|parser.AllErrors|parser.DeclarationErrors)
	if err != nil {
		return nil, err
	}

	packages := map[string]*ast.Package{
		"": {
			Name:  file.Name.Name,
			Files: map[string]*ast.File{"": file},
		},
	}

	return extractParsedInfo(packages, "", "")
}

// ParseDirectoryRecursive parses a directory recursively and returns the parsed information.
func ParseDirectoryRecursive(path string) ([]*ParsedInfo, error) {
	var results []*ParsedInfo

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.Contains(p, ".git") {
				return nil
			}
			parsed, err := ParseDirectoryWithFilter(p, func(info fs.FileInfo) bool {
				return true
			})
			if err != nil {
				return err
			}
			results = append(results, parsed)
		}
		return nil
	})
	return results, err
}

// ParseDirectoryWithFilter parses a directory with an optional filter function to include specific files.
func ParseDirectoryWithFilter(fileOrDirectory string, filter func(fs.FileInfo) bool) (*ParsedInfo, error) {
	fi, err := os.Stat(fileOrDirectory)
	if err != nil {
		return nil, err
	}

	var packages map[string]*ast.Package
	fset := token.NewFileSet()

	isDir := true
	switch mode := fi.Mode(); {
	case mode.IsDir():
		packages, err = parser.ParseDir(fset, fileOrDirectory, filter, parser.ParseComments|parser.AllErrors|parser.DeclarationErrors)
		if err != nil {
			return nil, err
		}
	case mode.IsRegular():
		isDir = false
		file, err := parser.ParseFile(fset, fileOrDirectory, nil, parser.ParseComments|parser.AllErrors|parser.DeclarationErrors)
		if err != nil {
			return nil, err
		}
		packages = map[string]*ast.Package{
			fileOrDirectory: {
				Name:  file.Name.Name,
				Files: map[string]*ast.File{fileOrDirectory: file},
			},
		}
	}

	modulePath, err := getModulePath(fileOrDirectory)
	if err != nil {
		return nil, fmt.Errorf("error retrieving module path: %w", err)
	}
	relPath, err := filepath.Rel(modulePath.Dir, fileOrDirectory)
	if err != nil {
		return nil, fmt.Errorf("error resolving relative path: %w", err)
	}

	parsedInfo, err := extractParsedInfo(packages, modulePath.Path, relPath)
	if err != nil {
		return nil, err
	}
	if isDir {
		parsedInfo.Directory = fileOrDirectory
		if abs, err := filepath.Abs(fileOrDirectory); err == nil {
			parsedInfo.Directory = abs
		}
	} else {
		parsedInfo.File = fileOrDirectory
		if abs, err := filepath.Abs(fileOrDirectory); err == nil {
			parsedInfo.File = abs
		}
	}
	return parsedInfo, nil
}

// getModulePath reads the module name from the nearest go.mod file.
func getModulePath(path string) (*struct {
	Path string
	Dir  string
}, error) {
	dir := path
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
			if err != nil {
				return nil, fmt.Errorf("error reading go.mod: %w", err)
			}
			modFile, err := modfile.Parse("go.mod", data, nil)
			if err != nil {
				return nil, fmt.Errorf("error parsing go.mod: %w", err)
			}
			return &struct {
				Path string
				Dir  string
			}{
				Path: modFile.Module.Mod.Path,
				Dir:  dir,
			}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
