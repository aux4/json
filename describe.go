package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	// enumLimit is how many distinct values a leaf may hold and still be
	// reported as an enum. Past this, the field is not an enum and no partial
	// list is emitted -- a truncated list would read as complete and mislead.
	enumLimit = 30
	// mapCollapseKeys is the key count past which an object whose values all
	// share one container shape is treated as a map (keyed by id) rather than a
	// record, so a 50k-key object collapses to a single value shape.
	mapCollapseKeys = 25
	// scalarCollapseKeys is the same threshold for objects whose values are all
	// scalars. It is much higher because a wide record (40 string columns) looks
	// exactly like a scalar map and must not be collapsed.
	scalarCollapseKeys = 100
	sampleMaxLen       = 48
	enumInlineMaxLen   = 72
)

// Colours follow the palette aux4 uses for help output: cyan for the things you
// copy (paths and field names), green for types, yellow for the optional
// marker, magenta for value lists and grey for everything incidental.
const (
	colorGlyph  = "90"
	colorName   = "36"
	colorMark   = "33"
	colorType   = "32"
	colorSample = "90"
	colorEnum   = "35"
	colorHeader = "90"
)

// useColor is resolved once from --color and the environment. It stays false
// for the data formats, which are meant to be piped.
var useColor bool

// resolveColor decides whether to emit escape codes. A TTY check on our own
// stdout is useless here: aux4 runs every command with its output through a
// pipe, so the child never sees a terminal even when aux4 does. aux4 sets
// CLICOLOR_FORCE for exactly this reason, so "auto" follows that signal and the
// NO_COLOR convention instead.
func resolveColor(mode string) bool {
	switch mode {
	case "always":
		return true
	case "never":
		return false
	}
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	return os.Getenv("CLICOLOR_FORCE") != ""
}

