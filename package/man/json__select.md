#### Description

The `select` command projects each record down to a chosen set of fields, keeping the result as JSON with types preserved. It is the JSON counterpart of `aux4 2table` and `aux4 render`: the same field notation that renders a table or a list here produces projected JSON.

The structure string uses the shared 2table/render notation:

- `name,age,city` — keep these fields
- `address[city,country]` — keep a nested object, selecting only its sub-fields
- `source:output` — rename a field (source first, output name second)
- `user.email` — dot paths reach into nested objects
- `total{format:currency}` — a trailing `{...}` display block is accepted but ignored (so specs are interchangeable with 2table/render)

Fields come out **in the order the structure lists them** -- a projection chooses fields *and their order*, so `select 'name,id'` puts `name` first regardless of where it sat in the source. Values carried across whole keep their own field order too.

Missing fields become `null`. A nested selection over an array of objects projects each element.

By default `select` reads a JSON array (or a single object) and emits projected JSON. With `--inputStream true` it reads NDJSON (one object per line) and emits one projected object per line, so it fits in a stream.

#### Usage

```bash
... | aux4 json select <structure>
```

structure       Fields to keep; e.g. `id,name,address[city,country]`
--inputStream   Read NDJSON from stdin and emit one projected object per line. Default: false

#### Example

```bash
echo '[{"id":1,"name":"Chai","price":18,"note":"drop"}]' | aux4 json select 'id,name,price'
```

```json
[{"id":1,"name":"Chai","price":18}]
```

Nested selection and renaming:

```bash
echo '[{"buyerName":"Sally","address":{"city":"Austin","country":"US","zip":"78701"}}]' \
  | aux4 json select 'buyerName:customer,address[city,country]'
```

```json
[{"address":{"city":"Austin","country":"US"},"customer":"Sally"}]
```

Streaming projection (NDJSON in, NDJSON out):

```bash
aux4 mdb stream --file inventory.mdb --table Products \
  | aux4 json select 'ProductName:name,UnitPrice:price' --inputStream true
```
