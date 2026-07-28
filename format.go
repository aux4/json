package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// formatStream re-indents or minifies every JSON value on stdin.
//
// It rewrites the raw bytes with json.Indent/json.Compact rather than
// unmarshalling into interface{} and marshalling back. That round trip turns
// objects into Go maps, and marshalling a map sorts its keys -- so it would
// silently reorder every object it touched. Field order is data: it is the
// order the producer chose, and the thing you scan to check output against its
// source. Rewriting only the whitespace leaves it intact.
//
// Reading value by value also means a stream of JSON documents (NDJSON) goes
// through the same path as a single document.
func formatStream(pretty bool) {
	dec := json.NewDecoder(bufio.NewReaderSize(os.Stdin, 1024*1024))
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	values := 0
	for dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			out.Flush()
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
			os.Exit(1)
		}

		var buf bytes.Buffer
		var err error
		if pretty {
			err = json.Indent(&buf, raw, "", "  ")
		} else {
			err = json.Compact(&buf, raw)
		}
		if err != nil {
			out.Flush()
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintln(out, buf.String())
		values++
	}

	if values == 0 {
		fmt.Fprintln(os.Stderr, "Error: no JSON input")
		os.Exit(1)
	}
}

func runPretty(args []string) {
	formatStream(true)
}

func runInline(args []string) {
	formatStream(false)
}
