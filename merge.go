package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func runMerge(args []string) {
	if len(args) < 1 || args[0] == "" {
		fmt.Fprintf(os.Stderr, "Error: id parameter is required\n")
		os.Exit(1)
	}

	id := args[0]
	fields := strings.Split(id, ",")

	// Files can be passed as a single space-separated string or as separate args
	var files []string
	for _, arg := range args[1:] {
		for _, f := range strings.Fields(arg) {
			if f != "" {
				files = append(files, f)
			}
		}
	}
	if len(files) < 2 {
		fmt.Fprintf(os.Stderr, "Error: at least 2 files are required\n")
		os.Exit(1)
	}

	// Read and index each file
	var indexes []map[string]interface{}

	for _, filePath := range files {
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		var items []map[string]interface{}
		if err := json.Unmarshal(data, &items); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		idx := make(map[string]interface{})
		for _, item := range items {
			key := generateId(item, fields)
			if existing, ok := idx[key]; ok {
				switch v := existing.(type) {
				case []map[string]interface{}:
					idx[key] = append(v, item)
				case map[string]interface{}:
					idx[key] = []map[string]interface{}{v, item}
				}
			} else {
				idx[key] = item
			}
		}
		indexes = append(indexes, idx)
	}

	// Merge: iterate first file's index, merge with remaining
	var merged []map[string]interface{}

	for key, value := range indexes[0] {
		var baseItems []map[string]interface{}
		switch v := value.(type) {
		case map[string]interface{}:
			baseItems = []map[string]interface{}{v}
		case []map[string]interface{}:
			baseItems = v
		}

		for _, baseItem := range baseItems {
			results := mergeItem(baseItem, key, indexes[1:])
			merged = append(merged, results...)
		}
	}

	output, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}

func mergeItem(base map[string]interface{}, key string, indexes []map[string]interface{}) []map[string]interface{} {
	results := []map[string]interface{}{copyMap(base)}

	for _, idx := range indexes {
		found, ok := idx[key]
		if !ok {
			continue
		}

		switch v := found.(type) {
		case map[string]interface{}:
			for i := range results {
				for k, val := range v {
					results[i][k] = val
				}
			}
		case []map[string]interface{}:
			var expanded []map[string]interface{}
			for _, item := range v {
				for _, result := range results {
					merged := copyMap(result)
					for k, val := range item {
						merged[k] = val
					}
					expanded = append(expanded, merged)
				}
			}
			results = expanded
		}
	}

	return results
}

func copyMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{}, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}
