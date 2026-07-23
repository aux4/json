# json select

## project records to selected fields

### should keep only the selected fields and preserve types

```execute
echo '[{"id":1,"name":"Chai","price":18,"note":"drop"},{"id":2,"name":"Chang","price":19,"note":"drop"}]' | aux4 json select 'id,name,price'
```

```expect
[{"id":1,"name":"Chai","price":18},{"id":2,"name":"Chang","price":19}]
```

### should select nested sub-fields with field[a,b]

```execute
echo '[{"id":1,"address":{"city":"Austin","country":"US","zip":"78701"}}]' | aux4 json select 'id,address[city,country]'
```

```expect
[{"address":{"city":"Austin","country":"US"},"id":1}]
```

### should rename fields with source:output

```execute
echo '[{"buyerName":"Sally","total":42}]' | aux4 json select 'buyerName:customer,total:amount'
```

```expect
[{"amount":42,"customer":"Sally"}]
```

### should project each object in a nested array

```execute
echo '[{"id":1,"items":[{"sku":"A","qty":2,"junk":1},{"sku":"B","qty":5,"junk":9}]}]' | aux4 json select 'id,items[sku,qty]'
```

```expect
[{"id":1,"items":[{"qty":2,"sku":"A"},{"qty":5,"sku":"B"}]}]
```

### should return null for missing fields

```execute
echo '[{"id":1}]' | aux4 json select 'id,missing'
```

```expect
[{"id":1,"missing":null}]
```

### should ignore a trailing display format block

```execute
echo '[{"name":"Chai","price":18}]' | aux4 json select 'name,price{format:currency}'
```

```expect
[{"name":"Chai","price":18}]
```

## stream mode

### should project NDJSON in and NDJSON out with --inputStream

```execute
printf '{"id":1,"name":"Chai","x":9}\n{"id":2,"name":"Chang","x":9}\n' | aux4 json select 'id,name' --inputStream true
```

```expect
{"id":1,"name":"Chai"}
{"id":2,"name":"Chang"}
```
