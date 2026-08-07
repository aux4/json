# 1.1.2

## json get

- Paths can now contain a `*` **wildcard** that projects across an array.
  `$.users.*` returns the array elements themselves; `$.users.*.name` maps every
  element through the rest of the path into a flat array (`["Alice","Bob"]`).
  Elements that do not resolve the remaining path are skipped, and wildcards
  compose (`$.teams.*.members.*.name`). A `*` on a non-array reports
  `expected array at '*'`. Numeric and negative indices are unchanged.

## json sort (new)

- Sort a JSON array by one or more dot-path keys: `aux4 json sort --by age`,
  repeatable for a multi-key sort. `--order asc|desc` (default `asc`). Numbers
  compare numerically, everything else as text; with no `--by` whole elements are
  compared, which sorts arrays of scalars. Records missing a key sort last, the
  sort is stable, and record field order is preserved. `--inputStream true` reads
  NDJSON from stdin.

## json filter (new)

- Keep the records of a JSON array matching a predicate:
  `aux4 json filter age --op gt --value 26`. Operators: `eq` (default), `ne`,
  `gt`, `lt`, `gte`, `lte`, `contains`. Numeric comparison when both sides are
  numbers, else string; `contains` is substring for strings and membership for
  arrays. Records missing the field are dropped, and kept records preserve their
  field order. `--inputStream true` reads NDJSON from stdin.
