package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// errPeekDone unwinds the walk once enough values have been printed, so a peek
// at the head of a huge file never reads the tail.
var errPeekDone = errors.New("peek done")

type peeker struct {
	limit  int
	inline bool
	found  int
	out    *bufio.Writer
}

func (p *peeker) done() bool { return p.limit > 0 && p.found >= p.limit }

func (p *peeker) emit(raw []byte) error {
	if p.done() {
		return errPeekDone
	}
	p.found++

	var buf bytes.Buffer
	if p.inline {
		if err := json.Compact(&buf, raw); err != nil {
			return err
		}
	} else if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(p.out, buf.String()); err != nil {
		return err
	}
	if p.done() {
		return errPeekDone
	}
	return nil
}

// writeValue rebuilds the JSON value whose opening token has already been read.
// json.Decoder cannot un-read a token, so once we have looked at a value to see
// whether it is an array we have to reconstruct it from the token stream.
func writeValue(dec *json.Decoder, t json.Token, buf *bytes.Buffer) error {
	switch v := t.(type) {
	case json.Delim:
		switch v {
		case '{':
			buf.WriteByte('{')
			first := true
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return err
				}
				key, ok := kt.(string)
				if !ok {
					return fmt.Errorf("expected object key, got %v", kt)
				}
				if !first {
					buf.WriteByte(',')
				}
				first = false
				encoded, err := json.Marshal(key)
				if err != nil {
					return err
				}
				buf.Write(encoded)
				buf.WriteByte(':')

				vt, err := dec.Token()
				if err != nil {
					return err
				}
				if err := writeValue(dec, vt, buf); err != nil {
					return err
				}
			}
			if _, err := dec.Token(); err != nil {
				return err
			}
			buf.WriteByte('}')
			return nil
		case '[':
			buf.WriteByte('[')
			first := true
			for dec.More() {
				if !first {
					buf.WriteByte(',')
				}
				first = false
				vt, err := dec.Token()
				if err != nil {
					return err
				}
				if err := writeValue(dec, vt, buf); err != nil {
					return err
				}
			}
			if _, err := dec.Token(); err != nil {
				return err
			}
			buf.WriteByte(']')
			return nil
		}
		return fmt.Errorf("unexpected delimiter %v", v)
	case string:
		encoded, err := json.Marshal(v)
		if err != nil {
			return err
		}
		buf.Write(encoded)
	case json.Number:
		buf.WriteString(v.String())
	case bool:
		buf.WriteString(strconv.FormatBool(v))
	case nil:
		buf.WriteString("null")
	}
	return nil
}

// emitTarget prints the value at the resolved path. An array is streamed element
// by element so that peeking into a million-record array never materialises it;
// anything else is rebuilt and printed whole.
func (p *peeker) emitTarget(dec *json.Decoder) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}

	if d, ok := t.(json.Delim); ok && d == '[' {
		for dec.More() {
			if p.done() {
				return errPeekDone
			}
			var raw json.RawMessage
			if err := dec.Decode(&raw); err != nil {
				return err
			}
			if err := p.emit(raw); err != nil {
				return err
			}
		}
		_, err := dec.Token()
		return err
	}

	var buf bytes.Buffer
	if err := writeValue(dec, t, &buf); err != nil {
		return err
	}
	return p.emit(buf.Bytes())
}

// seek walks to every value matching the path, printing as it goes and skipping
// everything else without materialising it.
func (p *peeker) seek(dec *json.Decoder, segs []pathSeg) error {
	if len(segs) == 0 {
		return p.emitTarget(dec)
	}

	seg := segs[0]
	t, err := dec.Token()
	if err != nil {
		return err
	}
	d, ok := t.(json.Delim)
	if !ok {
		return nil
	}

	if seg.name != "" {
		if d != '{' {
			return skipContainer(dec)
		}
		for dec.More() {
			if p.done() {
				return errPeekDone
			}
			kt, err := dec.Token()
			if err != nil {
				return err
			}
			if kt == seg.name {
				if err := p.seek(dec, segs[1:]); err != nil {
					return err
				}
			} else if err := skipValue(dec); err != nil {
				return err
			}
		}
	} else {
		if d != '[' {
			return skipContainer(dec)
		}
		i := 0
		for dec.More() {
			if p.done() {
				return errPeekDone
			}
			if seg.any || i == seg.index {
				if err := p.seek(dec, segs[1:]); err != nil {
					return err
				}
			} else if err := skipValue(dec); err != nil {
				return err
			}
			i++
		}
	}

	_, err = dec.Token()
	return err
}

func runPeek(args []string) {
	path := "$"
	limit := 3
	format := "pretty"

	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		path = strings.TrimSpace(args[0])
	}
	if len(args) > 1 && strings.TrimSpace(args[1]) != "" {
		v, err := strconv.Atoi(strings.TrimSpace(args[1]))
		if err != nil || v < 0 {
			fmt.Fprintln(os.Stderr, "Error: limit must be a non-negative number")
			os.Exit(1)
		}
		limit = v
	}
	if len(args) > 2 && strings.TrimSpace(args[2]) != "" {
		format = strings.TrimSpace(args[2])
	}

	switch format {
	case "pretty", "inline":
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown format '%s' (expected pretty or inline)\n", format)
		os.Exit(1)
	}

	segs, err := parseDescribePath(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	p := &peeker{limit: limit, inline: format == "inline", out: out}

	dec := json.NewDecoder(bufio.NewReaderSize(os.Stdin, 1024*1024))
	dec.UseNumber()

	records := 0
	for dec.More() {
		err := p.seek(dec, segs)
		records++
		if errors.Is(err, errPeekDone) {
			break
		}
		if err != nil {
			out.Flush()
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		if p.done() {
			break
		}
	}

	if records == 0 {
		fmt.Fprintln(os.Stderr, "Error: no JSON input")
		os.Exit(1)
	}
	if p.found == 0 {
		out.Flush()
		fmt.Fprintf(os.Stderr, "Error: path '%s' not found\n", path)
		os.Exit(1)
	}
}
