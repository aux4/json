# json peek

## show the first values at a path

### should print the first record indented

```execute
echo '{"orders":[{"id":"ord_01","status":"paid"},{"id":"ord_02","status":"pending"},{"id":"ord_03","status":"paid"}]}' | aux4 json peek '$.orders[]' --limit 1
```

```expect
{
  "id": "ord_01",
  "status": "paid"
}
```

### should stop at the limit

```execute
echo '{"orders":[{"id":"ord_01","total":149.9},{"id":"ord_02","total":10},{"id":"ord_03","total":72.5}]}' | aux4 json peek '$.orders[]' --limit 2 --format inline
```

```expect
{"id":"ord_01","total":149.9}
{"id":"ord_02","total":10}
```

### should default to three values

```execute
echo '[{"a":1},{"a":2},{"a":3},{"a":4},{"a":5}]' | aux4 json peek --format inline
```

```expect
{"a":1}
{"a":2}
{"a":3}
```

### should show every value with --limit 0

```execute
echo '[1,2,3,4,5]' | aux4 json peek --limit 0 --format inline
```

```expect
1
2
3
4
5
```

### should preserve the original key order

```execute
echo '{"b":2,"a":1,"c":3}' | aux4 json peek --format inline
```

```expect
{"b":2,"a":1,"c":3}
```

## path

### should peek into a nested path across records

```execute
echo '{"orders":[{"c":{"id":"c1"}},{"c":{"id":"c2"}}]}' | aux4 json peek '$.orders[].c' --format inline
```

```expect
{"id":"c1"}
{"id":"c2"}
```

### should peek into an array named without brackets

```execute
echo '{"rows":[{"a":1},{"a":2},{"a":3}]}' | aux4 json peek '$.rows' --limit 2 --format inline
```

```expect
{"a":1}
{"a":2}
```

### should peek a single element by index

```execute
echo '{"rows":[{"a":1},{"a":2},{"a":3}]}' | aux4 json peek '$.rows.1' --format inline
```

```expect
{"a":2}
```

### should print an object whole when the path is not an array

```execute
echo '{"meta":{"v":1,"name":"x"}}' | aux4 json peek '$.meta' --format inline
```

```expect
{"v":1,"name":"x"}
```

### should print a scalar at the path

```execute
echo '{"n":42}' | aux4 json peek '$.n' --format inline
```

```expect
42
```

### should fail when the path matches nothing

```execute
echo '{"a":1}' | aux4 json peek '$.nope'
```

```error
Error: path '$.nope' not found
```

## ndjson

### should peek at a stream of records

```execute
printf '{"i":1}\n{"i":2}\n{"i":3}\n' | aux4 json peek --limit 2 --format inline
```

```expect
{"i":1}
{"i":2}
```

## errors

### should fail on an unknown format

```execute
echo '{}' | aux4 json peek --format bogus
```

```error
Error: unknown format 'bogus' (expected pretty or inline)
```

### should fail on a negative limit

```execute
echo '{}' | aux4 json peek --limit -1
```

```error
Error: limit must be a non-negative number
```

## chaining

### should feed inline output into select as NDJSON

```execute
echo '{"orders":[{"id":1,"name":"Chai","junk":9},{"id":2,"name":"Chang","junk":9}]}' | aux4 json peek '$.orders[]' --format inline | aux4 json select 'id,name' --inputStream true
```

```expect
{"id":1,"name":"Chai"}
{"id":2,"name":"Chang"}
```
