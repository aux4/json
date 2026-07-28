# json index

## index by single field

### should index items by id

```execute
echo '[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]' | aux4 json index --id id
```

```expect:json
{
  "1": {
    "id": 1,
    "name": "Alice"
  },
  "2": {
    "id": 2,
    "name": "Bob"
  }
}
```

## index with duplicate keys

### should create array for duplicates

```execute
echo '[{"id":1,"name":"Alice"},{"id":1,"name":"Bob"}]' | aux4 json index --id id
```

```expect:partial
*"1": [*
```

## index by multiple fields

### should use composite key

```execute
echo '[{"dept":"eng","role":"dev","name":"Alice"},{"dept":"eng","role":"qa","name":"Bob"}]' | aux4 json index --id dept,role
```

```expect:partial
**"name": "Alice"**"name": "Bob"**
```

## field order

### should never reorder record fields

```execute
echo '[{"id":1,"zebra":1,"apple":2}]' | aux4 json index --id id | aux4 json inline
```

```expect
{"1":{"id":1,"zebra":1,"apple":2}}
```
