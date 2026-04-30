# json get

## extract nested path

### should get a nested value

```execute
echo '{"users":[{"name":"Alice"},{"name":"Bob"}]}' | aux4 json get '$.users.0.name'
```

```expect
"Alice"
```

### should get an array

```execute
echo '{"users":[{"name":"Alice"},{"name":"Bob"}]}' | aux4 json get $.users
```

```expect:partial
[
  {
    "name": "Alice"
**
```

### should get root with dollar sign

```execute
echo '{"id":1}' | aux4 json get $
```

```expect:json
{
  "id": 1
}
```

## extract from object

### should get a top-level field

```execute
echo '{"name":"Alice","age":30}' | aux4 json get $.name
```

```expect
"Alice"
```
