# json page

## paginate array

### should return first page

```execute
echo '[{"id":1},{"id":2},{"id":3},{"id":4},{"id":5}]' | aux4 json page --offset 0 --limit 2
```

```expect
[{"id":1},{"id":2}]
```

### should return second page

```execute
echo '[{"id":1},{"id":2},{"id":3},{"id":4},{"id":5}]' | aux4 json page --offset 2 --limit 2
```

```expect
[{"id":3},{"id":4}]
```

### should return last page

```execute
echo '[{"id":1},{"id":2},{"id":3},{"id":4},{"id":5}]' | aux4 json page --offset 4 --limit 10
```

```expect
[{"id":5}]
```

### should return empty array when offset exceeds length

```execute
echo '[{"id":1},{"id":2}]' | aux4 json page --offset 10 --limit 5
```

```expect
[]
```

## stream mode

### should output NDJSON

```execute
echo '[{"id":1},{"id":2},{"id":3}]' | aux4 json page --offset 0 --limit 2 --stream true
```

```expect
{"id":1}
{"id":2}
```
