# json describe

## structure of a document

### should report types, cardinality, optionality and example values

```execute
echo '{"exportedAt":"2026-07-27T09:00:00Z","orders":[{"id":"ord_01","status":"paid","total":149.9,"customer":{"id":"c1","email":"sally@aux4.io"},"note":"rush"},{"id":"ord_02","status":"pending","total":null,"customer":{"id":"c2","email":"pat@aux4.io"}},{"id":"ord_03","status":"paid","total":72.5,"customer":{"id":"c3","email":"lee@aux4.io"}}]}' | aux4 json describe
```

```expect
$                  object(2)
├─ exportedAt      string<date-time>      "2026-07-27T09:00:00Z"
└─ orders          array[3] of object(5)
   ├─ id           string                 "ord_01"
   ├─ status       string                 paid|pending
   ├─ total        number|null            149.9
   ├─ customer     object(2)
   │  ├─ id        string                 "c1"
   │  └─ email     string<email>          "sally@aux4.io"
   └─ note? (33%)  string                 "rush"
```

### should report a varying array length as a range

```execute
echo '[{"items":["a"]},{"items":["a","b","c"]}]' | aux4 json describe
```

```expect
$         array[2] of object(1)
└─ items  array[1..3] of string  a|b|c
```

### should put null last in a union so the real type leads

```execute
echo '[{"total":null},{"total":null},{"total":1}]' | aux4 json describe
```

```expect
$         array[3] of object(1)
└─ total  number|null            1
```

### should detect string formats

```execute
echo '[{"at":"2026-07-27T09:00:00Z","on":"2026-07-27","who":"sally@aux4.io","ref":"https://aux4.io/x"}]' | aux4 json describe
```

```expect
$       array[1] of object(4)
├─ at   string<date-time>      "2026-07-27T09:00:00Z"
├─ on   string<date>           "2026-07-27"
├─ who  string<email>          "sally@aux4.io"
└─ ref  string<uri>            "https://aux4.io/x"
```

### should not report a format when the values disagree

```execute
echo '[{"at":"2026-07-27T09:00:00Z"},{"at":"not a date"}]' | aux4 json describe
```

```expect
$      array[2] of object(1)
└─ at  string                 "2026-07-27T09:00:00Z"
```

## enums

### should list every value when a string field repeats a small vocabulary

```execute
echo '[{"s":"a"},{"s":"b"},{"s":"a"},{"s":"c"},{"s":"b"}]' | aux4 json describe
```

```expect
$     array[5] of object(1)
└─ s  string                 a|b|c
```

### should show an example instead when the values never repeat

```execute
echo '[{"s":"a"},{"s":"b"},{"s":"c"}]' | aux4 json describe
```

```expect
$     array[3] of object(1)
└─ s  string                 "a"
```

### should not report a numeric field as an enum

```execute
echo '[{"n":1},{"n":2},{"n":1},{"n":2}]' | aux4 json describe
```

```expect
$     array[4] of object(1)
└─ n  number                 1
```

## keyed objects

### should collapse an object whose many keys all hold the same shape

```execute
echo '{"users":{"u1":{"n":"a"},"u2":{"n":"b"},"u3":{"n":"c"},"u4":{"n":"d"},"u5":{"n":"e"},"u6":{"n":"f"},"u7":{"n":"g"},"u8":{"n":"h"},"u9":{"n":"i"},"u10":{"n":"j"},"u11":{"n":"k"},"u12":{"n":"l"},"u13":{"n":"m"},"u14":{"n":"n"},"u15":{"n":"o"},"u16":{"n":"p"},"u17":{"n":"q"},"u18":{"n":"r"},"u19":{"n":"s"},"u20":{"n":"t"},"u21":{"n":"u"},"u22":{"n":"v"},"u23":{"n":"w"},"u24":{"n":"x"},"u25":{"n":"y"},"u26":{"n":"z"}}}' | aux4 json describe
```

```expect
$           object(1)
└─ users    object<26 keys>
   └─ *     object(1)        keys: u1, u2, u3, ...
      └─ n  string           "a"
```

### should keep a wide record expanded rather than collapsing it

```execute
echo '[{"a":"1","b":"2","c":"3","d":"4","e":"5"}]' | aux4 json describe
```

```expect
$     array[1] of object(5)
├─ a  string                 "1"
├─ b  string                 "2"
├─ c  string                 "3"
├─ d  string                 "4"
└─ e  string                 "5"
```

## ndjson

### should merge a stream of records into one structure

```execute
printf '{"id":1,"kind":"a","meta":{"x":1}}\n{"id":2,"kind":"b"}\n{"id":3,"kind":"a","meta":{"x":2}}\n' | aux4 json describe
```

