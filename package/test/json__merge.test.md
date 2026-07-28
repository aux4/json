# json merge

```beforeAll
echo '[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]' > merge_users.json
echo '[{"id":1,"email":"alice@test.com"},{"id":2,"email":"bob@test.com"}]' > merge_emails.json
```

```afterAll
rm -f merge_users.json merge_emails.json
```

## merge two files

### should merge by id

```execute
aux4 json merge --id id --files "merge_users.json merge_emails.json"
```

```expect
[
  {
    "id": 1,
    "name": "Alice",
    "email": "alice@test.com"
  },
  {
    "id": 2,
    "name": "Bob",
    "email": "bob@test.com"
  }
]
```

### should keep the base record order and append fields the later file adds

Fields follow the first file's order, with anything a later file introduces
appended in the order that file listed it -- never alphabetically.

```execute
echo '[{"zebra":1,"id":1,"apple":2}]' > merge_order_a.json
echo '[{"id":1,"yak":3,"beta":4}]' > merge_order_b.json
aux4 json merge --id id --files "merge_order_a.json merge_order_b.json" | aux4 json inline
rm -f merge_order_a.json merge_order_b.json
```

```expect
[{"zebra":1,"id":1,"apple":2,"yak":3,"beta":4}]
```
