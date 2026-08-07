package main

import (
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

// extractPath navigates a dot-separated path through JSON data.
// Supports paths like "$.foo.bar", "foo.bar", "foo.0.name" (array index), and
// negative indices that count from the end, e.g. "foo.-1.name" (last element).
func extractPath(data []byte, path string) ([]byte, error) {
	// Strip leading $ and .
	path = strings.TrimPrefix(path, "$")
	path = strings.TrimPrefix(path, ".")

	if path == "" {
		return data, nil
	}

	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
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
