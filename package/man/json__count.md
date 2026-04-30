#### Description

The `count` command reads a JSON array from stdin and outputs the number of items as a plain integer. Useful for scripting and validation.

#### Usage

```bash
cat file.json | aux4 json count
```

#### Example

```bash
echo '[{"id":1},{"id":2},{"id":3}]' | aux4 json count
```

```text
3
```

```bash
echo '[]' | aux4 json count
```

```text
0
```
