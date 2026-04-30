# json count

## count items

### should count items in array

```execute
echo '[{"id":1},{"id":2},{"id":3}]' | aux4 json count
```

```expect
3
```

### should count empty array

```execute
echo '[]' | aux4 json count
```

```expect
0
```

### should count single item

```execute
echo '[{"id":1}]' | aux4 json count
```

```expect
1
```
