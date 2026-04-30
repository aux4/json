#### Description

The `pretty` command reads JSON from stdin and outputs it with 2-space indentation. Useful for making compact JSON human-readable.

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
