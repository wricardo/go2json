// Package main implements the go2json CLI tool.
// go2json is a Go development tool for code analysis and AST parsing.
package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	g2j "github.com/wricardo/go2json"
)

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
						Value: "llm",
						Usage: "format to print the parsed information: llm, text_short, test_long, json",
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
						Value: true,
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
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
