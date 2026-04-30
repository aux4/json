package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func runPretty(args []string) {
	data, err := readStdin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	output, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}

func runInline(args []string) {
	data, err := readStdin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	output, err := json.Marshal(parsed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}
