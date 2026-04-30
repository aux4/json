package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func runGroup(args []string) {
	if len(args) < 1 || args[0] == "" {
		fmt.Fprintf(os.Stderr, "Error: id parameter is required\n")
		os.Exit(1)
	}

	id := args[0]
	fields := strings.Split(id, ",")
	fieldSet := make(map[string]bool, len(fields))
	for _, f := range fields {
		fieldSet[f] = true
	}

	// Build index: key -> []item
	index := make(map[string][]map[string]interface{})
	var keyOrder []string

	err := streamArray(os.Stdin, func(raw json.RawMessage) error {
		var item map[string]interface{}
		if err := json.Unmarshal(raw, &item); err != nil {
			return fmt.Errorf("parsing item: %w", err)
		}

		key := generateId(item, fields)

		if _, exists := index[key]; !exists {
			keyOrder = append(keyOrder, key)
		}
		index[key] = append(index[key], item)

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Build grouped result
	var grouped []map[string]interface{}

	for _, key := range keyOrder {
		items := index[key]
		groupItem := make(map[string]interface{})

		// Set ID fields from first item
		for _, f := range fields {
			groupItem[f] = items[0][f]
		}

		// Collect non-ID fields into arrays (deduplicated)
		for _, item := range items {
			for k, v := range item {
				if fieldSet[k] {
					continue
				}
				arr, _ := groupItem[k].([]interface{})
				if !containsValue(arr, v) {
					groupItem[k] = append(arr, v)
				}
			}
		}

		grouped = append(grouped, groupItem)
	}

	output, err := json.MarshalIndent(grouped, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}

func containsValue(arr []interface{}, val interface{}) bool {
	for _, v := range arr {
		if fmt.Sprintf("%v", v) == fmt.Sprintf("%v", val) {
			return true
		}
	}
	return false
}
