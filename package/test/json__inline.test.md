# json inline

## minify JSON

### should output single line

```execute
printf '{\n  "id": 1,\n  "name": "Alice"\n}' | aux4 json inline
```

```expect
{"id":1,"name":"Alice"}
```

### should minify an array

```execute
printf '[\n  {"id": 1},\n  {"id": 2}\n]' | aux4 json inline
```

```expect
[{"id":1},{"id":2}]
```

### should preserve the original key order

```execute
echo '{"zebra":1,"apple":2,"mango":3}' | aux4 json inline
```

```expect
{"zebra":1,"apple":2,"mango":3}
```

### should preserve key order at every level of nesting

```execute
echo '{"z":{"y":1,"x":2},"a":[{"q":1,"b":2}]}' | aux4 json inline
```

```expect
{"z":{"y":1,"x":2},"a":[{"q":1,"b":2}]}
```

### should minify each value of an NDJSON stream

```execute
printf '{\n  "b": 1\n}\n{\n  "a": 2\n}\n' | aux4 json inline
```

```expect
{"b":1}
{"a":2}
```

### should fail on empty input

```execute
printf '' | aux4 json inline
```

```error
Error: no JSON input
```
