#### Description

The `pretty` command reads JSON from stdin and outputs it with 2-space indentation. Useful for making compact JSON human-readable.

Only whitespace is rewritten. Field order is preserved exactly as it appears in the input, so what you read back is the document you started with — the order the producer chose, not an alphabetical one.

Input may be a single JSON document or an NDJSON stream; each value in a stream is indented in turn.

#### Usage

```bash
echo '...' | aux4 json pretty
```

#### Example

```bash
echo '{"name":"Alice","age":30,"address":{"city":"NYC"}}' | aux4 json pretty
```

```text
{
  "name": "Alice",
  "age": 30,
  "address": {
    "city": "NYC"
  }
}
```
