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

```expect:partial
**"email": "alice@test.com"**"name": "Alice"**"email": "bob@test.com"**"name": "Bob"**
```
