package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func runCollect(args []string) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	first := true
	fmt.Print("[")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Validate JSON
		if !json.Valid([]byte(line)) {
			fmt.Fprintf(os.Stderr, "Error: invalid JSON: %s\n", line)
			continue
		}

		if !first {
			fmt.Print(",")
		}
		fmt.Print(line)
		first = false
	}

	fmt.Println("]")

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}
