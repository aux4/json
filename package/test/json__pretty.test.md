# json pretty

## format compact JSON

### should pretty-print with indentation

```execute
echo '{"id":1,"name":"Alice"}' | aux4 json pretty
```

```expect:json
{
  "id": 1,
  "name": "Alice"
}
```

### should pretty-print an array

```execute
echo '[{"id":1},{"id":2}]' | aux4 json pretty
```

```expect:partial
[
  {
    "id": 1
**
```

### should preserve the original key order

```execute
echo '{"zebra":1,"apple":2,"mango":3}' | aux4 json pretty
```

```expect
{
  "zebra": 1,
  "apple": 2,
  "mango": 3
}
```

### should format each value of an NDJSON stream

```execute
printf '{"b":1}\n{"a":2}\n' | aux4 json pretty
```

```expect
{
  "b": 1
}
{
  "a": 2
}
```

### should fail on empty input

```execute
printf '' | aux4 json pretty
```

```error
Error: no JSON input
```
