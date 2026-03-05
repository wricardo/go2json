package main

import (
	"fmt"
	"log"

	g2j "github.com/wricardo/go2json"
)

func main() {
	fmt.Println("Example 02: Parse directory recursively and output as JSON")

	// Parse the current directory recursively
	parsed, err := g2j.ParseDirectoryRecursive(".")
	if err != nil {
		log.Fatalf("Failed to parse directory: %v", err)
	}

	// Print the results in JSON format
	output := g2j.PrettyPrint(
		parsed,
		"json",                    // format
		nil,                       // ignore rules
		true,                      // plain-structs
		true,                      // fields-plain-structs
		true,                      // structs-with-method
		true,                      // fields-structs-with-method
		true,                      // methods
		true,                      // functions
		true,                      // tags
		true,                      // comments
		false,                     // omit-nulls
	)
	fmt.Println(output)
}
