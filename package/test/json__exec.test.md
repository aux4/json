# json exec

## run a command per record

### should inject a field from each record of a JSON array

```execute
echo '[{"file":"a.png"},{"file":"b.png"}]' | aux4 json exec 'echo resized ${item.file}'
```

```expect
resized a.png
resized b.png
```

### should expose the record index

```execute
echo '[{"x":7},{"x":9}]' | aux4 json exec 'echo ${index}:${item.x}'
```

```expect
0:7
1:9
```

### should consume an NDJSON stream

```execute
printf '{"file":"c.png"}\n{"file":"d.png"}\n' | aux4 json exec 'echo got ${item.file}'
```

```expect
got c.png
got d.png
```

### should resolve nested paths

```execute
echo '[{"payload":{"file":"deep.png"}}]' | aux4 json exec 'echo ${item.payload.file}'
```

```expect
deep.png
```

### should keep going with ignoreErrors

```execute
echo '[{"n":1},{"n":2},{"n":3}]' | aux4 json exec --ignoreErrors true 'echo n=${item.n}'
```

```expect
n=1
n=2
n=3
```
