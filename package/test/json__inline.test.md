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
