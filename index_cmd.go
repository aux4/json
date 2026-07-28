package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func runIndex(args []string) {
	if len(args) < 1 || args[0] == "" {
		fmt.Fprintf(os.Stderr, "Error: id parameter is required\n")
		os.Exit(1)
	}

	id := args[0]
	fields := strings.Split(id, ",")

	// Records are kept as raw bytes and the index keeps its insertion order, so
	// both the field order inside each record and the order the keys appeared
	// in the input survive to the output.
	index := newOrderedObject()
	duplicates := map[string][]json.RawMessage{}

	err := streamArray(os.Stdin, func(raw json.RawMessage) error {
		item, err := parseOrderedObject(raw)
		if err != nil {
			return fmt.Errorf("parsing item: %w", err)
		}

		key := generateId(item, fields)
		record := append(json.RawMessage(nil), raw...)

		if _, seen := index.Get(key); seen {
			duplicates[key] = append(duplicates[key], record)
			index.Set(key, rawArray(duplicates[key]))
			return nil
		}

		duplicates[key] = []json.RawMessage{record}
		index.Set(key, record)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	output, err := indentRaw(index.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(output)
}

func generateId(item *orderedObject, fields []string) string {
	parts := make([]string, len(fields))
	for i, field := range fields {
		if val, ok := item.Get(field); ok {
			parts[i] = rawText(val)
		}
	}
	return strings.Join(parts, "|")
}