func paint(s, code string) string {
	if !useColor || s == "" {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

// descNode is one position in the inferred structure. It is observed once per
// occurrence of that position in the input, so Count on a field divided by
// Count on its parent gives how often the field is present.
type descNode struct {
	Count int
	Types map[string]int

	// object
	Fields   map[string]*descNode
	Order    []string
	MapLike  bool
	MapValue *descNode
	MapKeys  int
	KeyNames []string

	// array
	ArrMin    int
	ArrMax    int
	Elem      *descNode
	Truncated bool

	// scalar. Sample holds the raw value; quoting happens at render time so the
	// data formats can emit it with its real type.
	Values         map[string]int
	Overflow       bool
	Sample         string
	SampleIsString bool
	HasSample      bool
	Format         string
	FormatSet      bool

	// set when --maxDepth stopped the walk before reading this container
	Deeper bool
}

func newDescNode() *descNode {
	return &descNode{Types: map[string]int{}, ArrMin: -1}
}

func (n *descNode) addType(t string) { n.Types[t]++ }

// has reports whether this position was ever observed as the given type.
func (n *descNode) has(t string) bool { return n.Types[t] > 0 }

// typeNames returns the observed types ordered by frequency, ties broken
// alphabetically so output is deterministic.
func (n *descNode) typeNames() []string {
	names := make([]string, 0, len(n.Types))
	for t := range n.Types {
		names = append(names, t)
	}
	// "null" always sorts last so a nullable field reads as "number|null"
	// rather than leading with the absence of a value.
	sort.Slice(names, func(i, j int) bool {
		if (names[i] == "null") != (names[j] == "null") {
			return names[j] == "null"
		}
		if n.Types[names[i]] != n.Types[names[j]] {
			return n.Types[names[i]] > n.Types[names[j]]
		}
		return names[i] < names[j]
	})
	return names
}

func (n *descNode) observeLength(length int, truncated bool) {
	if n.ArrMin < 0 || length < n.ArrMin {
		n.ArrMin = length
	}
	if length > n.ArrMax {
		n.ArrMax = length
	}
	if truncated {
		n.Truncated = true
	}
}

func (n *descNode) observeScalar(key string, isString bool) {
	if !n.HasSample {
		n.Sample = key
		n.SampleIsString = isString
		n.HasSample = true
	}
	if n.Overflow {
		return
	}
	if n.Values == nil {
		n.Values = map[string]int{}
	}
	if _, seen := n.Values[key]; !seen && len(n.Values) >= enumLimit {
		n.Overflow = true
		n.Values = nil
		return
	}
	n.Values[key]++
}

var (
	dateTimePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[Tt ]\d{2}:\d{2}:\d{2}`)
	datePattern     = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	uuidPattern     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	emailPattern    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	uriPattern      = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://\S+$`)
)

func detectFormat(s string) string {
	switch {
	case dateTimePattern.MatchString(s):
		return "date-time"
	case datePattern.MatchString(s):
		return "date"
	case uuidPattern.MatchString(s):
		return "uuid"
	case emailPattern.MatchString(s):
		return "email"
	case uriPattern.MatchString(s):
		return "uri"
	}
	return ""
}

// observeFormat keeps a format only while every observed string agrees on it.
func (n *descNode) observeFormat(s string) {
	f := detectFormat(s)
	if !n.FormatSet {
		n.Format = f
		n.FormatSet = true
		return
	}
	if n.Format != f {
		n.Format = ""
	}
}

// signature is a shallow shape fingerprint used to decide whether an object's
// values are uniform enough to collapse into a map.
func (n *descNode) signature() string {
	parts := n.typeNames()
	if n.has("object") && !n.MapLike {
		keys := append([]string{}, n.Order...)
		sort.Strings(keys)
		return "object:" + strings.Join(keys, ",")
	}
	if n.has("array") {
		if n.Elem != nil {
			return "array:" + n.Elem.signature()
		}
		return "array:"
	}
	return strings.Join(parts, "|")
}

func (n *descNode) isContainer() bool { return n.has("object") || n.has("array") }

// collapseToMap merges every field seen so far into a single value shape. Only
// called once the key count and shape uniformity checks have passed.
func (n *descNode) collapseToMap() {
	merged := newDescNode()
	for _, key := range n.Order {
		mergeDescNode(merged, n.Fields[key])
	}
	n.KeyNames = append([]string{}, n.Order...)
	if len(n.KeyNames) > 3 {
		n.KeyNames = n.KeyNames[:3]
	}
	n.MapLike = true
	n.MapValue = merged
	n.Fields = nil
	n.Order = nil
}

// shouldCollapse decides whether an object with many keys is really a map.
func (n *descNode) shouldCollapse() bool {
	if len(n.Order) < mapCollapseKeys {
		return false
	}
	first := n.Fields[n.Order[0]]
	sig := first.signature()
	for _, key := range n.Order[1:] {
		if n.Fields[key].signature() != sig {
			return false
		}
	}
	if first.isContainer() {
		return true
	}
	return len(n.Order) >= scalarCollapseKeys
}

// mergeDescNode folds src into dst, combining two observations of the same
// structural position.
func mergeDescNode(dst, src *descNode) {
	if src == nil {
		return
	}
	dst.Count += src.Count
	for t, c := range src.Types {
		dst.Types[t] += c
	}
	if src.Deeper {
		dst.Deeper = true
	}

	if src.MapLike {
		if !dst.MapLike {
			if len(dst.Order) > 0 {
				dst.collapseToMap()
			} else {
				dst.MapLike = true
				dst.MapValue = newDescNode()
			}
		}
		if dst.MapValue == nil {
			dst.MapValue = newDescNode()
		}
		mergeDescNode(dst.MapValue, src.MapValue)
		if src.MapKeys > dst.MapKeys {
			dst.MapKeys = src.MapKeys
		}
		if len(dst.KeyNames) == 0 {
			dst.KeyNames = src.KeyNames
		}
	} else {
		for _, key := range src.Order {
			if dst.MapLike {
				mergeDescNode(dst.MapValue, src.Fields[key])
				continue
			}
			child, ok := dst.Fields[key]
			if !ok {
				child = newDescNode()
				if dst.Fields == nil {
					dst.Fields = map[string]*descNode{}
				}
				dst.Fields[key] = child
				dst.Order = append(dst.Order, key)
			}
			mergeDescNode(child, src.Fields[key])
		}
	}

	if src.ArrMin >= 0 {
		dst.observeLength(src.ArrMin, false)
		dst.observeLength(src.ArrMax, src.Truncated)
	}
	if src.Elem != nil {
		if dst.Elem == nil {
			dst.Elem = newDescNode()
		}
		mergeDescNode(dst.Elem, src.Elem)
	}

	if !dst.HasSample && src.HasSample {
		dst.Sample = src.Sample
		dst.SampleIsString = src.SampleIsString
		dst.HasSample = true
	}
	if src.Overflow {
		dst.Overflow = true
		dst.Values = nil
	} else if !dst.Overflow {
		for v, c := range src.Values {
			dst.observeScalar(v, src.SampleIsString)
			if dst.Values != nil {
				dst.Values[v] += c - 1
			}
		}
	}
	if !dst.FormatSet {
		dst.Format = src.Format
		dst.FormatSet = src.FormatSet
	} else if dst.Format != src.Format {
		dst.Format = ""
	}
}

// descWalker holds the options for one describe run plus whether any limit
// actually truncated the walk, so the output can say so.
type descWalker struct {
	maxDepth int
	sample   int
	sampled  bool
	// stop is set once --sample has been satisfied. Every loop bails out
	// immediately rather than tokenising the rest of the input just to reach
	// the closing delimiter -- on a large file that skip costs as much as the
	// full scan the sample was meant to avoid.
	stop bool
}

// skipContainer consumes tokens until the container whose opening delimiter was
// already read is closed.
func skipContainer(dec *json.Decoder) error {
	depth := 1
	for depth > 0 {
		t, err := dec.Token()
		if err != nil {
			return err
		}
		if d, ok := t.(json.Delim); ok {
			if d == '{' || d == '[' {
				depth++
			} else {
				depth--
			}
		}
	}
	return nil
}

func (w *descWalker) walkValue(dec *json.Decoder, n *descNode, depth int) error {
	t, err := dec.Token()
	if err != nil {
		return err
	}
	n.Count++

	switch v := t.(type) {
	case json.Delim:
		switch v {
		case '{':
			n.addType("object")
			if w.maxDepth > 0 && depth >= w.maxDepth {
				n.Deeper = true
				return skipContainer(dec)
			}
			return w.walkObject(dec, n, depth)
		case '[':
			n.addType("array")
			if w.maxDepth > 0 && depth >= w.maxDepth {
				n.Deeper = true
				return skipContainer(dec)
			}
			return w.walkArray(dec, n, depth)
		}
		return fmt.Errorf("unexpected delimiter %v", v)
	case string:
		n.addType("string")
		n.observeScalar(v, true)
		n.observeFormat(v)
	case json.Number:
		n.addType("number")
		n.observeScalar(v.String(), false)
	case bool:
		n.addType("boolean")
		n.observeScalar(strconv.FormatBool(v), false)
	case nil:
		n.addType("null")
	}
	return nil
}

func (w *descWalker) walkObject(dec *json.Decoder, n *descNode, depth int) error {
	keys := 0
	for dec.More() {
		if w.stop {
			return nil
		}
		kt, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := kt.(string)
		if !ok {
			return fmt.Errorf("expected object key, got %v", kt)
		}
		keys++

		if !n.MapLike && n.Fields[key] == nil && n.shouldCollapse() {
			n.collapseToMap()
		}

		if n.MapLike {
			if n.MapValue == nil {
				n.MapValue = newDescNode()
			}
			if len(n.KeyNames) < 3 {
				n.KeyNames = append(n.KeyNames, key)
			}
			if err := w.walkValue(dec, n.MapValue, depth+1); err != nil {
				return err
			}
			continue
		}

		child, ok := n.Fields[key]
		if !ok {
			child = newDescNode()
			if n.Fields == nil {
				n.Fields = map[string]*descNode{}
			}
			n.Fields[key] = child
			n.Order = append(n.Order, key)
		}
		if err := w.walkValue(dec, child, depth+1); err != nil {
			return err
		}
	}
	if w.stop {
		return nil
	}
	if _, err := dec.Token(); err != nil {
		return err
	}
	if keys > n.MapKeys {
		n.MapKeys = keys
	}
	return nil
}

func (w *descWalker) walkArray(dec *json.Decoder, n *descNode, depth int) error {
	length := 0
	for dec.More() {
		if w.stop {
			return nil
		}
		if w.sample > 0 && length >= w.sample {
			w.sampled = true
			w.stop = true
			n.observeLength(length, true)
			return nil
		}
		if n.Elem == nil {
			n.Elem = newDescNode()
		}
		// Array elements sit at the same depth as the array itself: the tree
		// renders them on one line ("array[4] of object(6)"), so --maxDepth
		// should not spend a level on the element.
		if err := w.walkValue(dec, n.Elem, depth); err != nil {
			return err
		}
		length++
	}
	if w.stop {
		return nil
	}
	if _, err := dec.Token(); err != nil {
		return err
	}
	n.observeLength(length, false)
	return nil
}

// pathSeg is one step of a --path expression. index is -1 unless the step is a
// concrete array index; any is true for the [] "every element" step.
type pathSeg struct {
	name  string
	index int
	any   bool
}

func parseDescribePath(p string) ([]pathSeg, error) {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "$")
	segs := []pathSeg{}
	i := 0
	for i < len(p) {
		switch p[i] {
		case '.':
			i++
		case '[':
			j := strings.IndexByte(p[i:], ']')
			if j < 0 {
				return nil, fmt.Errorf("unclosed '[' in path")
			}
			inner := strings.TrimSpace(p[i+1 : i+j])
			if inner == "" || inner == "*" {
				segs = append(segs, pathSeg{index: -1, any: true})
			} else {
				idx, err := strconv.Atoi(inner)
				if err != nil {
					return nil, fmt.Errorf("invalid array index '%s' in path", inner)
				}
				segs = append(segs, pathSeg{index: idx})
			}
			i += j + 1
		default:
			j := i
			for j < len(p) && p[j] != '.' && p[j] != '[' {
				j++
			}
			name := p[i:j]
			if name != "" {
				if idx, err := strconv.Atoi(name); err == nil {
					segs = append(segs, pathSeg{index: idx})
				} else {
					segs = append(segs, pathSeg{name: name, index: -1})
				}
			}
			i = j
		}
	}
	return segs, nil
}

