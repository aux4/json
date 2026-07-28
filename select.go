package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// selectField is a parsed node of the structure spec. A node with a non-empty
// group is a nested object selection (field[sub1,sub2]); otherwise it is a leaf.
type selectField struct {
	field string
	name  string
	group []selectField
}

var nestPattern = regexp.MustCompile(`^([^\[]+)\[(.+)\]$`)

// parseStructure parses the shared 2table/render field notation:
//
//	name,age,city                     simple fields
//	address[street,city]              nested object selection
//	name:fullName                     rename (source:output)
//	total{format:currency}            display formatting is ignored by select
func parseStructure(spec string) []selectField {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil
	}
	return parseSelectFields(spec)
}

func parseSelectFields(s string) []selectField {
	fields := []selectField{}
	current := strings.Builder{}
	bracket, brace := 0, 0

	flush := func() {
		f := strings.TrimSpace(current.String())
		if f != "" {
			fields = append(fields, parseSelectField(f))
		}
		current.Reset()
	}

	for _, ch := range s {
		switch ch {
		case '[':
			bracket++
		case ']':
			bracket--
		case '{':
			brace++
		case '}':
			brace--
		case ',':
			if bracket == 0 && brace == 0 {
				flush()
				continue
			}
		}
		current.WriteRune(ch)
	}
	flush()
	return fields
}

func parseSelectField(fieldString string) selectField {
	// Strip a trailing {format} block -- display-only, not relevant to a projection.
	if i := strings.LastIndex(fieldString, "{"); i >= 0 && strings.HasSuffix(fieldString, "}") {
		fieldString = strings.TrimSpace(fieldString[:i])
	}

	// Rename: source:output, at the first top-level colon (outside brackets).
	field := fieldString
	name := ""
	depth := 0
	for i, ch := range fieldString {
		if ch == '[' {
			depth++
		} else if ch == ']' {
			depth--
		} else if ch == ':' && depth == 0 {
			field = strings.TrimSpace(fieldString[:i])
			name = strings.TrimSpace(fieldString[i+1:])
			break
		}
	}

	// Nested object selection: field[sub1,sub2].
	if m := nestPattern.FindStringSubmatch(field); m != nil {
		fieldName := strings.TrimSpace(m[1])
		if name == "" {
			name = fieldName
		}
		return selectField{field: fieldName, name: name, group: parseSelectFields(m[2])}
	}

	if name == "" {
		name = field
	}
	return selectField{field: field, name: name}
}

// projectRaw builds a new object holding only the selected fields, in the order
// the structure spec listed them. A projection chooses fields *and their order*
// -- the 2table/render notation it borrows is an ordered spec -- so the output
// follows the spec, and values are carried across as raw bytes so nested objects
// keep their own field order too.
//
// Missing fields become null; nested groups recurse into objects and map over
// arrays of objects.
func projectRaw(raw json.RawMessage, fields []selectField) json.RawMessage {
	result := newOrderedObject()

	for _, f := range fields {
		value, ok := resolveRaw(raw, strings.Split(f.field, "."))
		if !ok {
			result.Set(f.name, json.RawMessage("null"))
			continue
		}

		if len(f.group) == 0 {
			result.Set(f.name, value)
			continue
		}

		trimmed := bytes.TrimSpace(value)
		switch {
		case len(trimmed) > 0 && trimmed[0] == '[':
			var arr []json.RawMessage
			if err := json.Unmarshal(trimmed, &arr); err != nil {
				result.Set(f.name, json.RawMessage("null"))
				continue
			}
			projected := make([]json.RawMessage, 0, len(arr))
			for _, el := range arr {
				projected = append(projected, projectRaw(el, f.group))
			}
			result.Set(f.name, rawArray(projected))
		case len(trimmed) > 0 && trimmed[0] == '{':
			result.Set(f.name, projectRaw(trimmed, f.group))
		default:
			result.Set(f.name, json.RawMessage("null"))
		}
	}

	return result.Bytes()
}

// runSelect projects each record to the selected fields. Without --inputStream it
// reads a JSON array (or object) and emits projected JSON; with --inputStream it
// reads NDJSON and emits one projected object per line.
func runSelect(args []string) {
	structure := ""
	if len(args) > 0 {
		structure = args[0]
	}
	inputStream := len(args) > 1 && args[1] == "true"

	if strings.TrimSpace(structure) == "" {
		fmt.Fprintln(os.Stderr, "structure is required")
		os.Exit(1)
	}
	fields := parseStructure(structure)

	if inputStream {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var obj json.RawMessage
			dec := json.NewDecoder(strings.NewReader(line))
			dec.UseNumber()
			if err := dec.Decode(&obj); err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid JSON: %s\n", line)
				os.Exit(1)
			}
			fmt.Println(string(projectRaw(obj, fields)))
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		return
	}

	dec := json.NewDecoder(bufio.NewReaderSize(os.Stdin, 1024*1024))
	dec.UseNumber()
	var input json.RawMessage
	if err := dec.Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	trimmed := bytes.TrimSpace(input)
	switch {
	case len(trimmed) > 0 && trimmed[0] == '[':
		var arr []json.RawMessage
		if err := json.Unmarshal(trimmed, &arr); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		projected := make([]json.RawMessage, 0, len(arr))
		for _, el := range arr {
			projected = append(projected, projectRaw(el, fields))
		}
		fmt.Println(string(rawArray(projected)))
	case len(trimmed) > 0 && trimmed[0] == '{':
		fmt.Println(string(projectRaw(trimmed, fields)))
	default:
		fmt.Fprintln(os.Stderr, "Error: expected a JSON object or array")
		os.Exit(1)
	}
}
