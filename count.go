package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func runCount(args []string) {
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

	count := 0
	for dec.More() {
		if err := skipValue(dec); err != nil {
			fmt.Fprintf(os.Stderr, "Error counting: %v\n", err)
			os.Exit(1)
		}
		count++
	}

	fmt.Println(count)
}