// seek walks down to every value matching the path and describes those into n,
// skipping everything else without materialising it. It returns how many values
// matched, so a path that exists in no record can be reported as an error.
func (w *descWalker) seek(dec *json.Decoder, segs []pathSeg, n *descNode) (int, error) {
	if len(segs) == 0 {
		if err := w.walkValue(dec, n, 0); err != nil {
			return 0, err
		}
		return 1, nil
	}

	seg := segs[0]
	t, err := dec.Token()
	if err != nil {
		return 0, err
	}
	d, ok := t.(json.Delim)
	if !ok {
		return 0, nil
	}

	matches := 0
	if seg.name != "" {
		if d != '{' {
			return 0, skipContainer(dec)
		}
		for dec.More() {
			if w.stop {
				return matches, nil
			}
			kt, err := dec.Token()
			if err != nil {
				return matches, err
			}
			if kt == seg.name {
				m, err := w.seek(dec, segs[1:], n)
				matches += m
				if err != nil {
					return matches, err
				}
			} else if err := skipValue(dec); err != nil {
				return matches, err
			}
		}
	} else {
		if d != '[' {
			return 0, skipContainer(dec)
		}
		i := 0
		for dec.More() {
			if w.stop {
				return matches, nil
			}
			if seg.any && w.sample > 0 && i >= w.sample {
				w.sampled = true
				w.stop = true
				return matches, nil
			}
			if seg.any || i == seg.index {
				m, err := w.seek(dec, segs[1:], n)
				matches += m
				if err != nil {
					return matches, err
				}
			} else if err := skipValue(dec); err != nil {
				return matches, err
			}
			i++
		}
	}
	if w.stop {
		return matches, nil
	}
	if _, err := dec.Token(); err != nil {
		return matches, err
	}
	return matches, nil
}

