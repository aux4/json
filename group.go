package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// groupState accumulates one group. Non-id fields are recorded in the order
// they were first seen in the input rather than by ranging over a Go map, which
// would make the output order differ between runs.
type groupState struct {
	idValues   map[string]json.RawMessage
	valueOrder []string
	values     map[string][]json.RawMessage
	seen       map[string]map[string]bool
}

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

	groups := map[string]*groupState{}
	var keyOrder []string

	err := streamArray(os.Stdin, func(raw json.RawMessage) error {
		item, err := parseOrderedObject(raw)
		if err != nil {
			return fmt.Errorf("parsing item: %w", err)
		}

		key := generateId(item, fields)
		group, exists := groups[key]
		if !exists {
			group = &groupState{
				idValues: map[string]json.RawMessage{},
				values:   map[string][]json.RawMessage{},
				seen:     map[string]map[string]bool{},
			}
			groups[key] = group
			keyOrder = append(keyOrder, key)

			for _, f := range fields {
				if value, ok := item.Get(f); ok {
					group.idValues[f] = value
				}
			}
		}

		for _, k := range item.Keys() {
			if fieldSet[k] {
				continue
			}
			value, _ := item.Get(k)
			if _, known := group.seen[k]; !known {
				group.valueOrder = append(group.valueOrder, k)
				group.seen[k] = map[string]bool{}
			}
			text := string(value)
			if group.seen[k][text] {
				continue
			}
			group.seen[k][text] = true
			group.values[k] = append(group.values[k], value)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	grouped := make([]json.RawMessage, 0, len(keyOrder))
	for _, key := range keyOrder {
		group := groups[key]
		obj := newOrderedObject()

		// Id fields first, in the order they were requested.
		for _, f := range fields {
			if value, ok := group.idValues[f]; ok {
				obj.Set(f, value)
			} else {
				obj.Set(f, json.RawMessage("null"))
			}
		}
		for _, k := range group.valueOrder {
			obj.Set(k, rawArray(group.values[k]))
		}

		grouped = append(grouped, obj.Bytes())
	}

	output, err := indentRaw(rawArray(grouped))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(output)
}
