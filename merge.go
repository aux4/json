package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// fileIndex holds one file's records keyed by id, remembering the order the
// keys first appeared so the merged output is stable between runs.
type fileIndex struct {
	order   []string
	records map[string][]*orderedObject
}

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

	indexes := make([]*fileIndex, 0, len(files))

	for _, filePath := range files {
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		var items []json.RawMessage
		if err := json.Unmarshal(data, &items); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		idx := &fileIndex{records: map[string][]*orderedObject{}}
		for _, raw := range items {
			item, err := parseOrderedObject(raw)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing file %s: %v\n", filePath, err)
				os.Exit(1)
			}
			key := generateId(item, fields)
			if _, seen := idx.records[key]; !seen {
				idx.order = append(idx.order, key)
			}
			idx.records[key] = append(idx.records[key], item)
		}
		indexes = append(indexes, idx)
	}

	// Walk the first file in its own record order rather than ranging over a map,
	// which would shuffle the output between runs.
	merged := []json.RawMessage{}

	for _, key := range indexes[0].order {
		for _, baseItem := range indexes[0].records[key] {
			for _, result := range mergeItem(baseItem, key, indexes[1:]) {
				merged = append(merged, result.Bytes())
			}
		}
	}

	output, err := indentRaw(rawArray(merged))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(output)
}

// mergeItem folds each later file's matching records into the base record.
// Fields keep the base record's order, with any field a later file introduces
// appended in the order that file listed it.
func mergeItem(base *orderedObject, key string, indexes []*fileIndex) []*orderedObject {
	results := []*orderedObject{copyOrdered(base)}

	for _, idx := range indexes {
		found, ok := idx.records[key]
		if !ok || len(found) == 0 {
			continue
		}

		if len(found) == 1 {
			for i := range results {
				overlay(results[i], found[0])
			}
			continue
		}

		var expanded []*orderedObject
		for _, item := range found {
			for _, result := range results {
				merged := copyOrdered(result)
				overlay(merged, item)
				expanded = append(expanded, merged)
			}
		}
		results = expanded
	}

	return results
}

func copyOrdered(o *orderedObject) *orderedObject {
	cp := newOrderedObject()
	for _, key := range o.Keys() {
		value, _ := o.Get(key)
		cp.Set(key, value)
	}
	return cp
}

func overlay(dst, src *orderedObject) {
	for _, key := range src.Keys() {
		value, _ := src.Get(key)
		dst.Set(key, value)
	}
}