// ---- rendering ----------------------------------------------------------

// renderType builds the type column. withElem appends "of <element>" for
// arrays, which the tree wants inline and the path list does not.
func renderType(n *descNode, withElem bool) string {
	parts := []string{}
	for _, t := range n.typeNames() {
		switch t {
		case "object":
			switch {
			case n.Deeper:
				parts = append(parts, "object(...)")
			case n.MapLike:
				parts = append(parts, fmt.Sprintf("object<%d keys>", n.MapKeys))
			default:
				parts = append(parts, fmt.Sprintf("object(%d)", len(n.Order)))
			}
		case "array":
			parts = append(parts, renderArrayType(n, withElem))
		case "string":
			if n.FormatSet && n.Format != "" {
				parts = append(parts, "string<"+n.Format+">")
			} else {
				parts = append(parts, "string")
			}
		default:
			parts = append(parts, t)
		}
	}
	if len(parts) == 0 {
		return "unknown"
	}
	return strings.Join(parts, "|")
}

func renderArrayType(n *descNode, withElem bool) string {
	if n.Deeper {
		return "array[...]"
	}
	length := ""
	switch {
	case n.ArrMin < 0:
		length = "0"
	case n.ArrMin == n.ArrMax:
		length = strconv.Itoa(n.ArrMax)
	default:
		length = fmt.Sprintf("%d..%d", n.ArrMin, n.ArrMax)
	}
	if n.Truncated {
		length += "+, sampled"
	}
	out := "array[" + length + "]"
	if withElem && n.Elem != nil {
		out += " of " + renderType(n.Elem, true)
	}
	return out
}

