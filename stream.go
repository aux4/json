package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var errDone = errors.New("done")

// streamArray streams elements from a JSON array using token-level parsing.
// The callback receives each element as a raw JSON message.
// Return errDone from the callback to stop early.
func streamArray(reader io.Reader, callback func(json.RawMessage) error) error {
	dec := json.NewDecoder(reader)
	dec.UseNumber()

	t, err := dec.Token()
	if err != nil {
		return fmt.Errorf("reading opening token: %w", err)
	}
	if delim, ok := t.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("expected JSON array, got %v", t)
	}

	for dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return fmt.Errorf("decoding array element: %w", err)
		}
		if err := callback(raw); err != nil {
			if errors.Is(err, errDone) {
				return nil
			}
			return err
		}
	}

	return nil
}

// skipValue skips a single JSON value in the decoder without allocating memory.
// It tracks brace/bracket depth to skip objects and arrays entirely.
func skipValue(dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}

	switch v := t.(type) {
	case json.Delim:
		if v == '{' || v == '[' {
			closing := '}'
			if v == '[' {
				closing = ']'
			}
			for dec.More() {
				if v == '{' {
					// skip key
					if _, err := dec.Token(); err != nil {
						return err
					}
				}
				if err := skipValue(dec); err != nil {
					return err
				}
			}
			// read closing delimiter
			ct, err := dec.Token()
			if err != nil {
				return err
			}
			if d, ok := ct.(json.Delim); !ok || rune(d) != closing {
				return fmt.Errorf("expected closing %c", closing)
			}
		}
	}
	return nil
}

// readStdin reads all of stdin into memory.
func readStdin() ([]byte, error) {
	return io.ReadAll(os.Stdin)
}

// readRecords reads array records from stdin. Without inputStream it expects a
// single JSON array; with inputStream it reads NDJSON, one JSON value per line.
// Records are returned as raw messages so field order and number formatting are
// preserved untouched.
func readRecords(inputStream bool) ([]json.RawMessage, error) {
	if inputStream {
		records := []json.RawMessage{}
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var raw json.RawMessage
			dec := json.NewDecoder(strings.NewReader(line))
			dec.UseNumber()
			if err := dec.Decode(&raw); err != nil {
				return nil, fmt.Errorf("invalid JSON: %s", line)
			}
			records = append(records, raw)
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return records, nil
	}

	dec := json.NewDecoder(bufio.NewReaderSize(os.Stdin, 1024*1024))
	dec.UseNumber()
	var input json.RawMessage
	if err := dec.Decode(&input); err != nil {
		return nil, err
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(input, &arr); err != nil {
		return nil, fmt.Errorf("expected a JSON array")
	}
	return arr, nil
}

// extractPath navigates a dot-separated path through JSON data.
// Supports paths like "$.foo.bar", "foo.bar", "foo.0.name" (array index),
// negative indices that count from the end, e.g. "foo.-1.name" (last element),
// and a "*" wildcard segment that projects across an array.
//
// A "*" segment maps every element of the array at that point:
//   - "$.users.*"      returns the array elements themselves (["a","b"]).
//   - "$.users.*.name" returns a flat array of each element's name
//     (["Alice","Bob"]). Elements that do not have the remaining path are
//     skipped, so the result only contains values that resolve cleanly.
//
// Wildcards compose: any number of "*" segments may appear in a path.
func extractPath(data []byte, path string) ([]byte, error) {
	// Strip leading $ and .
	path = strings.TrimPrefix(path, "$")
	path = strings.TrimPrefix(path, ".")

	if path == "" {
		return data, nil
	}

	parts := strings.Split(path, ".")
	return extractParts(data, parts)
}

// extractParts walks the remaining path segments through current. It is
// recursive so a "*" segment can project across an array and then map the rest
// of the path over each element.
func extractParts(current []byte, parts []string) ([]byte, error) {
	for i, part := range parts {
		if part == "*" {
			var arr []json.RawMessage
			if err := json.Unmarshal(current, &arr); err != nil {
				return nil, fmt.Errorf("expected array at '*': %w", err)
			}
			remaining := parts[i+1:]
			if len(remaining) == 0 {
				// No further path: return the elements as an array.
				return rawArray(arr), nil
			}
			// Map each element through the remaining path, skipping any element
			// that does not resolve (missing field, wrong shape, out of bounds).
			projected := make([]json.RawMessage, 0, len(arr))
			for _, el := range arr {
				val, err := extractParts(el, remaining)
				if err != nil {
					continue
				}
				projected = append(projected, val)
			}
			return rawArray(projected), nil
		}

		// Try as array index
		if idx, err := strconv.Atoi(part); err == nil {
			var arr []json.RawMessage
			if err := json.Unmarshal(current, &arr); err != nil {
				return nil, fmt.Errorf("expected array at '%s': %w", part, err)
			}
			// Negative indices count from the end: -1 is the last element.
			resolved := idx
			if resolved < 0 {
				resolved += len(arr)
			}
			if resolved < 0 || resolved >= len(arr) {
				return nil, fmt.Errorf("array index %d out of bounds (length %d)", idx, len(arr))
			}
			current = arr[resolved]
		} else {
			// Object field
			var obj map[string]json.RawMessage
			if err := json.Unmarshal(current, &obj); err != nil {
				return nil, fmt.Errorf("expected object at '%s': %w", part, err)
			}
			val, ok := obj[part]
			if !ok {
				return nil, fmt.Errorf("field '%s' not found", part)
			}
			current = val
		}
	}

	return current, nil
}
