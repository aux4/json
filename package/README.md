# aux4/json

JSON CLI tools with streaming, pagination, and high performance. Extract values by path, pretty-print or minify JSON, index and group arrays by fields, merge multiple files, paginate large datasets with token-level streaming, collect NDJSON streams into arrays, count items, describe the structure of a file too large to read, and peek at the records inside it -- all from the command line with zero dependencies.

## Field order

Every command preserves the order of an object's fields. Nothing is ever sorted alphabetically on its way through — the document you read back is the one the producer wrote, which matters both for scanning output against a source and for feeding a renderer whose columns follow the field order.

Where you specify an order, that order is used: `aux4 json select 'name,id'` puts `name` first no matter where it sat in the input.

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

# Learn the structure of a file too large to read
cat orders.json | aux4 json describe

# Look at the first few records at a path
cat orders.json | aux4 json peek '$.orders[]'
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

### `aux4 json exec`

Run a command for each record read from stdin. The input can be a JSON array or an NDJSON stream (one object per line), so `exec` consumes the output of any streaming command directly.

Each record is exposed to the command through aux4's `${...}` templating — the same convention as the core `each:` executor:

- `${item}` — the whole record, as JSON
- `${item.field}` — a field of the record (nested paths like `${item.user.name}` work)
- `${index}` — the record's position, starting at `0`

The raw record is also written to the command's stdin. By default `exec` stops at the first failing command; pass `--ignoreErrors true` to keep going.

```bash
... | aux4 json exec <command>
```

#### Examples

```bash
# Run a command per record
echo '[{"file":"a.png"},{"file":"b.png"}]' | aux4 json exec 'echo resized ${item.file}'
# resized a.png
# resized b.png

# Consume a stream and dispatch a background job per record
aux4 queue receive --name resize | aux4 json exec 'aux4 jobs run "echo resized ${item.file}"'

# Read the whole record from stdin instead of templating
cat products.ndjson | aux4 json exec 'curl -X POST https://api.example.com/import -d @-'
```

### `aux4 json select`

Project each record down to a chosen set of fields, keeping the result as JSON with types preserved. It is the JSON counterpart of `aux4 2table` and `aux4 render` — the same field notation that renders a table or a list here produces projected JSON:

- `name,age,city` — keep these fields
- `address[city,country]` — keep a nested object, selecting only its sub-fields
- `source:output` — rename a field (source first, output second)
- `user.email` — dot paths reach into nested objects
- `total{format:currency}` — a trailing `{...}` display block is accepted but ignored, so specs are interchangeable with 2table/render

Missing fields become `null`. With `--inputStream true`, `select` reads NDJSON and emits one projected object per line.

```bash
... | aux4 json select <structure>
```

#### Examples

```bash
# Keep a subset of fields (types preserved)
echo '[{"id":1,"name":"Chai","price":18,"note":"drop"}]' | aux4 json select 'id,name,price'
# [{"id":1,"name":"Chai","price":18}]

# Nested selection and renaming
echo '[{"buyerName":"Sally","address":{"city":"Austin","country":"US","zip":"78701"}}]' \
  | aux4 json select 'buyerName:customer,address[city,country]'
# [{"address":{"city":"Austin","country":"US"},"customer":"Sally"}]

# Streaming projection (NDJSON in, NDJSON out)
aux4 mdb stream --file inventory.mdb --table Products \
  | aux4 json select 'ProductName:name,UnitPrice:price' --inputStream true
```

### `aux4 json describe`

Report the structure — the schema — of a JSON document without printing the data. It streams the input, so a multi-gigabyte file is summarised in one pass with memory proportional to the structure rather than the file.

This is the command to reach for when a file is too large to read: describe it first, learn the paths, then use `get`, `select`, `page` or `count` on the parts that matter.

For every position it reports the type (or a union such as `number|null`), cardinality (`array[482913]`, `array[1..37]`, `object(9)`), optionality (`note? (25%)` for a field only some records have), enums when a string or boolean field draws on a small vocabulary, an example value, and a detected string format (`date-time`, `date`, `uuid`, `email`, `uri`).

Input may be a single JSON document or an NDJSON stream, which is merged into one structure. Objects with many keys that all hold the same shape — a map keyed by id — collapse to `object<N keys>` with a single `*` value shape; a wide record whose fields merely share a type is not collapsed.

```bash
... | aux4 json describe [path]
```

- `path` — describe only this sub-path, e.g. `$.orders[]` where `[]` means every element. Default `$`
- `--format` — `tree` (default), `paths`, `json`, `jsonschema` or `select`
- `--maxDepth` — stop descending after this many levels; `0` for unlimited. Arrays do not consume a level
- `--sample` — read at most this many array elements or records; `0` for all. Sampled counts print as `array[100+, sampled]`
- `--values` — include example values and enums. Default `true`
- `--color` — `auto`, `always` or `never`. Default `auto`

