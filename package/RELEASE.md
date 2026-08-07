# 1.1.1

## json get

- Array indices can now be negative and count from the end: `-1` is the last
  element, `-2` the second to last. Works at any depth, e.g.
  `aux4 json get '$.users.-1.name'` or `aux4 json get '$.-1'`. Out-of-range
  negatives still report `array index N out of bounds`. Positive indices are
  unchanged. The same path navigator backs `get`, `describe`, `peek` and `exec`,
  so they all accept negative indices too.
