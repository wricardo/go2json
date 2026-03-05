// Package main implements the go2json CLI tool.
// go2json is a Go development tool for code analysis and AST parsing.
package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	g2j "github.com/wricardo/go2json"
)

//go:embed skills/go2json/SKILL.md
var skillFS embed.FS

func main() {
	app := &cli.App{
		Name:  "go2json",
		Usage: "A CLI tool to parse and analyze Go code",
		Action: func(*cli.Context) error {
			fmt.Println("--help for more information.")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "parse",
				Aliases: []string{"p"},
				Usage:   "parse a file or directory",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "path",
						Aliases:  []string{"f"},
						Usage:    "path to the file or directory to parse",
						Value:    ".",
						Required: false,
					},
					&cli.BoolFlag{
						Name:    "recursive",
						Aliases: []string{"r"},
						Usage:   "recursively parse directories",
					},
					&cli.StringFlag{
						Name:  "format",
						Value: "json",
						Usage: "format to print the parsed information: json, llm (Go-syntax), or grepindex",
					},
					&cli.BoolFlag{
						Name:  "omit-nulls",
						Usage: "omit null and empty values from JSON output",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "plain-structs",
						Usage: "print plain structs",
						Value: true,
					},
					&cli.BoolFlag{
						Name:  "fields-plain-structs",
						Usage: "print fields of plain structs",
						Value: true,
					},
					&cli.BoolFlag{
						Name:  "structs-with-method",
						Usage: "print structs with methods",
						Value: true,
					},
					&cli.BoolFlag{
						Name:  "fields-structs-with-method",
						Usage: "print fields of structs with methods",
						Value: true,
					},
					&cli.BoolFlag{
						Name:  "methods",
						Usage: "print methods",
						Value: true,
					},
					&cli.BoolFlag{
						Name:  "functions",
						Usage: "print functions",
						Value: true,
					},
					&cli.BoolFlag{
						Name:  "comments",
						Usage: "print comments",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "tags",
						Usage: "print tags",
						Value: true,
					},
					&cli.StringSliceFlag{
						Name:  "ignore-rule",
						Usage: "ignore files or directories that match the rule. ",
					},
				},
				Action: func(cCtx *cli.Context) error {
					path := cCtx.String("path")
					ignores := cCtx.StringSlice("ignore-rule")

					if cCtx.Bool("recursive") {
						parsed, err := g2j.ParseDirectoryRecursive(path)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Failed to parse directory: %v\n", err)
							os.Exit(1)
						}

						fmt.Println(g2j.PrettyPrint(parsed, cCtx.String("format"), ignores, cCtx.Bool("plain-structs"), cCtx.Bool("fields-plain-structs"), cCtx.Bool("structs-with-method"), cCtx.Bool("fields-structs-with-method"), cCtx.Bool("methods"), cCtx.Bool("functions"), cCtx.Bool("tags"), cCtx.Bool("comments"), cCtx.Bool("omit-nulls")))
					} else {
						parsed, err := g2j.ParseDirectoryWithFilter(path, nil)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Failed to parse directory: %v\n", err)
							os.Exit(1)
						}
						fmt.Println(g2j.PrettyPrint([]*g2j.ParsedInfo{parsed}, cCtx.String("format"), ignores, cCtx.Bool("plain-structs"), cCtx.Bool("fields-plain-structs"), cCtx.Bool("structs-with-method"), cCtx.Bool("fields-structs-with-method"), cCtx.Bool("methods"), cCtx.Bool("functions"), cCtx.Bool("tags"), cCtx.Bool("comments"), cCtx.Bool("omit-nulls")))
					}
					return nil
				},
			},
			{
				Name:    "describe",
				Aliases: []string{"d"},
				Usage:   "describe a type and its dependencies",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "type",
						Aliases:  []string{"t"},
						Usage:    "the type name to describe",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "path",
						Aliases: []string{"f"},
						Usage:   "path to the directory to parse",
						Value:   ".",
					},
					&cli.IntFlag{
						Name:    "depth",
						Aliases: []string{"n"},
						Usage:   "max depth of type dependency traversal",
						Value:   1,
					},
					&cli.StringFlag{
						Name:  "format",
						Usage: "output format: llm, json, grepindex",
						Value: "llm",
					},
					&cli.BoolFlag{
						Name:  "omit-nulls",
						Usage: "omit null and empty values from JSON output",
						Value: false,
					},
				},
				Action: func(cCtx *cli.Context) error {
					path := cCtx.String("path")
					parsed, err := g2j.ParseDirectoryRecursive(path)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to parse directory: %v\n", err)
						os.Exit(1)
					}
					result, err := g2j.DescribeType(cCtx.String("type"), parsed, cCtx.Int("depth"))
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
						os.Exit(1)
					}
					fmt.Println(g2j.PrettyPrint(result, cCtx.String("format"), nil, true, true, true, true, true, true, true, false, cCtx.Bool("omit-nulls")))
					return nil
				},
			},
			{
				Name:  "install-skill",
				Usage: "install the go2json Claude Code skill to ~/.claude/skills/go2json/",
				Action: func(cCtx *cli.Context) error {
					homeDir, err := os.UserHomeDir()
					if err != nil {
						return fmt.Errorf("could not determine home directory: %w", err)
					}

					destDir := filepath.Join(homeDir, ".claude", "skills", "go2json")
					if err := os.MkdirAll(destDir, 0755); err != nil {
						return fmt.Errorf("could not create directory %s: %w", destDir, err)
					}

					data, err := skillFS.ReadFile("skills/go2json/SKILL.md")
					if err != nil {
						return fmt.Errorf("could not read embedded skill file: %w", err)
					}

					destPath := filepath.Join(destDir, "SKILL.md")
					if err := os.WriteFile(destPath, data, 0644); err != nil {
						return fmt.Errorf("could not write skill file: %w", err)
					}

					fmt.Printf("Installed go2json skill to %s\n", destPath)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
