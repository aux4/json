#### Description

The `group` command groups a JSON array by one or more ID fields. Objects that share the same ID values are combined: the ID fields appear once, and all other fields are collected into arrays.

When multiple ID fields are provided (comma-separated), objects must match on all specified fields to be grouped together.

Each group lists the id fields first, in the order they were requested, followed by the remaining fields in the order they first appeared in the input. Nothing is reordered alphabetically, and the output is identical between runs.

#### Usage

```bash
cat file.json | aux4 json group --id <fields>
```

--id    ID field(s) separated by comma (required)

#### Example

```bash
echo '[{"dept":"eng","name":"Alice"},{"dept":"eng","name":"Bob"},{"dept":"hr","name":"Carol"}]' | aux4 json group --id dept
```

```text
[{"dept":"eng","name":["Alice","Bob"]},{"dept":"hr","name":["Carol"]}]
```

```bash
echo '[{"region":"us","env":"prod","host":"h1"},{"region":"us","env":"prod","host":"h2"},{"region":"eu","env":"prod","host":"h3"}]' | aux4 json group --id region,env
```

```text
[{"region":"us","env":"prod","host":["h1","h2"]},{"region":"eu","env":"prod","host":["h3"]}]
```
