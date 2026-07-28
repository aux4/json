#### Description

The `inline` command reads JSON from stdin and removes all unnecessary whitespace, outputting compact single-line JSON. Useful for preparing JSON for storage, transmission, or piping into other commands.

Only whitespace is removed. Field order is preserved exactly as it appears in the input, so minifying and re-expanding a document round-trips without reordering it.

Input may be a single JSON document or an NDJSON stream; each value in a stream is minified onto its own line.

#### Usage

```bash
cat file.json | aux4 json inline
```

#### Example

```bash
printf '{\n  "name": "Alice",\n  "age": 30\n}' | aux4 json inline
```

```text
{"name":"Alice","age":30}
```
