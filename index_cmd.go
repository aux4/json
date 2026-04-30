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

	index := make(map[string]interface{})

	err := streamArray(os.Stdin, func(raw json.RawMessage) error {
		var item map[string]interface{}
		if err := json.Unmarshal(raw, &item); err != nil {
			return fmt.Errorf("parsing item: %w", err)
		}

		key := generateId(item, fields)

		if existing, ok := index[key]; ok {
			// Convert to array if not already
			switch v := existing.(type) {
			case []interface{}:
				index[key] = append(v, item)
			default:
				index[key] = []interface{}{v, item}
			}
		} else {
			index[key] = item
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	output, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}

func generateId(item map[string]interface{}, fields []string) string {
	parts := make([]string, len(fields))
	for i, field := range fields {
		if val, ok := item[field]; ok {
			parts[i] = fmt.Sprintf("%v", val)
		}
	}
	return strings.Join(parts, "|")
}
