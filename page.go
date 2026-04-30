package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func runPage(args []string) {
	offset := 0
	limit := 10
	stream := false

	if len(args) > 0 && args[0] != "" {
		if v, err := strconv.Atoi(args[0]); err == nil {
			offset = v
		}
	}
	if len(args) > 1 && args[1] != "" {
		if v, err := strconv.Atoi(args[1]); err == nil {
			limit = v
		}
	}
	if len(args) > 2 && args[2] == "true" {
		stream = true
	}

	dec := json.NewDecoder(os.Stdin)
	dec.UseNumber()

	t, err := dec.Token()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		fmt.Fprintf(os.Stderr, "Error: expected JSON array\n")
		os.Exit(1)
	}

	// Skip offset items using token-level skipping (no allocation)
	for i := 0; i < offset && dec.More(); i++ {
		if err := skipValue(dec); err != nil {
			fmt.Fprintf(os.Stderr, "Error skipping: %v\n", err)
			os.Exit(1)
		}
	}

	// Collect limit items
	collected := 0
	if !stream {
		fmt.Print("[")
	}

	for collected < limit && dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading item: %v\n", err)
			os.Exit(1)
		}

		if stream {
			fmt.Println(string(raw))
		} else {
			if collected > 0 {
				fmt.Print(",")
			}
			fmt.Print(string(raw))
		}
		collected++
	}

	if !stream {
		fmt.Println("]")
	}
}
