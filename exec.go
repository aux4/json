package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var itemTokenPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// runExec reads records from stdin (NDJSON, one object per line, or a single
// JSON array) and runs a command for each one. The current record is exposed to
// the command through aux4's ${item} / ${index} templating (nested access with
// ${item.field.sub}), matching the core "each:" executor. The raw record is
// also written to the command's stdin.
func runExec(args []string) {
	command := ""
	if len(args) > 0 {
		command = args[0]
	}
	ignoreErrors := len(args) > 1 && args[1] == "true"

	if strings.TrimSpace(command) == "" {
		fmt.Fprintln(os.Stderr, "command is required")
		os.Exit(1)
	}

	reader := bufio.NewReaderSize(os.Stdin, 1024*1024)
	index := 0

	// Peek the first non-space byte to tell a JSON array from NDJSON.
	if b, err := firstNonSpace(reader); err == nil && b == '[' {
		dec := json.NewDecoder(reader)
		dec.UseNumber()
		dec.Token() // consume opening '['
		for dec.More() {
			var item interface{}
			if err := dec.Decode(&item); err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid JSON element: %v\n", err)
				os.Exit(1)
			}
			raw, _ := json.Marshal(item)
			if err := runOne(command, index, item, raw); err != nil && !ignoreErrors {
				os.Exit(1)
			}
			index++
		}
		return
	}

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var item interface{}
		dec := json.NewDecoder(strings.NewReader(line))
		dec.UseNumber()
		if err := dec.Decode(&item); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid JSON: %s\n", line)
			if ignoreErrors {
				continue
			}
			os.Exit(1)
		}
		if err := runOne(command, index, item, []byte(line)); err != nil && !ignoreErrors {
			os.Exit(1)
		}
		index++
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}

func firstNonSpace(reader *bufio.Reader) (byte, error) {
	for {
		b, err := reader.Peek(1)
		if err != nil {
			return 0, err
		}
		if b[0] == ' ' || b[0] == '\t' || b[0] == '\n' || b[0] == '\r' {
			reader.ReadByte()
			continue
		}
		return b[0], nil
	}
}

func runOne(command string, index int, item interface{}, raw []byte) error {
	instruction := injectItem(command, index, item)
	cmd := exec.Command("sh", "-c", instruction)
	cmd.Stdin = strings.NewReader(string(raw))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: command failed for record %d: %v\n", index, err)
		return err
	}
	return nil
}

// injectItem replaces ${item}, ${index} and ${item.path} tokens in the template.
// Unknown tokens are left untouched so the shell handles them as it normally would.
func injectItem(template string, index int, item interface{}) string {
	return itemTokenPattern.ReplaceAllStringFunc(template, func(match string) string {
		token := strings.TrimSpace(itemTokenPattern.FindStringSubmatch(match)[1])
		switch {
		case token == "index":
			return strconv.Itoa(index)
		case token == "item":
			return valueToString(item)
		case strings.HasPrefix(token, "item."):
			return valueToString(resolvePath(item, strings.Split(token[len("item."):], ".")))
		default:
			return match
		}
	})
}

func resolvePath(value interface{}, path []string) interface{} {
	current := value
	for _, key := range path {
		switch typed := current.(type) {
		case map[string]interface{}:
			current = typed[key]
		case []interface{}:
			i, err := strconv.Atoi(key)
			if err != nil || i < 0 || i >= len(typed) {
				return nil
			}
			current = typed[i]
		default:
			return nil
		}
	}
	return current
}

func valueToString(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case json.Number:
		return typed.String()
	case bool:
		return strconv.FormatBool(typed)
	default:
		bytes, err := json.Marshal(typed)
		if err != nil {
			return ""
		}
		return string(bytes)
	}
}
