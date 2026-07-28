# json group

## group by single field

### should group and combine other fields into arrays

```execute
echo '[{"dept":"eng","name":"Alice"},{"dept":"eng","name":"Bob"},{"dept":"sales","name":"Charlie"}]' | aux4 json group --id dept
```

```expect:json
[
  {
    "dept": "eng",
    "name": [
      "Alice",
      "Bob"
    ]
  },
  {
    "dept": "sales",
    "name": [
      "Charlie"
    ]
  }
]
```

## group with deduplication

### should deduplicate values in arrays

```execute
echo '[{"dept":"eng","name":"Alice"},{"dept":"eng","name":"Alice"}]' | aux4 json group --id dept
```

```expect:json
[
  {
    "dept": "eng",
    "name": [
      "Alice"
    ]
  }
]
```

## field order

### should keep id fields first then non-id fields in input order

```execute
echo '[{"dept":"eng","zebra":1,"apple":2}]' | aux4 json group --id dept | aux4 json inline
```

```expect
[{"dept":"eng","zebra":[1],"apple":[2]}]
```
