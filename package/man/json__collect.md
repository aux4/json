#### Description

The `collect` command reads an NDJSON (newline-delimited JSON) stream from stdin and combines all lines into a single JSON array. Each line of input is parsed as a JSON value and appended to the output array.

This is useful for converting streaming output from other commands into a standard JSON array.

#### Usage

```bash
process-that-streams | aux4 json collect
```

#### Example

```bash
printf '{"id":1,"name":"Alice"}\n{"id":2,"name":"Bob"}\n{"id":3,"name":"Carol"}\n' | aux4 json collect
```

```text
[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"},{"id":3,"name":"Carol"}]
```
