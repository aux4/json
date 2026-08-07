#### Description

The `get` command extracts a value from JSON by path. It uses JSONPath syntax starting with `$`. Object fields are accessed with dot notation and array elements by numeric index. Negative indices count from the end, so `-1` is the last element, `-2` the second to last.

When the result is a string, it is output without quotes. When the result is an object or array, it is output as JSON.

Field order is preserved exactly as it appears in the source: only whitespace is rewritten, so the value you read back is the one the producer wrote.

#### Usage

```bash
echo '...' | aux4 json get <path>
```

path    JSON path expression (e.g. $.users.0.name). Default: $

#### Example

```bash
echo '{"user":{"name":"Alice","age":30}}' | aux4 json get '$.user.name'
```

```text
Alice
```

```bash
echo '{"items":["a","b","c"]}' | aux4 json get '$.items.1'
```

```text
b
```

```bash
echo '{"items":["a","b","c"]}' | aux4 json get '$.items.-1'
```

```text
c
```

```bash
echo '{"user":{"name":"Alice","age":30}}' | aux4 json get '$.user'
```

```text
{"name":"Alice","age":30}
```
