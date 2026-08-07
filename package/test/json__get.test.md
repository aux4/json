# json get

## extract nested path

### should get a nested value

```execute
echo '{"users":[{"name":"Alice"},{"name":"Bob"}]}' | aux4 json get '$.users.0.name'
```

```expect
"Alice"
```

## negative array index

### should get the last element with -1

```execute
echo '[{"q":"a"},{"q":"b"},{"q":"c"}]' | aux4 json get '$.-1.q'
```

```expect
"c"
```

### should count back from the end with -2

```execute
echo '[{"q":"a"},{"q":"b"},{"q":"c"}]' | aux4 json get '$.-2.q'
```

```expect
"b"
```

### should work on a nested array

```execute
echo '{"users":[{"name":"Alice"},{"name":"Bob"}]}' | aux4 json get '$.users.-1.name'
```

```expect
"Bob"
```

### should reach -1 at the root

```execute
echo '[10,20,30]' | aux4 json get '$.-1'
```

```expect
30
```

### should report a negative index that is out of bounds

```execute
echo '[1,2,3]' | aux4 json get '$.-9'
```

```error:partial
array index -9 out of bounds (length 3)
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

## field order

### should never reorder object fields

```execute
echo '{"o":{"zebra":1,"apple":2,"mango":3}}' | aux4 json get '$.o' | aux4 json inline
```

```expect
{"zebra":1,"apple":2,"mango":3}
```

## wildcard projection

### should project a field across an array

```execute
echo '{"users":[{"name":"Alice"},{"name":"Bob"}]}' | aux4 json get '$.users.*.name' | aux4 json inline
```

```expect
["Alice","Bob"]
```

### should return the elements themselves for a trailing wildcard

```execute
echo '{"users":[{"name":"Alice"},{"name":"Bob"}]}' | aux4 json get '$.users.*' | aux4 json inline
```

```expect
[{"name":"Alice"},{"name":"Bob"}]
```

### should project across the root array

```execute
echo '[{"id":1},{"id":2},{"id":3}]' | aux4 json get '$.*.id' | aux4 json inline
```

```expect
[1,2,3]
```

### should skip elements missing the projected field

```execute
echo '{"u":[{"name":"A"},{"x":1},{"name":"B"}]}' | aux4 json get '$.u.*.name' | aux4 json inline
```

```expect
["A","B"]
```

### should compose a wildcard with a nested field

```execute
echo '{"users":[{"addr":{"city":"NY"}},{"addr":{"city":"LA"}}]}' | aux4 json get '$.users.*.addr.city' | aux4 json inline
```

```expect
["NY","LA"]
```

### should support multiple wildcards

```execute
echo '{"teams":[{"members":[{"name":"A"}]},{"members":[{"name":"B"},{"name":"C"}]}]}' | aux4 json get '$.teams.*.members.*.name' | aux4 json inline
```

```expect
[["A"],["B","C"]]
```

### should error when the wildcard target is not an array

```execute
echo '{"users":{"name":"Alice"}}' | aux4 json get '$.users.*'
```

```error:partial
expected array at '*'
```
