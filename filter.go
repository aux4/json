package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// runFilter keeps the records of a JSON array whose --field satisfies the
// predicate --op against --value. Numeric comparison is used when both sides are
// numbers, otherwise string comparison. Records missing the field are dropped.
// Field order within the kept records is preserved.
func runFilter(args []string) {
	fs := flag.NewFlagSet("filter", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	field := fs.String("field", "", "dot-path to test on each record")
	op := fs.String("op", "eq", "operator: eq|ne|gt|lt|gte|lte|contains")
	value := fs.String("value", "", "value to compare against")
	inputStream := fs.String("inputStream", "false", "read NDJSON from stdin")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if strings.TrimSpace(*field) == "" {
		fmt.Fprintln(os.Stderr, "Error: --field is required")
		os.Exit(1)
	}

	switch *op {
	case "eq", "ne", "gt", "lt", "gte", "lte", "contains":
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid op '%s' (use eq|ne|gt|lt|gte|lte|contains)\n", *op)
		os.Exit(1)
	}

	records, err := readRecords(*inputStream == "true")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	path := strings.Split(*field, ".")
	filtered := make([]json.RawMessage, 0, len(records))
	for _, rec := range records {
		if matchRecord(rec, path, *op, *value) {
			filtered = append(filtered, rec)
		}
	}

	output, err := indentRaw(rawArray(filtered))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(output)
}

// matchRecord evaluates the predicate for one record. A record that does not
// have the field is dropped for every operator.
func matchRecord(rec json.RawMessage, path []string, op, value string) bool {
	fv, ok := resolveRaw(rec, path)
	if !ok {
		return false
	}

	switch op {
	case "eq":
		return valuesEqual(fv, value)
	case "ne":
		return !valuesEqual(fv, value)
	case "gt", "lt", "gte", "lte":
		c := compareToValue(fv, value)
		switch op {
		case "gt":
			return c > 0
		case "lt":
			return c < 0
		case "gte":
			return c >= 0
		case "lte":
			return c <= 0
		}
	case "contains":
		return containsValue(fv, value)
	}
	return false
}

// valuesEqual compares a raw field value to the string --value, numerically when
// both are numbers and lexically otherwise.
func valuesEqual(fv json.RawMessage, value string) bool {
	if fn, ok := asNumber(fv); ok {
		if vn, err := strconv.ParseFloat(value, 64); err == nil {
			return fn == vn
		}
	}
	return rawText(fv) == value
}

// compareToValue returns -1, 0 or 1 comparing a raw field value to --value,
// numerically when both are numbers and lexically otherwise.
func compareToValue(fv json.RawMessage, value string) int {
	if fn, ok := asNumber(fv); ok {
		if vn, err := strconv.ParseFloat(value, 64); err == nil {
			switch {
			case fn < vn:
				return -1
			case fn > vn:
				return 1
			default:
				return 0
			}
		}
	}
	return strings.Compare(rawText(fv), value)
}

// containsValue is substring matching for strings and membership for arrays.
func containsValue(fv json.RawMessage, value string) bool {
	trimmed := bytes.TrimSpace(fv)
	if len(trimmed) == 0 {
		return false
	}
	switch trimmed[0] {
	case '"':
		return strings.Contains(rawText(fv), value)
	case '[':
		var arr []json.RawMessage
		if err := json.Unmarshal(trimmed, &arr); err != nil {
			return false
		}
		for _, el := range arr {
			if valuesEqual(el, value) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
