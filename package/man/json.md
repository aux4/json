#### Description

The `json` command group provides JSON CLI tools with streaming, pagination, and high performance. It contains the following subcommands:

- **get** -- Extract a value from JSON by path
- **pretty** -- Pretty-print JSON with indentation
- **inline** -- Minify JSON to a single line
- **index** -- Index a JSON array by ID field(s)
- **group** -- Group a JSON array by ID field(s)
- **collect** -- Collect an NDJSON stream into a JSON array
- **merge** -- Merge multiple JSON files by ID field(s)
- **page** -- Paginate a JSON array
- **count** -- Count items in a JSON array

#### Usage

```bash
aux4 json <command>
```

#### Example

```bash
echo '{"name":"Alice"}' | aux4 json pretty
```

```text
{
  "name": "Alice"
}
```
