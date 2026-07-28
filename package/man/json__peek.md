#### Description

The `peek` command prints the first few values at a path and stops reading. It is the companion to `describe`: `describe` tells you the shape of a document, `peek` shows you what the data actually looks like.

It streams and exits as soon as it has printed enough, so peeking at the head of a multi-gigabyte file costs nothing — the tail is never read. Everything off the path is skipped without being materialised.

When the value at the path is an array, `peek` prints its first `--limit` elements one after another. Anything else — an object, a string, a number — is printed whole as a single value. Key order is preserved exactly as it appears in the source.

Input may be a single JSON document or an NDJSON stream.

#### Usage

```bash
... | aux4 json peek [path]
```

path       Peek at this path, e.g. `$.orders[]` where `[]` means every element. Default: `$`
--limit    How many values to show; 0 for all. Default: 3
--format   `pretty` (indented) or `inline` (NDJSON, one value per line). Default: pretty

`--format inline` emits one compact value per line, which is valid NDJSON and so feeds straight into `aux4 json select --inputStream true` or `aux4 json collect`.

Point `peek` at an array path. Pointing it at an object prints that whole object, so on a document whose root is one large object, peek the array inside it rather than the root — run `describe` first if you do not know the shape yet.

#### Example

```bash
echo '{"orders":[{"id":"ord_01","status":"paid"},{"id":"ord_02","status":"pending"}]}' \
  | aux4 json peek '$.orders[]' --limit 1
```

```json
{
  "id": "ord_01",
  "status": "paid"
}
```

One line per record, ready to pipe onward:

```bash
cat orders.json | aux4 json peek '$.orders[]' --limit 2 --format inline
```

```json
{"id":"ord_01","status":"paid","total":149.9}
{"id":"ord_02","status":"pending","total":10}
```

Peek into a nested path — every matching value, not just the first record's:

```bash
cat orders.json | aux4 json peek '$.orders[].customer' --limit 2 --format inline
```

```json
{"id":"c1","email":"sally@aux4.io"}
{"id":"c2","email":"pat@aux4.io"}
```

A single element by index:

```bash
cat orders.json | aux4 json peek '$.orders.1'
```

The pair that makes a large unknown file workable — shape first, then real rows:

```bash
aux4 json describe --maxDepth 2 < export.json
aux4 json peek '$.rows[]' --limit 3 < export.json
```
