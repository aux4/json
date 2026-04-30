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