// enumValues returns the complete value list for a leaf that qualifies as an
// enum, or nil. Values must repeat -- otherwise every unique id in a small
// document would be reported as an enum. Only strings and booleans qualify:
// a numeric field with a few repeated values is almost never a vocabulary, and
// listing it buries the fields that are.
func enumValues(n *descNode) []string {
	if n.Overflow || len(n.Values) == 0 || n.has("object") || n.has("array") || n.has("number") {
		return nil
	}
	// A single repeated value is a constant, not a vocabulary; showing it as an
	// enum would make every constant field in a small document look like one.
	if len(n.Values) < 2 {
		return nil
	}
	total := 0
	for _, c := range n.Values {
		total += c
	}
	if total <= len(n.Values) {
		return nil
	}
	values := make([]string, 0, len(n.Values))
	for v := range n.Values {
		values = append(values, v)
	}
	sort.Slice(values, func(i, j int) bool {
		if n.Values[values[i]] != n.Values[values[j]] {
			return n.Values[values[i]] > n.Values[values[j]]
		}
		return values[i] < values[j]
	})
	return values
}

func truncateSample(s string) string {
	runes := []rune(s)
	if len(runes) <= sampleMaxLen {
		return s
	}
	return string(runes[:sampleMaxLen]) + "..."
}

// renderNote builds the right-hand column: the enum list when the field is an
// enum, otherwise one example value. The second result says which it is, so the
// two can be coloured differently.
func renderNote(n *descNode, showValues bool) (string, bool) {
	if !showValues {
		return "", false
	}
	if values := enumValues(n); values != nil {
		joined := strings.Join(values, "|")
		if len(joined) <= enumInlineMaxLen {
			return joined, true
		}
		short := []string{}
		width := 0
		for _, v := range values {
			if width+len(v) > enumInlineMaxLen {
				break
			}
			short = append(short, v)
			width += len(v) + 1
		}
		return fmt.Sprintf("%s|... (%d values)", strings.Join(short, "|"), len(values)), true
	}
	if n.HasSample {
		if n.SampleIsString {
			return strconv.Quote(truncateSample(n.Sample)), false
		}
		return truncateSample(n.Sample), false
	}
	return "", false
}

// typedValue turns a stored raw scalar back into its JSON type, for the data
// output formats.
func typedValue(raw string, isString bool) interface{} {
	if isString {
		return raw
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return raw
	}
	return parsed
}

func typedEnum(n *descNode, values []string) []interface{} {
	out := make([]interface{}, 0, len(values))
	for _, v := range values {
		out = append(out, typedValue(v, n.SampleIsString))
	}
	return out
}

// treeNote is renderNote for the tree view, where an array of scalars has no
// child row of its own -- the element's values belong on the array's line.
func treeNote(n *descNode, showValues bool) (string, bool) {
	if n.has("array") && !n.has("object") && n.Elem != nil && !n.Elem.isContainer() {
		return renderNote(n.Elem, showValues)
	}
	return renderNote(n, showValues)
}

