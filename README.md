# @aux4/json
JSON cli tools

![npm](https://img.shields.io/npm/v/@aux4/json) 

## Install

```bash
npm install -g @aux4/json
```

## Usage

### Merge
Merge multiple json files into one.

```bash
$ aux4-json merge file1.json file2.json ... > merged.json
```

### Group
Group json file by an id (one or multiple fields). Combine the other fields into an array.

```bash
$ cat file.json | aux4-json group --id field1,field2 > grouped.json
```

### Index
Indexes json file by an id (one or multiple fields).

```bash
$ cat file.json | aux4-json index --id field1,field2 > indexed.json
```

### Collect
Collect multiple json from stream into one array.

```bash
$ process that streams json | aux4-json collect > collected.json
```

### Get
Get value from json file by path. Don't forget to escape the `$`.

```bash
$ cat file.json | aux4-json get \$.path.to.value
```

### Inline
Inline json file into a single line.

```bash
$ cat file.json | aux4-json inline > inlined.json
```

### Pretty
Pretty print json file.

```bash
$ cat file.json | aux4-json pretty > pretty.json
```