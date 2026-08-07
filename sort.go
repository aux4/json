package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

// stringSlice collects a repeatable string flag (e.g. --by name --by age).
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ",") }

func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

// runSort sorts a JSON array. Records are compared by one or more dot-path keys
// (--by, repeatable). When both values at a key are numbers they compare
// numerically, otherwise lexically. With no --by the whole element is compared,
// which sorts arrays of scalars. Sorting is stable, so records that compare
// equal keep their input order. Field order within records is preserved.
func runSort(args []string) {
	fs := flag.NewFlagSet("sort", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var by stringSlice
	fs.Var(&by, "by", "field to sort by (dot-path, repeatable)")
	order := fs.String("order", "asc", "sort order: asc or desc")
	inputStream := fs.String("inputStream", "false", "read NDJSON from stdin")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	desc := false
	switch strings.ToLower(*order) {
	case "asc", "":
		desc = false
	case "desc":
		desc = true
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid order '%s' (use asc or desc)\n", *order)
		os.Exit(1)
	}

	// Drop empty keys so an empty --by (e.g. its default) falls back to
	// comparing whole elements, which sorts arrays of scalars.
	keys := make([]string, 0, len(by))
	for _, k := range by {
		if strings.TrimSpace(k) != "" {
			keys = append(keys, k)
		}
	}

	records, err := readRecords(*inputStream == "true")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	sort.SliceStable(records, func(i, j int) bool {
		c := compareRecords(records[i], records[j], keys)
		if desc {
			return c > 0
		}
		return c < 0
	})

	output, err := indentRaw(rawArray(records))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(output)
}

// compareRecords compares two records across the given keys, returning the first
// non-zero comparison. With no keys the whole records are compared.
func compareRecords(a, b json.RawMessage, keys []string) int {
	if len(keys) == 0 {
		return compareValues(a, true, b, true)
	}
	for _, key := range keys {
		path := strings.Split(key, ".")
		av, aok := resolveRaw(a, path)
		bv, bok := resolveRaw(b, path)
		if c := compareValues(av, aok, bv, bok); c != 0 {
			return c
		}
	}
	return 0
}

// compareValues orders two raw values. Missing values sort after present ones.
// Two numbers compare numerically; anything else compares as text.
func compareValues(a json.RawMessage, aok bool, b json.RawMessage, bok bool) int {
	switch {
	case !aok && !bok:
		return 0
	case !aok:
		return 1
	case !bok:
		return -1
	}

	an, aIsNum := asNumber(a)
	bn, bIsNum := asNumber(b)
	if aIsNum && bIsNum {
		switch {
		case an < bn:
			return -1
		case an > bn:
			return 1
		default:
			return 0
		}
	}
	return strings.Compare(rawText(a), rawText(b))
}
