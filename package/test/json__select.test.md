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
[{"id":1,"address":{"city":"Austin","country":"US"}}]
```

### should rename fields with source:output

```execute
echo '[{"buyerName":"Sally","total":42}]' | aux4 json select 'buyerName:customer,total:amount'
```

```expect
[{"customer":"Sally","amount":42}]
```

### should project each object in a nested array

```execute
echo '[{"id":1,"items":[{"sku":"A","qty":2,"junk":1},{"sku":"B","qty":5,"junk":9}]}]' | aux4 json select 'id,items[sku,qty]'
```

```expect
[{"id":1,"items":[{"sku":"A","qty":2},{"sku":"B","qty":5}]}]
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

## field order

### should emit fields in the order the structure asked for

```execute
echo '[{"zebra":1,"apple":2,"mango":3}]' | aux4 json select 'mango,zebra,apple'
```

```expect
[{"mango":3,"zebra":1,"apple":2}]
```

### should keep the order of nested values that are passed through whole

```execute
echo '[{"id":1,"c":{"zebra":1,"apple":2}}]' | aux4 json select 'c,id'
```

```expect
[{"c":{"zebra":1,"apple":2},"id":1}]
```
