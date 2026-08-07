#### Description

The `sort` command sorts a JSON array and prints the sorted array. Records are ordered by one or more keys given with `--by`, each a dot-path into the record (e.g. `age` or `address.city`). Repeat `--by` for a multi-key sort: earlier keys take precedence and later keys break ties.

- **Type-aware comparison** — when both values at a key are numbers they compare numerically, so `9` sorts before `100`. Otherwise they compare as text.
- **Multi-key** — `--by a --by b` sorts by `a`, then by `b` within equal `a`.
- **Scalar arrays** — with no `--by`, whole elements are compared, so an array of numbers or strings sorts by value.
- **Order** — `--order asc` (default) or `--order desc`.
- **Missing keys** — records that do not have a sort key sort after those that do.
- **Stable** — records that compare equal keep their input order.
- **Field order preserved** — records are moved, never rewritten, so each object keeps its own field order.

By default `sort` reads a single JSON array from stdin. With `--inputStream true` it reads NDJSON (one JSON value per line) and sorts those records.

#### Usage

```bash
... | aux4 json sort [--by <field>]... [--order <asc|desc>] [--inputStream <true|false>]
```

--by            Field to sort by (dot-path). Repeat for a multi-key sort. Omit to sort scalar arrays by value.
--order         Sort order: `asc` (default) or `desc`.
--inputStream   Read NDJSON from stdin instead of a JSON array. Default: false

#### Example

```bash
echo '[{"name":"Bob","age":30},{"name":"Alice","age":25}]' | aux4 json sort --by age
```

```json
[{"name":"Alice","age":25},{"name":"Bob","age":30}]
```

Multi-key, descending:

```bash
echo '[{"team":"a","score":2},{"team":"a","score":9},{"team":"b","score":5}]' \
  | aux4 json sort --by team --by score --order desc
```

```json
[{"team":"b","score":5},{"team":"a","score":9},{"team":"a","score":2}]
```

Streaming (NDJSON in):

```bash
aux4 mdb stream --file inventory.mdb --table Products \
  | aux4 json sort --by UnitPrice --inputStream true
```
