#### Description

The `merge` command merges multiple JSON files by one or more ID fields. Each file must contain a JSON array of objects. Objects from different files with matching ID values are combined into a single object. Fields from later files overwrite earlier ones (except the ID fields which remain unchanged).

When multiple ID fields are provided (comma-separated), objects must match on all specified fields to be merged.

#### Usage

```bash
aux4 json merge --id <fields> <file1> <file2> ...
```

--id    ID field(s) separated by comma (required)
files   JSON files to merge, space-separated (required)

#### Example

```bash
# users.json contains: [{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]
# emails.json contains: [{"id":1,"email":"alice@example.com"},{"id":2,"email":"bob@example.com"}]
aux4 json merge --id id users.json emails.json
```

```text
[{"id":1,"name":"Alice","email":"alice@example.com"},{"id":2,"name":"Bob","email":"bob@example.com"}]
```