The `tree` and `paths` formats are colorized — cyan for paths and field names, green for types, yellow for the optional marker, magenta for value lists, grey for glyphs and examples. The `json`, `jsonschema` and `select` formats are never colorized, since they exist to be piped. Use `--color never` or `NO_COLOR=1` when feeding the output to another program.

#### Examples

```bash
# The shape of a document, with counts and example values
aux4 json describe < orders.json
# $                  object(3)
# ├─ exportedAt      string<date-time>         "2026-07-27T09:00:00Z"
# ├─ source          string                    "hub"
# └─ orders          array[4] of object(6)
#    ├─ id           string                    "ord_01"
#    ├─ status       string                    paid|pending|shipped
#    ├─ total        number|null               149.9
#    ├─ customer     object(3)
#    │  ├─ id        string                    "c1"
#    │  ├─ email     string<email>             "sally@aux4.io"
#    │  └─ vip       boolean                   false|true
#    ├─ items        array[1..3] of object(2)
#    │  ├─ sku       string                    "A"
#    │  └─ qty       number                    2
#    └─ note? (25%)  string                    "rush"

# A flat, greppable path per position
aux4 json describe --format paths < orders.json
# $.orders                   array[4]
# $.orders[].status          string             paid|pending|shipped
# $.orders[].customer.email  string<email>      "sally@aux4.io"

# Drill into one branch only
aux4 json describe '$.orders[].customer' < orders.json

# Generate the argument for `select` instead of writing field paths by hand
aux4 json describe '$.orders[]' --format select < orders.json
# id,status,total,customer[id,email,vip],items[sku,qty],note

# Describe an NDJSON stream
aux4 mdb stream --file inventory.mdb --table Products | aux4 json describe

# Sample a huge array and leave example values out
aux4 json describe '$.rows[]' --sample 1000 --values false < export.json
```

`[]` stands for any element. To read one with `get`, replace it with a concrete index — `$.orders[].id` becomes `$.orders.0.id`:

```bash
cat orders.json \
  | aux4 json get '$.orders' \
  | aux4 json page --offset 0 --limit 50 \
  | aux4 json select 'id,status,total,customer[id,email]'
```

`--format json` returns the structure as data for further processing, and `--format jsonschema` returns JSON Schema draft 2020-12 for validators and code generators.

### `aux4 json peek`

Print the first few values at a path and stop reading. This is the companion to `describe`: `describe` tells you the shape of a document, `peek` shows you what the data actually looks like.

It streams and exits as soon as it has printed enough, so peeking at the head of a multi-gigabyte file costs nothing — the tail is never read, and everything off the path is skipped without being materialised. Key order is preserved exactly as it appears in the source.

When the value at the path is an array, `peek` prints its first `--limit` elements. Anything else — an object, a string, a number — is printed whole as a single value.

```bash
... | aux4 json peek [path]
```

- `path` — peek at this path, e.g. `$.orders[]` where `[]` means every element. Default `$`
- `--limit` — how many values to show; `0` for all. Default `3`
- `--format` — `pretty` (indented) or `inline` (NDJSON, one value per line). Default `pretty`

Point `peek` at an array path. Pointing it at an object prints that whole object, so on a document whose root is one large object, peek the array inside it rather than the root — run `describe` first if you do not know the shape yet.

#### Examples

```bash
# The first record, indented
echo '{"orders":[{"id":"ord_01","status":"paid"},{"id":"ord_02","status":"pending"}]}' \
  | aux4 json peek '$.orders[]' --limit 1
# {
#   "id": "ord_01",
#   "status": "paid"
# }

# One line per record
cat orders.json | aux4 json peek '$.orders[]' --limit 2 --format inline
# {"id":"ord_01","status":"paid","total":149.9}
# {"id":"ord_02","status":"pending","total":10}

# Every matching value at a nested path, not just the first record's
cat orders.json | aux4 json peek '$.orders[].customer' --limit 2 --format inline
# {"id":"c1","email":"sally@aux4.io"}
# {"id":"c2","email":"pat@aux4.io"}

# A single element by index
cat orders.json | aux4 json peek '$.orders.1'
```

`--format inline` emits valid NDJSON, so it feeds straight into the streaming commands:

```bash
cat orders.json | aux4 json peek '$.orders[]' --format inline \
  | aux4 json select 'id,name' --inputStream true
```

The pair that makes a large unknown file workable — shape first, then real rows:

```bash
aux4 json describe --maxDepth 2 < export.json
aux4 json peek '$.rows[]' --limit 3 < export.json
```

## License

Apache-2.0
