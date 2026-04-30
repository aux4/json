#### Description

The `index` command transforms a JSON array into an object keyed by one or more ID fields. Each array element becomes a value in the resulting object, with its key derived from the specified fields.

When multiple ID fields are provided (comma-separated), the key is the field values joined by a dash.

#### Usage

```bash
cat file.json | aux4 json index --id <fields>
```

--id    ID field(s) separated by comma (required)

#### Example

```bash
echo '[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]' | aux4 json index --id id
```

```text
{"1":{"id":1,"name":"Alice"},"2":{"id":2,"name":"Bob"}}
```

```bash
echo '[{"region":"us","env":"prod","url":"us.example.com"},{"region":"eu","env":"prod","url":"eu.example.com"}]' | aux4 json index --id region,env
```

```text
{"us-prod":{"region":"us","env":"prod","url":"us.example.com"},"eu-prod":{"region":"eu","env":"prod","url":"eu.example.com"}}
```
