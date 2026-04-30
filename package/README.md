# aux4/json

JSON CLI tools with streaming, pagination, and high performance. Extract values by path, pretty-print or minify JSON, index and group arrays by fields, merge multiple files, paginate large datasets with token-level streaming, collect NDJSON streams into arrays, and count items -- all from the command line with zero dependencies.

## Installation

```bash
aux4 aux4 pkger install aux4/json
```

## Quick Start

```bash
# Extract a value by path
echo '{"users":[{"name":"Alice"}]}' | aux4 json get '$.users.0.name'

# Pretty-print JSON
echo '{"name":"Alice","age":30}' | aux4 json pretty

# Minify JSON
cat data.json | aux4 json inline

# Count items in an array
cat users.json | aux4 json count

# Paginate a large array
cat users.json | aux4 json page --offset 0 --limit 10

# Merge multiple files by ID
aux4 json merge --id id users.json profiles.json

# Collect NDJSON into a JSON array
cat stream.ndjson | aux4 json collect
```

## Commands

### `aux4 json get`

Extract a value from JSON by path. Uses JSONPath syntax starting with `$`. Array elements are accessed by numeric index.

```bash
echo '...' | aux4 json get <path>
```

| Flag | Description | Default |
|------|-------------|---------|
| `path` | JSON path (e.g. `$.users.0.name`) | `$` |

#### Examples

```bash
# Get a nested field
echo '{"user":{"name":"Alice","email":"alice@example.com"}}' | aux4 json get '$.user.name'
# Output: Alice

# Get an array element
echo '{"items":["a","b","c"]}' | aux4 json get '$.items.1'
# Output: b

# Get the root object
cat config.json | aux4 json get '$'
```

### `aux4 json pretty`

Pretty-print JSON with indentation. Reads JSON from stdin and outputs formatted JSON with 2-space indentation.

```bash
echo '...' | aux4 json pretty
```

#### Examples

```bash
echo '{"name":"Alice","age":30}' | aux4 json pretty
# Output:
# {
#   "name": "Alice",
#   "age": 30
# }
```

### `aux4 json inline`

Minify JSON to a single line. Reads JSON from stdin and removes all unnecessary whitespace.

```bash
cat file.json | aux4 json inline
```

#### Examples

```bash
# Minify a pretty-printed file
cat pretty.json | aux4 json inline
# Output: {"name":"Alice","age":30}

# Pipe into another command
cat config.json | aux4 json inline | aux4 curl request --method POST --header "Content-Type: application/json" https://api.example.com/config
```

### `aux4 json index`

Index a JSON array by one or more ID fields. Transforms an array of objects into an object keyed by the specified field values. When multiple ID fields are provided, the key is the values joined by a dash.

```bash
cat file.json | aux4 json index --id <fields>
```

| Flag | Description | Default |
|------|-------------|---------|
| `--id` | ID field(s) separated by comma | required |

#### Examples

```bash
# Index by a single field
echo '[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]' | aux4 json index --id id
# Output: {"1":{"id":1,"name":"Alice"},"2":{"id":2,"name":"Bob"}}

# Index by composite key
echo '[{"region":"us","env":"prod","url":"us.prod.example.com"}]' | aux4 json index --id region,env
# Output: {"us-prod":{"region":"us","env":"prod","url":"us.prod.example.com"}}
```

### `aux4 json group`

Group a JSON array by one or more ID fields. Objects sharing the same ID values are combined: the ID fields appear once, and all other fields are collected into arrays.

```bash
cat file.json | aux4 json group --id <fields>
```

| Flag | Description | Default |
|------|-------------|---------|
| `--id` | ID field(s) separated by comma | required |

#### Examples

```bash
# Group by department
echo '[{"dept":"eng","name":"Alice"},{"dept":"eng","name":"Bob"},{"dept":"hr","name":"Carol"}]' | aux4 json group --id dept
# Output: [{"dept":"eng","name":["Alice","Bob"]},{"dept":"hr","name":["Carol"]}]
```

### `aux4 json collect`

Collect an NDJSON (newline-delimited JSON) stream into a single JSON array. Each line of input is parsed as a JSON value and added to the output array.

```bash
process-that-streams | aux4 json collect
```

#### Examples

```bash
# Collect NDJSON into an array
printf '{"id":1}\n{"id":2}\n{"id":3}\n' | aux4 json collect
# Output: [{"id":1},{"id":2},{"id":3}]

# Collect streaming output from another command
aux4 curl stream https://api.example.com/events | aux4 json collect > events.json
```

### `aux4 json merge`

Merge multiple JSON files by one or more ID fields. Objects from different files with matching ID values are combined into a single object. Fields from later files overwrite earlier ones (except the ID fields which remain unchanged).

```bash
aux4 json merge --id <fields> <file1> <file2> ...
```

| Flag | Description | Default |
|------|-------------|---------|
| `--id` | ID field(s) separated by comma | required |
| `files` | JSON files to merge (space-separated) | required |

#### Examples

```bash
# Merge user data with profile data
# users.json:  [{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]
# emails.json: [{"id":1,"email":"alice@example.com"},{"id":2,"email":"bob@example.com"}]
aux4 json merge --id id users.json emails.json
# Output: [{"id":1,"name":"Alice","email":"alice@example.com"},{"id":2,"name":"Bob","email":"bob@example.com"}]

# Merge with composite key
aux4 json merge --id region,env infra.json configs.json overrides.json
```

### `aux4 json page`

Paginate a JSON array. Returns a slice of the array starting at `--offset` with up to `--limit` items. When `--stream` is `true`, outputs each item as a separate line (NDJSON) instead of a JSON array, enabling token-level streaming for large datasets.

```bash
cat file.json | aux4 json page [--offset N] [--limit M] [--stream true]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--offset` | Number of items to skip | `0` |
| `--limit` | Number of items to return | `10` |
| `--stream` | Output as NDJSON stream | `false` |

#### Examples

```bash
# First page of 10 items
cat users.json | aux4 json page --offset 0 --limit 10

# Second page
cat users.json | aux4 json page --offset 10 --limit 10

# Stream mode for large datasets
cat large-dataset.json | aux4 json page --offset 0 --limit 100 --stream true

# Pipe streamed output for per-item processing
cat users.json | aux4 json page --offset 0 --limit 50 --stream true | while read -r line; do echo "$line" | aux4 json get '$.name'; done
```

### `aux4 json count`

Count the number of items in a JSON array. Reads a JSON array from stdin and outputs the count as a plain number.

```bash
cat file.json | aux4 json count
```

#### Examples

```bash
# Count items
echo '[{"id":1},{"id":2},{"id":3}]' | aux4 json count
# Output: 3

# Use in a script
total=$(cat users.json | aux4 json count)
echo "Total users: $total"
```

## Performance

All commands are implemented in Go for high performance and low memory usage. Key design decisions:

- **Streaming I/O** -- Commands like `collect`, `pretty`, `inline`, and `count` process input as a stream, keeping memory usage constant regardless of input size.
- **Token-level pagination** -- The `page` command with `--stream true` outputs items one at a time as NDJSON, avoiding the need to hold the entire result set in memory.
- **No dependencies** -- The binary is statically compiled with no external runtime requirements.
