# json collect

## collect NDJSON

### should collect lines into array

```execute
printf '{"id":1}\n{"id":2}\n{"id":3}\n' | aux4 json collect
```

```expect
[{"id":1},{"id":2},{"id":3}]
```

### should skip empty lines

```execute
printf '{"id":1}\n\n{"id":2}\n' | aux4 json collect
```

```expect
[{"id":1},{"id":2}]
```

### should handle single line

```execute
echo '{"id":1}' | aux4 json collect
```

```expect
[{"id":1}]
```
