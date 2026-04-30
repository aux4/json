#### Description

The `page` command paginates a JSON array. It returns a slice of the array starting at `--offset` with up to `--limit` items. The output is a JSON array by default.

When `--stream` is set to `true`, each item is output on its own line as NDJSON instead of wrapping in an array. This enables token-level streaming for large datasets, keeping memory usage low.

#### Usage

```bash
cat file.json | aux4 json page [--offset N] [--limit M] [--stream true]
```

--offset    Number of items to skip (default: 0)
--limit     Number of items to return (default: 10)
--stream    Output as NDJSON stream (default: false)

#### Example

```bash
echo '[{"id":1},{"id":2},{"id":3},{"id":4},{"id":5}]' | aux4 json page --offset 0 --limit 3
```

```text
[{"id":1},{"id":2},{"id":3}]
```

```bash
echo '[{"id":1},{"id":2},{"id":3},{"id":4},{"id":5}]' | aux4 json page --offset 2 --limit 2 --stream true
```

```text
{"id":3}
{"id":4}
```
