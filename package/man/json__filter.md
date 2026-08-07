#### Description

The `filter` command keeps the records of a JSON array whose `--field` satisfies a predicate, and prints the array of matches. The field is a dot-path into each record (e.g. `age` or `address.city`).

- **Operators** (`--op`): `eq` (default), `ne`, `gt`, `lt`, `gte`, `lte`, `contains`.
- **Type-aware comparison** — when the field value and `--value` are both numbers they compare numerically; otherwise as text. So `--field age --op gt --value 26` compares numbers, while `--field name --value Bob` compares strings.
- **contains** — substring match for a string field, membership match for an array field. Other value types never match `contains`.
- **Missing fields** — a record that does not have the field is dropped for every operator.
- **Field order preserved** — matching records are carried through untouched, keeping their own field order.

By default `filter` reads a single JSON array from stdin. With `--inputStream true` it reads NDJSON (one JSON value per line) and filters those records.

#### Usage

```bash
... | aux4 json filter <field> [--op <operator>] [--value <value>] [--inputStream <true|false>]
```

field           Dot-path tested on each record (e.g. `age`, `address.city`).
--op            Comparison operator: `eq`, `ne`, `gt`, `lt`, `gte`, `lte`, `contains`. Default: eq
--value         Value to compare the field against.
--inputStream   Read NDJSON from stdin instead of a JSON array. Default: false

#### Example

```bash
echo '[{"name":"Alice","age":30},{"name":"Bob","age":25}]' | aux4 json filter age --op gt --value 26
```

```json
[{"name":"Alice","age":30}]
```

String equality on a nested field:

```bash
echo '[{"addr":{"city":"NY"}},{"addr":{"city":"LA"}}]' | aux4 json filter addr.city --value LA
```

```json
[{"addr":{"city":"LA"}}]
```

Array membership with `contains`:

```bash
echo '[{"tags":["a","b"]},{"tags":["c"]}]' | aux4 json filter tags --op contains --value b
```

```json
[{"tags":["a","b"]}]
```

Streaming (NDJSON in):

```bash
aux4 mdb stream --file inventory.mdb --table Products \
  | aux4 json filter UnitsInStock --op lte --value 5 --inputStream true
```