// presence renders the optional marker for a field seen in only some of its
// parent's occurrences.
func presence(child *descNode, parentCount int) string {
	if parentCount <= 0 || child.Count >= parentCount {
		return ""
	}
	pct := int(float64(child.Count) / float64(parentCount) * 100)
	if pct < 1 {
		pct = 1
	}
	return fmt.Sprintf(" (%d%%)", pct)
}

// structural resolves through arrays to the node that actually carries fields,
// so an array of objects shows the object's fields as its children.
func structural(n *descNode) *descNode {
	if n.has("object") {
		return n
	}
	if n.has("array") && n.Elem != nil {
		return structural(n.Elem)
	}
	return n
}

// treeRow keeps the parts of a line separate so each can be coloured on its own
// and so column widths are measured on the visible text, not on escape codes.
type treeRow struct {
	glyph string // tree branch characters, empty in the path list
	name  string // field name, full path, or "$"
	mark  string // presence marker such as " (25%)"
	typ   string
	note  string
	enum  bool // note is a value list rather than a single example
}

func (r treeRow) plainLabel() string { return r.glyph + r.name + r.mark }

func buildTree(n *descNode, label, prefix string, showValues bool, rows *[]treeRow) {
	note, isEnum := treeNote(n, showValues)
	*rows = append(*rows, treeRow{glyph: prefix, name: label, typ: renderType(n, true), note: note, enum: isEnum})
	buildTreeChildren(n, prefix, showValues, rows)
}

func buildTreeChildren(n *descNode, prefix string, showValues bool, rows *[]treeRow) {
	target := structural(n)
	if target.Deeper {
		return
	}
	if target.MapLike && target.MapValue != nil {
		note := ""
		if showValues && len(target.KeyNames) > 0 {
			note = "keys: " + strings.Join(target.KeyNames, ", ") + ", ..."
		}
		*rows = append(*rows, treeRow{
			glyph: prefix + "└─ ",
			name:  "*",
			typ:   renderType(target.MapValue, true),
			note:  note,
		})
		buildTreeChildren(target.MapValue, prefix+"   ", showValues, rows)
		return
	}
	for i, key := range target.Order {
		child := target.Fields[key]
		last := i == len(target.Order)-1
		branch := "├─ "
		next := "│  "
		if last {
			branch = "└─ "
			next = "   "
		}
		name := key
		mark := presence(child, target.Count)
		if mark != "" {
			name += "?"
		}
		note, isEnum := treeNote(child, showValues)
		*rows = append(*rows, treeRow{
			glyph: prefix + branch,
			name:  name,
			mark:  mark,
			typ:   renderType(child, true),
			note:  note,
			enum:  isEnum,
		})
		buildTreeChildren(child, prefix+next, showValues, rows)
	}
}

// printRows lays out the two aligned columns. Padding is computed from the
// plain text and colour is applied afterwards, so escape codes never shift a
// column.
func printRows(rows []treeRow, records int, sampled bool) {
	labelWidth := 0
	typeWidth := 0
	for _, r := range rows {
		if w := len([]rune(r.plainLabel())); w > labelWidth {
			labelWidth = w
		}
		if w := len([]rune(r.typ)); w > typeWidth {
			typeWidth = w
		}
	}

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	if records > 1 {
		fmt.Fprintln(out, paint(fmt.Sprintf("%d records", records), colorHeader))
	}
	if sampled {
		fmt.Fprintln(out, paint("counts are lower bounds -- input was sampled", colorHeader))
	}
	for _, r := range rows {
		line := paint(r.glyph, colorGlyph) + paint(r.name, colorName) + paint(r.mark, colorMark) +
			spaces(labelWidth-len([]rune(r.plainLabel()))) + "  " + paint(r.typ, colorType)
		if r.note != "" {
			note := colorSample
			if r.enum {
				note = colorEnum
			}
			line += spaces(typeWidth-len([]rune(r.typ))) + "  " + paint(r.note, note)
		}
		fmt.Fprintln(out, line)
	}
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}

func printTree(root *descNode, records int, showValues bool, sampled bool) {
	rows := []treeRow{}
	buildTree(root, "$", "", showValues, &rows)
	printRows(rows, records, sampled)
}

