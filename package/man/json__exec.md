#### Description

The `exec` command runs a command once for each record read from stdin. The input can be a JSON array or an NDJSON stream (one JSON object per line), so it consumes the output of any streaming aux4 command directly.

Each record is exposed to the command through aux4's `${...}` templating, the same convention used by the core `each:` executor:

- `${item}` — the whole record, as JSON
- `${item.field}` — a field of the record (nested paths like `${item.user.name}` work)
- `${index}` — the record's position, starting at `0`

The raw record is also written to the command's stdin, so complex payloads can be read with `aux4 json get` or `jq` instead of templating.

By default `exec` stops at the first command that fails. Pass `--ignoreErrors true` to keep going.

#### Usage

```bash
... | aux4 json exec <command>
```

command    Command to run per record; use ${item}, ${item.field} and ${index}
--ignoreErrors    Keep going when a command fails instead of stopping. Default: false

#### Example

```bash
echo '[{"file":"a.png"},{"file":"b.png"}]' | aux4 json exec 'echo resized ${item.file}'
```

```text
resized a.png
resized b.png
```

Consume an NDJSON stream and dispatch a job per record:

```bash
aux4 queue receive --name resize | aux4 json exec 'aux4 jobs run "echo resized ${item.file}"'
```

Read the whole record from stdin instead of templating:

```bash
cat products.ndjson | aux4 json exec 'curl -X POST https://api.example.com/import -d @-'
```
