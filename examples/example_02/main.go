package main

import (
	"fmt"
	"log"

	"github.com/wricardo/go2json"
)

func main() {
	fmt.Println("Example 02: Parse directory recursively and output as JSON")

	// Parse the current directory recursively
	parsed, err := go2json.ParseDirectoryRecursive(".")
	if err != nil {
		log.Fatalf("Failed to parse directory: %v", err)
	}

	// Print the results in JSON format
	output := go2json.PrettyPrint(
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