```expect
3 records
$               object(3)
├─ id           number     1
├─ kind         string     a|b
└─ meta? (66%)  object(1)
   └─ x         number     1
```

## path

### should describe only the requested sub-path

```execute
echo '{"orders":[{"id":"ord_01","customer":{"id":"c1","email":"sally@aux4.io"}},{"id":"ord_02","customer":{"id":"c2","email":"pat@aux4.io"}}]}' | aux4 json describe '$.orders[].customer'
```

```expect
$         object(2)
├─ id     string         "c1"
└─ email  string<email>  "sally@aux4.io"
```

### should describe a single element by index

```execute
echo '{"orders":[{"id":"ord_01"},{"id":"ord_02"}]}' | aux4 json describe '$.orders.0'
```

```expect
$      object(1)
└─ id  string     "ord_01"
```

### should fail when the path matches nothing

```execute
echo '{"a":1}' | aux4 json describe '$.nope'
```

```error
Error: path '$.nope' not found
```

## maxDepth

### should stop descending after the given number of levels

```execute
echo '{"a":{"b":{"c":{"d":1}}}}' | aux4 json describe --maxDepth 2
```

```expect
$        object(1)
└─ a     object(1)
   └─ b  object(...)
```

### should not spend a depth level on array elements

```execute
echo '{"orders":[{"id":"ord_01","customer":{"id":"c1"}}]}' | aux4 json describe --maxDepth 2
```

```expect
$               object(1)
└─ orders       array[1] of object(2)
   ├─ id        string                 "ord_01"
   └─ customer  object(...)
```

## sample

### should mark counts as lower bounds when the input was sampled

```execute
echo '{"rows":[{"n":1},{"n":2},{"n":3},{"n":4}]}' | aux4 json describe --sample 2
```

```expect
counts are lower bounds -- input was sampled
$        object(1)
└─ rows  array[2+, sampled] of object(1)
   └─ n  number                           1
```

## formats

### should emit one line per position with --format paths

```execute
echo '{"orders":[{"id":"ord_01","customer":{"email":"sally@aux4.io"}}]}' | aux4 json describe --format paths
```

```expect
$                          object(1)
$.orders                   array[1]
$.orders[]                 object(2)
$.orders[].id              string         "ord_01"
$.orders[].customer        object(1)
$.orders[].customer.email  string<email>  "sally@aux4.io"
```

### should emit a ready select structure with --format select

```execute
echo '{"orders":[{"id":"ord_01","status":"paid","customer":{"id":"c1","email":"sally@aux4.io"}}]}' | aux4 json describe '$.orders[]' --format select
```

```expect
id,status,customer[id,email]
```

### should emit the structure as data with --format json

```execute
echo '{"orders":[{"id":"ord_01"},{"id":"ord_02"}]}' | aux4 json describe '$.orders[]' --format json
```

```expect
{"count":2,"fields":{"id":{"count":2,"sample":"ord_01","type":"string"}},"required":["id"],"type":"object"}
```

### should emit draft 2020-12 with --format jsonschema

```execute
echo '[{"id":"a","status":"x","n":null},{"id":"b","status":"y","n":1},{"id":"c","status":"x"}]' | aux4 json describe '$[]' --format jsonschema
```

```expect
{"$schema":"https://json-schema.org/draft/2020-12/schema","properties":{"id":{"type":"string"},"n":{"type":["null","number"]},"status":{"enum":["x","y"],"type":"string"}},"required":["id","status"],"type":"object"}
```

### should fail on an unknown format

```execute
echo '{}' | aux4 json describe --format bogus
```

```error
Error: unknown format 'bogus' (expected tree, paths, json, jsonschema or select)
```

## color

### should emit plain text with --color never

```execute
echo '{"a":1,"b":"x"}' | aux4 json describe --color never
```

```expect
$     object(2)
├─ a  number     1
└─ b  string     "x"
```

### should colorize paths and types with --color always

```execute
echo '{"a":1,"b":"x"}' | aux4 json describe --format paths --color always | cat -v
```

```expect
^[[36m$^[[0m    ^[[32mobject(2)^[[0m
^[[36m$.a^[[0m  ^[[32mnumber^[[0m     ^[[90m1^[[0m
^[[36m$.b^[[0m  ^[[32mstring^[[0m     ^[[90m"x"^[[0m
```

### should not colorize the data formats

```execute
echo '{"a":1}' | aux4 json describe --format json --color always
```

```expect
{"count":1,"fields":{"a":{"count":1,"sample":1,"type":"number"}},"required":["a"],"type":"object"}
```

## values

### should leave example values out with --values false

```execute
echo '{"id":"c1","email":"sally@aux4.io","tags":["a","b","a"]}' | aux4 json describe --values false
```

```expect
$         object(3)
├─ id     string
├─ email  string<email>
└─ tags   array[3] of string
```
