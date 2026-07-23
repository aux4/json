package main

import (
	"bufio"
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

// projectObject builds a new object holding only the selected fields, preserving
// value types. Missing fields become null; nested groups recurse into objects
// and map over arrays of objects.
func projectObject(obj interface{}, fields []selectField) map[string]interface{} {
	result := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		if len(f.group) > 0 {
			nested := resolvePath(obj, strings.Split(f.field, "."))
			switch typed := nested.(type) {
			case []interface{}:
				arr := make([]interface{}, 0, len(typed))
				for _, el := range typed {
					arr = append(arr, projectObject(el, f.group))
				}
				result[f.name] = arr
			case map[string]interface{}:
				result[f.name] = projectObject(typed, f.group)
			default:
				result[f.name] = nil
			}
		} else {
			result[f.name] = resolvePath(obj, strings.Split(f.field, "."))
		}
	}
	return result
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
			var obj interface{}
			dec := json.NewDecoder(strings.NewReader(line))
			dec.UseNumber()
			if err := dec.Decode(&obj); err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid JSON: %s\n", line)
				os.Exit(1)
			}
			out, _ := json.Marshal(projectObject(obj, fields))
			fmt.Println(string(out))
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		return
	}

	dec := json.NewDecoder(bufio.NewReaderSize(os.Stdin, 1024*1024))
	dec.UseNumber()
	var input interface{}
	if err := dec.Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	switch typed := input.(type) {
	case []interface{}:
		arr := make([]interface{}, 0, len(typed))
		for _, el := range typed {
			arr = append(arr, projectObject(el, fields))
		}
		out, _ := json.Marshal(arr)
		fmt.Println(string(out))
	case map[string]interface{}:
		out, _ := json.Marshal(projectObject(typed, fields))
		fmt.Println(string(out))
	default:
		fmt.Fprintln(os.Stderr, "Error: expected a JSON object or array")
		os.Exit(1)
	}
}
