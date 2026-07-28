#### Description

The `describe` command reports the structure (the schema) of a JSON document without printing the data. It streams the input, so a multi-gigabyte file is summarised in a single pass with memory proportional to the *structure*, not to the file.

It is built for the case where you need to work with a JSON file that is far too large to read: run `describe` first, learn the paths, then use `get`, `select`, `page` or `count` on the parts that matter.

For every position in the document it reports:

- the **type**, or a union of types when a field varies (`number|null`)
- **cardinality** — `array[482913]` for a fixed length, `array[1..37]` when the length varies, `object(9)` for the field count
- **optionality** — a field missing from some records is marked `name?` with the percentage of records that have it
- **enums** — when a string or boolean field draws on a small vocabulary, every value is listed
- an **example value** for each leaf
- a detected **format** for strings (`date-time`, `date`, `uuid`, `email`, `uri`) when every observed value agrees

Input may be a single JSON document or an NDJSON stream (one document per line); an NDJSON stream is merged into one structure and the record count is printed above it.

Objects with many keys that all hold the same shape — a map keyed by id, `{"usr_1": {...}, "usr_2": {...}}` — collapse to `object<N keys>` with a single `*` value shape instead of listing every key. A wide record whose fields merely happen to share a type is not collapsed.

#### Usage

```bash
... | aux4 json describe [path]
```

path         Describe only this sub-path, e.g. `$.orders[]` where `[]` means every element. Default: `$`
--format     `tree`, `paths`, `json`, `jsonschema` or `select`. Default: tree
--maxDepth   Stop descending after this many levels; 0 for unlimited. Default: 0
--sample     Read at most this many array elements or records; 0 for all. Default: 0
--values     Include example values and enums in the output. Default: true
--color      `auto`, `always` or `never`. Default: auto

Arrays do not consume a depth level: an array and its element type render on one line, so `--maxDepth 2` shows the fields of `$.orders[]`.

When `--sample` cuts the input short, the counts are lower bounds and the output says so — a sampled length prints as `array[100+, sampled]`.

#### Color

The `tree` and `paths` formats are colorized: cyan for paths and field names (the parts you copy), green for types, yellow for the optional marker, magenta for value lists and grey for tree glyphs and example values. The data formats — `json`, `jsonschema` and `select` — are never colorized, since they exist to be piped.

A terminal check on the command's own output would be useless here, because aux4 runs every command with its output through a pipe: the command never sees a terminal even when aux4 does. aux4 sets `CLICOLOR_FORCE` for exactly this reason, so `--color auto` follows that signal and the `NO_COLOR` convention rather than probing for a TTY. Use `--color never` (or `NO_COLOR=1`) to force plain output when feeding another program, and `--color always` to keep color through a pipe.

#### Formats

`tree` is the default and the most readable:

```bash
aux4 json describe < orders.json
```

```
$                  object(3)
├─ exportedAt      string<date-time>         "2026-07-27T09:00:00Z"
├─ source          string                    "hub"
└─ orders          array[4] of object(6)
   ├─ id           string                    "ord_01"
   ├─ status       string                    paid|pending|shipped
   ├─ total        number|null               149.9
   ├─ customer     object(3)
   │  ├─ id        string                    "c1"
   │  ├─ email     string<email>             "sally@aux4.io"
   │  └─ vip       boolean                   false|true
   ├─ items        array[1..3] of object(2)
   │  ├─ sku       string                    "A"
   │  └─ qty       number                    2
   └─ note? (25%)  string                    "rush"
```

`paths` emits one line per position, which greps well and gives paths you can pass straight to `get`:

```bash
aux4 json describe --format paths < orders.json
```

```
$.orders                   array[4]
$.orders[]                 object(6)
$.orders[].id              string             "ord_01"
$.orders[].status          string             paid|pending|shipped
$.orders[].customer.email  string<email>      "sally@aux4.io"
```

`[]` stands for any element. To read one with `get`, replace it with a concrete index — `$.orders[].id` becomes `$.orders.0.id`.

`select` emits the shared 2table/render field notation, so the structure of a document becomes the argument for `aux4 json select`:

```bash
aux4 json describe '$.orders[]' --format select < orders.json
```

```
id,status,total,customer[id,email,vip],items[sku,qty],note
```

`json` returns the structure as data for further processing, and `jsonschema` returns JSON Schema draft 2020-12 for validators and code generators.

#### Example

Working through a large file without ever loading it — describe it, then act on what you found:

```bash
aux4 json describe --maxDepth 2 < orders.json
aux4 json describe '$.orders[].customer' < orders.json

aux4 json describe '$.orders[]' --format select < orders.json
# id,status,total,customer[id,email,vip],items[sku,qty],note

cat orders.json \
  | aux4 json get '$.orders' \
  | aux4 json page --offset 0 --limit 50 \
  | aux4 json select 'id,status,total,customer[id,email]'
```

Describing an NDJSON stream:

```bash
aux4 mdb stream --file inventory.mdb --table Products | aux4 json describe
```

Sampling the first 1000 elements of a huge array, and leaving example values out:

```bash
aux4 json describe '$.rows[]' --sample 1000 --values false < export.json
```
