# json filter

## equality

### should keep records equal to a string value

```execute
echo '[{"name":"Alice"},{"name":"Bob"}]' | aux4 json filter --field name --value Bob | aux4 json inline
```

```expect
[{"name":"Bob"}]
```

### should compare numbers numerically for eq

```execute
echo '[{"age":9},{"age":10},{"age":100}]' | aux4 json filter --field age --value 10 | aux4 json inline
```

```expect
[{"age":10}]
```

### should keep records not equal with ne

```execute
echo '[{"name":"Alice"},{"name":"Bob"}]' | aux4 json filter --field name --op ne --value Bob | aux4 json inline
```

```expect
[{"name":"Alice"}]
```

## numeric comparison

### should keep records greater than a value

```execute
echo '[{"age":30},{"age":25},{"age":40}]' | aux4 json filter --field age --op gt --value 26 | aux4 json inline
```

```expect
[{"age":30},{"age":40}]
```

### should keep records less than or equal

```execute
echo '[{"age":30},{"age":25},{"age":40}]' | aux4 json filter --field age --op lte --value 30 | aux4 json inline
```

```expect
[{"age":30},{"age":25}]
```

## nested field

### should filter on a dot-path

```execute
echo '[{"addr":{"city":"NY"}},{"addr":{"city":"LA"}}]' | aux4 json filter --field addr.city --value LA | aux4 json inline
```

```expect
[{"addr":{"city":"LA"}}]
```

## contains

### should match a substring in a string field

```execute
echo '[{"t":"hello world"},{"t":"goodbye"}]' | aux4 json filter --field t --op contains --value wor | aux4 json inline
```

```expect
[{"t":"hello world"}]
```

### should match membership in an array field

```execute
echo '[{"tags":["a","b"]},{"tags":["c"]}]' | aux4 json filter --field tags --op contains --value b | aux4 json inline
```

```expect
[{"tags":["a","b"]}]
```

## missing field

### should drop records missing the field

```execute
echo '[{"n":1},{"x":2},{"n":3}]' | aux4 json filter --field n --op gte --value 1 | aux4 json inline
```

```expect
[{"n":1},{"n":3}]
```

## preserve field order

### should keep each record's field order

```execute
echo '[{"zebra":1,"apple":2}]' | aux4 json filter --field zebra --value 1 | aux4 json inline
```

```expect
[{"zebra":1,"apple":2}]
```

## input stream

### should filter NDJSON from stdin

```execute
printf '{"n":3}\n{"n":1}\n{"n":2}\n' | aux4 json filter --field n --op lt --value 3 --inputStream true | aux4 json inline
```

```expect
[{"n":1},{"n":2}]
```

## invalid op

### should reject an unknown operator

```execute
echo '[]' | aux4 json filter --field n --op between --value 5
```

```error:partial
invalid op 'between'
```
