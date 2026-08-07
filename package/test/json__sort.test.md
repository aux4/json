# json sort

## sort by a single field

### should sort strings ascending by default

```execute
echo '[{"name":"Bob"},{"name":"Alice"},{"name":"Carol"}]' | aux4 json sort --by name | aux4 json inline
```

```expect
[{"name":"Alice"},{"name":"Bob"},{"name":"Carol"}]
```

### should sort descending

```execute
echo '[{"name":"Bob"},{"name":"Alice"},{"name":"Carol"}]' | aux4 json sort --by name --order desc | aux4 json inline
```

```expect
[{"name":"Carol"},{"name":"Bob"},{"name":"Alice"}]
```

### should sort numbers numerically not lexically

```execute
echo '[{"age":9},{"age":100},{"age":25}]' | aux4 json sort --by age | aux4 json inline
```

```expect
[{"age":9},{"age":25},{"age":100}]
```

## sort by a nested field

### should sort by a dot-path

```execute
echo '[{"addr":{"city":"NY"}},{"addr":{"city":"LA"}}]' | aux4 json sort --by addr.city | aux4 json inline
```

```expect
[{"addr":{"city":"LA"}},{"addr":{"city":"NY"}}]
```

## multi-key sort

### should break ties with a second key

```execute
echo '[{"a":1,"b":2},{"a":1,"b":1},{"a":0,"b":9}]' | aux4 json sort --by a --by b | aux4 json inline
```

```expect
[{"a":0,"b":9},{"a":1,"b":1},{"a":1,"b":2}]
```

## sort scalar arrays

### should sort a scalar array with no --by

```execute
echo '[3,1,2,10]' | aux4 json sort | aux4 json inline
```

```expect
[1,2,3,10]
```

## preserve field order

### should keep each record's field order

```execute
echo '[{"zebra":2,"apple":1},{"zebra":1,"apple":2}]' | aux4 json sort --by zebra | aux4 json inline
```

```expect
[{"zebra":1,"apple":2},{"zebra":2,"apple":1}]
```

## input stream

### should sort NDJSON from stdin

```execute
printf '{"n":3}\n{"n":1}\n{"n":2}\n' | aux4 json sort --by n --inputStream true | aux4 json inline
```

```expect
[{"n":1},{"n":2},{"n":3}]
```

## missing keys

### should sort records missing the key last

```execute
echo '[{"n":2},{"x":9},{"n":1}]' | aux4 json sort --by n | aux4 json inline
```

```expect
[{"n":1},{"n":2},{"x":9}]
```

## invalid order

### should reject an unknown order

```execute
echo '[]' | aux4 json sort --by n --order sideways
```

```error:partial
invalid order 'sideways'
```
