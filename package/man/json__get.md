#### Description

The `get` command extracts a value from JSON by path. It uses JSONPath syntax starting with `$`. Object fields are accessed with dot notation and array elements by numeric index.

When the result is a string, it is output without quotes. When the result is an object or array, it is output as JSON.

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
echo '{"user":{"name":"Alice","age":30}}' | aux4 json get '$.user'
```

```text
{"name":"Alice","age":30}
```