func buildPaths(n *descNode, path string, showValues bool, rows *[]treeRow) {
	note, isEnum := renderNote(n, showValues)
	*rows = append(*rows, treeRow{name: path, typ: renderType(n, false), note: note, enum: isEnum})
	if n.Deeper {
		return
	}
	if n.has("object") {
		if n.MapLike {
			if n.MapValue != nil {
				buildPaths(n.MapValue, path+".*", showValues, rows)
			}
		} else {
			for _, key := range n.Order {
				child := n.Fields[key]
				label := key
				if presence(child, n.Count) != "" {
					label += "?"
				}
				buildPaths(child, path+"."+label, showValues, rows)
			}
		}
	}
	if n.has("array") && n.Elem != nil {
		buildPaths(n.Elem, path+"[]", showValues, rows)
	}
}

func printPaths(root *descNode, records int, showValues bool, sampled bool) {
	rows := []treeRow{}
	buildPaths(root, "$", showValues, &rows)
	printRows(rows, records, sampled)
}

// toJSONShape renders the inferred structure as data, for piping into other
// tools rather than reading.
func toJSONShape(n *descNode, showValues bool) map[string]interface{} {
	out := map[string]interface{}{}

	types := n.typeNames()
	if len(types) == 1 {
		out["type"] = types[0]
	} else if len(types) > 1 {
		out["type"] = types
	}
	out["count"] = n.Count
	if n.Deeper {
		out["truncated"] = "maxDepth"
	}

	if n.has("object") && !n.Deeper {
		if n.MapLike {
			out["keys"] = n.MapKeys
			out["keyed"] = true
			if len(n.KeyNames) > 0 && showValues {
				out["sampleKeys"] = n.KeyNames
			}
			if n.MapValue != nil {
				out["values"] = toJSONShape(n.MapValue, showValues)
			}
		} else {
			fields := map[string]interface{}{}
			required := []string{}
			for _, key := range n.Order {
				child := n.Fields[key]
				fields[key] = toJSONShape(child, showValues)
				if child.Count >= n.Count {
					required = append(required, key)
				}
			}
			out["fields"] = fields
			out["required"] = required
		}
	}

	if n.has("array") && !n.Deeper {
		length := map[string]interface{}{"min": maxInt(n.ArrMin, 0), "max": n.ArrMax}
		if n.Truncated {
			length["sampled"] = true
		}
		out["length"] = length
		if n.Elem != nil {
			out["items"] = toJSONShape(n.Elem, showValues)
		}
	}

	if n.FormatSet && n.Format != "" {
		out["format"] = n.Format
	}
	if showValues {
		if values := enumValues(n); values != nil {
			out["enum"] = typedEnum(n, values)
		}
		if n.HasSample {
			out["sample"] = typedValue(n.Sample, n.SampleIsString)
		}
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// toJSONSchema renders draft 2020-12, for validation and codegen tools.
func toJSONSchema(n *descNode, showValues bool) map[string]interface{} {
	out := map[string]interface{}{}

	types := n.typeNames()
	if len(types) == 1 {
		out["type"] = types[0]
	} else if len(types) > 1 {
		sorted := append([]string{}, types...)
		sort.Strings(sorted)
		out["type"] = sorted
	}

	if n.has("object") && !n.Deeper {
		if n.MapLike {
			if n.MapValue != nil {
				out["additionalProperties"] = toJSONSchema(n.MapValue, showValues)
			}
		} else {
			properties := map[string]interface{}{}
			required := []string{}
			for _, key := range n.Order {
				child := n.Fields[key]
				properties[key] = toJSONSchema(child, showValues)
				if child.Count >= n.Count {
					required = append(required, key)
				}
			}
			out["properties"] = properties
			if len(required) > 0 {
				out["required"] = required
			}
		}
	}

	if n.has("array") && !n.Deeper && n.Elem != nil {
		out["items"] = toJSONSchema(n.Elem, showValues)
	}

	if n.FormatSet && n.Format != "" {
		out["format"] = n.Format
	}
	if showValues {
		if values := enumValues(n); values != nil {
			out["enum"] = typedEnum(n, values)
		}
	}
	return out
}

// selectSpec emits the shared 2table/render field notation, so the output of
// describe is a ready argument for `aux4 json select`.
func selectSpec(n *descNode, depth, maxDepth int) string {
	target := structural(n)
	if !target.has("object") || target.MapLike || target.Deeper {
		return ""
	}
	fields := []string{}
	for _, key := range target.Order {
		child := structural(target.Fields[key])
		if child.has("object") && !child.MapLike && !child.Deeper && (maxDepth <= 0 || depth+1 < maxDepth) {
			if nested := selectSpec(child, depth+1, maxDepth); nested != "" {
				fields = append(fields, key+"["+nested+"]")
				continue
			}
		}
		fields = append(fields, key)
	}
	return strings.Join(fields, ",")
}

func runDescribe(args []string) {
	path := "$"
	format := "tree"
	maxDepth := 0
	sample := 0
	showValues := true

	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		path = strings.TrimSpace(args[0])
	}
	if len(args) > 1 && strings.TrimSpace(args[1]) != "" {
		format = strings.TrimSpace(args[1])
	}
	if len(args) > 2 && strings.TrimSpace(args[2]) != "" {
		v, err := strconv.Atoi(strings.TrimSpace(args[2]))
		if err != nil || v < 0 {
			fmt.Fprintf(os.Stderr, "Error: maxDepth must be a non-negative number\n")
			os.Exit(1)
		}
		maxDepth = v
	}
	if len(args) > 3 && strings.TrimSpace(args[3]) != "" {
		v, err := strconv.Atoi(strings.TrimSpace(args[3]))
		if err != nil || v < 0 {
			fmt.Fprintf(os.Stderr, "Error: sample must be a non-negative number\n")
			os.Exit(1)
		}
		sample = v
	}
	if len(args) > 4 && strings.TrimSpace(args[4]) == "false" {
		showValues = false
	}
	colorMode := "auto"
	if len(args) > 5 && strings.TrimSpace(args[5]) != "" {
		colorMode = strings.TrimSpace(args[5])
	}

	switch format {
	case "tree", "paths", "json", "jsonschema", "select":
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown format '%s' (expected tree, paths, json, jsonschema or select)\n", format)
		os.Exit(1)
	}

	switch colorMode {
	case "auto", "always", "never":
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown color '%s' (expected auto, always or never)\n", colorMode)
		os.Exit(1)
	}
	if format == "tree" || format == "paths" {
		useColor = resolveColor(colorMode)
	}

	segs, err := parseDescribePath(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	walker := &descWalker{maxDepth: maxDepth, sample: sample}
	root := newDescNode()

	dec := json.NewDecoder(bufio.NewReaderSize(os.Stdin, 1024*1024))
	dec.UseNumber()

	records := 0
	matched := 0
	for dec.More() {
		if len(segs) == 0 {
			if err := walker.walkValue(dec, root, 0); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
				os.Exit(1)
			}
			matched++
		} else {
			m, err := walker.seek(dec, segs, root)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
				os.Exit(1)
			}
			matched += m
		}
		records++
		if walker.stop {
			break
		}
		if sample > 0 && records >= sample {
			walker.sampled = true
			break
		}
	}

	if records == 0 {
		fmt.Fprintln(os.Stderr, "Error: no JSON input")
		os.Exit(1)
	}
	if matched == 0 {
		fmt.Fprintf(os.Stderr, "Error: path '%s' not found\n", path)
		os.Exit(1)
	}

	switch format {
	case "tree":
		printTree(root, records, showValues, walker.sampled)
	case "paths":
		printPaths(root, records, showValues, walker.sampled)
	case "json":
		out, err := json.Marshal(toJSONShape(root, showValues))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
	case "jsonschema":
		schema := toJSONSchema(root, showValues)
		schema["$schema"] = "https://json-schema.org/draft/2020-12/schema"
		out, err := json.Marshal(schema)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
	case "select":
		spec := selectSpec(root, 0, maxDepth)
		if spec == "" {
			fmt.Fprintln(os.Stderr, "Error: select format needs an object (or an array of objects) at the path")
			os.Exit(1)
		}
		fmt.Println(spec)
	}
}
