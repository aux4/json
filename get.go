package main

import (
	"encoding/json"
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

	// Pretty-print the result
	var parsed interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		// Not valid JSON (e.g. raw string), print as-is
		fmt.Println(string(result))
		return
	}

	output, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		fmt.Println(string(result))
		return
	}

	fmt.Println(string(output))
}
