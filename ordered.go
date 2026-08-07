package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// orderedObject is a JSON object that keeps its members in insertion order.
//
// Go's encoding/json marshals a map with its keys sorted, so any command that
// parsed a record into map[string]interface{} and marshalled it back silently
// rewrote the field order of every object it touched. Field order is data: it
// is the order the producer chose, and for `select` it is the order the caller
// asked for. Values stay as raw bytes so nested objects are never rewritten
// either, and numbers keep their exact source text.
type orderedObject struct {
	keys   []string
	values map[string]json.RawMessage
}

func newOrderedObject() *orderedObject {
	return &orderedObject{values: map[string]json.RawMessage{}}
}

// Set appends a new key or overwrites an existing one in place, keeping its
// original position.
func (o *orderedObject) Set(key string, value json.RawMessage) {
	if _, exists := o.values[key]; !exists {
		o.keys = append(o.keys, key)
	}
	o.values[key] = value
}

func (o *orderedObject) Get(key string) (json.RawMessage, bool) {
	value, ok := o.values[key]
	return value, ok
}

func (o *orderedObject) Keys() []string { return o.keys }

// Bytes renders the object compactly. It writes the raw member bytes directly
// rather than going through json.Marshal, which would HTML-escape the values
// and so alter strings containing < > or &.
func (o *orderedObject) Bytes() json.RawMessage {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range o.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		writeJSONString(&buf, key)
		buf.WriteByte(':')
		value := o.values[key]
		if len(bytes.TrimSpace(value)) == 0 {
			buf.WriteString("null")
		} else {
			buf.Write(value)
		}
	}
	buf.WriteByte('}')
	return buf.Bytes()
}

// writeJSONString encodes a string without HTML escaping.
func writeJSONString(buf *bytes.Buffer, s string) {
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	// Encode writes a trailing newline; trim it back off.
	before := buf.Len()
	if err := enc.Encode(s); err != nil {
		buf.Truncate(before)
		buf.WriteString(`""`)
		return
	}
	trimmed := bytes.TrimRight(buf.Bytes()[before:], "\n")
	buf.Truncate(before)
	buf.Write(trimmed)
}

// parseOrderedObject decodes a JSON object while keeping its key order.
func parseOrderedObject(raw []byte) (*orderedObject, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()

	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return nil, fmt.Errorf("expected a JSON object")
	}

	obj := newOrderedObject()
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := kt.(string)
		if !ok {
			return nil, fmt.Errorf("expected an object key, got %v", kt)
		}
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return nil, err
		}
		obj.Set(key, value)
	}
	return obj, nil
}

// rawArray joins raw values into a JSON array without touching their bytes.
func rawArray(items []json.RawMessage) json.RawMessage {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, item := range items {
		if i > 0 {
			buf.WriteByte(',')
		}
		if len(bytes.TrimSpace(item)) == 0 {
			buf.WriteString("null")
		} else {
			buf.Write(item)
		}
	}
	buf.WriteByte(']')
	return buf.Bytes()
}

// rawText renders a raw scalar the way a composite-key component should read:
// strings unquoted, everything else as written. Using the source text avoids
// the float formatting that %v on an unmarshalled interface{} produces, where a
// large integer id comes out as 1e+06.
func rawText(raw json.RawMessage) string {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return ""
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err == nil {
			return s
		}
	}
	return string(trimmed)
}

// asNumber reads a raw JSON value as a float64 when it is a JSON number.
// Quoted strings, booleans and null return ok=false so callers fall back to a
// string comparison. The exact source text is parsed so large integer ids are
// compared by value rather than by their %v float rendering.
func asNumber(raw json.RawMessage) (float64, bool) {
	t := string(bytes.TrimSpace(raw))
	if t == "" || t[0] == '"' {
		return 0, false
	}
	f, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// resolveRaw walks a dotted path through raw JSON and returns the raw value at
// the end of it, leaving that value's own bytes untouched.
func resolveRaw(raw json.RawMessage, path []string) (json.RawMessage, bool) {
	current := raw
	for _, key := range path {
		trimmed := bytes.TrimSpace(current)
		if len(trimmed) == 0 {
			return nil, false
		}
		switch trimmed[0] {
		case '{':
			obj, err := parseOrderedObject(trimmed)
			if err != nil {
				return nil, false
			}
			value, ok := obj.Get(key)
			if !ok {
				return nil, false
			}
			current = value
		case '[':
			var arr []json.RawMessage
			if err := json.Unmarshal(trimmed, &arr); err != nil {
				return nil, false
			}
			i, err := strconv.Atoi(key)
			if err != nil || i < 0 || i >= len(arr) {
				return nil, false
			}
			current = arr[i]
		default:
			return nil, false
		}
	}
	return current, true
}

// indentRaw re-indents raw JSON for display without reordering or re-encoding.
func indentRaw(raw json.RawMessage) (string, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return "", err
	}
	return buf.String(), nil
}
