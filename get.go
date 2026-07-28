package main

import (
	"fmt"
	"os"
)

func runGet(args []string) {
	path := "$"
	if len(args) > 0 && args[0] != "" {
		path = args[0]
	}

	data, err := readStdin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	result, err := extractPath(data, path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Pretty-print the result. Indenting the raw bytes rewrites whitespace only,
	// so field order and number formatting survive untouched.
	output, err := indentRaw(result)
	if err != nil {
		// Not valid JSON (e.g. raw string), print as-is
		fmt.Println(string(result))
		return
	}

	fmt.Println(output)
}
