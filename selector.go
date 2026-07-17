package cheerio

import (
	"strconv"
	"strings"
)

// selectorList is a comma-separated group of complex selectors. An element
// matches the list if it matches any member.
type selectorList []complexSelector

// complexSelector is a sequence of compound selectors joined by combinators,
// stored right-to-left evaluation friendly (in source order here).
type complexSelector struct {
	// parts[0] has combinator combNone; subsequent parts carry the combinator
	// that links them to the previous part.
	parts []selectorPart
}

type combinator int

const (
	combNone       combinator = iota
	combDescendant            // space
	combChild                 // >
	combAdjacent              // +
	combSibling               // ~
)

// selectorPart is a single compound selector (e.g. div.foo#bar[x=y]:first-child).
type selectorPart struct {
	comb    combinator
	tag     string // "" or "*" means any type
	id      string
	classes []string
	attrs   []attrMatcher
	pseudos []pseudoMatcher
}

// attrMatcher matches an attribute condition.
type attrMatcher struct {
	key string
	op  string // "" (exists), "=", "^=", "$=", "*=", "~=", "|="
	val string
	ci  bool // case-insensitive match ([a=v i])
}

// pseudoMatcher matches a pseudo-class.
type pseudoMatcher struct {
	name string       // first-child, last-child, nth-child, not, ...
	a, b int          // for nth-child: an+b
	arg  selectorList // for :not(...) and :has(...)
	text string       // for :contains(...)
}

// parseSelector compiles a selector string into a selectorList. A parse
// failure yields an empty list that matches nothing.
func parseSelector(s string) selectorList {
	var list selectorList
	for _, group := range splitTopLevel(s, ',') {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		if cs, ok := parseComplex(group); ok {
			list = append(list, cs)
		}
	}
	return list
}

// splitTopLevel splits s on sep, ignoring separators nested inside [] or ().
func splitTopLevel(s string, sep byte) []string {
	var out []string
	depthBracket, depthParen := 0, 0
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '[':
			depthBracket++
		case ']':
			if depthBracket > 0 {
				depthBracket--
			}
		case '(':
			depthParen++
		case ')':
			if depthParen > 0 {
				depthParen--
			}
		case sep:
			if depthBracket == 0 && depthParen == 0 {
				out = append(out, s[start:i])
				start = i + 1
			}
		}
	}
	out = append(out, s[start:])
	return out
}

// parseComplex parses a sequence of compound selectors and combinators.
func parseComplex(s string) (complexSelector, bool) {
	var cs complexSelector
	i := 0
	pendingComb := combNone
	for i < len(s) {
		// Skip whitespace, tracking descendant combinator.
		wsSeen := false
		for i < len(s) && isSpace(s[i]) {
			i++
			wsSeen = true
		}
		if i >= len(s) {
			break
		}
		// Explicit combinator symbols.
		switch s[i] {
		case '>':
			pendingComb = combChild
			i++
			continue
		case '+':
			pendingComb = combAdjacent
			i++
			continue
		case '~':
			pendingComb = combSibling
			i++
			continue
		}
		if wsSeen && len(cs.parts) > 0 && pendingComb == combNone {
			pendingComb = combDescendant
		}

		part, ni, ok := parseCompound(s, i)
		if !ok {
			return complexSelector{}, false
		}
		if len(cs.parts) == 0 {
			part.comb = combNone
		} else {
			part.comb = pendingComb
		}
		cs.parts = append(cs.parts, part)
		pendingComb = combNone
		i = ni
	}
	if len(cs.parts) == 0 {
		return complexSelector{}, false
	}
	return cs, true
}

// parseCompound parses one compound selector beginning at index i and returns
// the part, the index after it, and success.
func parseCompound(s string, i int) (selectorPart, int, bool) {
	var part selectorPart
	started := false
	for i < len(s) {
		c := s[i]
		switch {
		case c == '*':
			part.tag = "*"
			started = true
			i++
		case c == '#':
			i++
			name, ni := readIdent(s, i)
			if name == "" {
				return part, i, false
			}
			part.id = name
			i = ni
			started = true
		case c == '.':
			i++
			name, ni := readIdent(s, i)
			if name == "" {
				return part, i, false
			}
			part.classes = append(part.classes, name)
			i = ni
			started = true
		case c == '[':
			am, ni, ok := parseAttr(s, i)
			if !ok {
				return part, i, false
			}
			part.attrs = append(part.attrs, am)
			i = ni
			started = true
		case c == ':':
			pm, ni, ok := parsePseudo(s, i)
			if !ok {
				return part, i, false
			}
			part.pseudos = append(part.pseudos, pm)
			i = ni
			started = true
		case c == '|':
			// CSS namespace separator: the local name follows. We have no
			// namespace information, so we match on the local name only.
			i++
			name, ni := readTagName(s, i)
			part.tag = strings.ToLower(name)
			i = ni
			started = true
		case isTagNameStart(c):
			name, ni := readTagName(s, i)
			part.tag = strings.ToLower(name)
			i = ni
			started = true
		default:
			// End of compound (whitespace, combinator, comma, or unknown).
			if started {
				return part, i, true
			}
			return part, i, false
		}
	}
	if !started {
		return part, i, false
	}
	return part, i, true
}

// parseAttr parses "[key]", "[key=val]" and its variants beginning at '['.
func parseAttr(s string, i int) (attrMatcher, int, bool) {
	// s[i] == '['
	end := strings.IndexByte(s[i:], ']')
	if end < 0 {
		return attrMatcher{}, i, false
	}
	body := s[i+1 : i+end]
	next := i + end + 1

	var am attrMatcher
	ops := []string{"^=", "$=", "*=", "~=", "|=", "="}
	opIdx := -1
	opStr := ""
	for _, op := range ops {
		if idx := strings.Index(body, op); idx >= 0 {
			if opIdx == -1 || idx < opIdx || (idx == opIdx && len(op) > len(opStr)) {
				opIdx = idx
				opStr = op
			}
		}
	}
	if opIdx == -1 {
		am.key = strings.ToLower(strings.TrimSpace(body))
		if am.key == "" {
			return am, i, false
		}
		return am, next, true
	}
	am.key = strings.ToLower(strings.TrimSpace(body[:opIdx]))
	am.op = opStr
	val := strings.TrimSpace(body[opIdx+len(opStr):])
	// Detect a trailing case-sensitivity flag (" i" or " s") that sits outside
	// any quotes, e.g. [type="TEXT" i] or [type=text i].
	if n := len(val); n >= 2 && (val[n-2] == ' ' || val[n-2] == '\t') {
		switch val[n-1] {
		case 'i', 'I':
			if cand := strings.TrimSpace(val[:n-2]); cand != "" {
				am.ci = true
				val = cand
			}
		case 's', 'S':
			if cand := strings.TrimSpace(val[:n-2]); cand != "" {
				val = cand
			}
		}
	}
	if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') && val[len(val)-1] == val[0] {
		val = val[1 : len(val)-1]
	}
	am.val = val
	if am.key == "" {
		return am, i, false
	}
	return am, next, true
}

// parsePseudo parses a pseudo-class such as :first-child or :nth-child(2n+1)
// or :not(sel).
func parsePseudo(s string, i int) (pseudoMatcher, int, bool) {
	i++ // skip ':'
	name, ni := readIdent(s, i)
	if name == "" {
		return pseudoMatcher{}, i, false
	}
	name = strings.ToLower(name)
	i = ni
	var pm pseudoMatcher
	pm.name = name

	// Optional argument in parentheses.
	if i < len(s) && s[i] == '(' {
		depth := 0
		start := i + 1
		j := i
		for ; j < len(s); j++ {
			if s[j] == '(' {
				depth++
			} else if s[j] == ')' {
				depth--
				if depth == 0 {
					break
				}
			}
		}
		if j >= len(s) {
			return pm, i, false
		}
		arg := s[start:j]
		i = j + 1
		switch name {
		case "not", "has":
			pm.arg = parseSelector(arg)
		case "contains":
			pm.text = unquoteArg(arg)
		case "nth-child", "nth-last-child", "nth-of-type", "nth-last-of-type":
			a, b, ok := parseNth(arg)
			if !ok {
				return pm, i, false
			}
			pm.a, pm.b = a, b
		default:
			return pm, i, false
		}
	}
	return pm, i, true
}

// unquoteArg trims whitespace and a single layer of matching quotes from a
// pseudo-class argument (used by :contains).
func unquoteArg(arg string) string {
	arg = strings.TrimSpace(arg)
	if len(arg) >= 2 && (arg[0] == '"' || arg[0] == '\'') && arg[len(arg)-1] == arg[0] {
		return arg[1 : len(arg)-1]
	}
	return arg
}

// parseNth parses the argument of nth-child: "odd", "even", integer, or the
// "an+b" microsyntax.
func parseNth(arg string) (int, int, bool) {
	arg = strings.TrimSpace(strings.ToLower(arg))
	switch arg {
	case "odd":
		return 2, 1, true
	case "even":
		return 2, 0, true
	}
	// Pure integer.
	if n, err := strconv.Atoi(arg); err == nil {
		return 0, n, true
	}
	// an+b forms.
	if idx := strings.IndexByte(arg, 'n'); idx >= 0 {
		aStr := strings.TrimSpace(arg[:idx])
		bStr := strings.TrimSpace(arg[idx+1:])
		var a int
		switch aStr {
		case "", "+":
			a = 1
		case "-":
			a = -1
		default:
			v, err := strconv.Atoi(aStr)
			if err != nil {
				return 0, 0, false
			}
			a = v
		}
		b := 0
		if bStr != "" {
			bStr = strings.ReplaceAll(bStr, " ", "")
			v, err := strconv.Atoi(bStr)
			if err != nil {
				return 0, 0, false
			}
			b = v
		}
		return a, b, true
	}
	return 0, 0, false
}

// readIdent reads a CSS identifier (letters, digits, -, _, escaped chars are
// not supported).
func readIdent(s string, i int) (string, int) {
	start := i
	for i < len(s) && isIdentChar(s[i]) {
		i++
	}
	return s[start:i], i
}

// readTagName reads a tag/type name in a selector. Unlike the tokenizer's
// tag-name reader it stops at ':' so pseudo-classes are not absorbed.
func readTagName(s string, i int) (string, int) {
	start := i
	for i < len(s) && isSelectorTagChar(s[i]) {
		i++
	}
	return s[start:i], i
}

// isSelectorTagChar reports whether b can appear in a selector type name.
func isSelectorTagChar(b byte) bool {
	return isTagNameStart(b) || (b >= '0' && b <= '9') || b == '-' || b == '_'
}

func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') || b == '-' || b == '_'
}

// ---------------------------------------------------------------------------
// Matching
// ---------------------------------------------------------------------------

// matches reports whether node n matches any selector in the list.
func (list selectorList) matches(n *Node) bool {
	for _, cs := range list {
		if cs.matches(n) {
			return true
		}
	}
	return false
}

// matches reports whether n matches this complex selector, evaluating
// combinators right-to-left starting from the rightmost compound.
func (cs complexSelector) matches(n *Node) bool {
	if len(cs.parts) == 0 {
		return false
	}
	return cs.matchFrom(n, len(cs.parts)-1)
}

// matchFrom checks that n matches parts[idx] and that the chain of preceding
// parts is satisfied by ancestors/siblings.
func (cs complexSelector) matchFrom(n *Node, idx int) bool {
	part := cs.parts[idx]
	if !part.matchSelf(n) {
		return false
	}
	if idx == 0 {
		return true
	}
	prev := idx - 1
	switch part.comb {
	case combDescendant:
		for a := n.Parent; a != nil && a.Type == ElementNode; a = a.Parent {
			if cs.matchFrom(a, prev) {
				return true
			}
		}
		return false
	case combChild:
		p := n.Parent
		if p != nil && p.Type == ElementNode {
			return cs.matchFrom(p, prev)
		}
		return false
	case combAdjacent:
		sib := previousElementSibling(n)
		if sib != nil {
			return cs.matchFrom(sib, prev)
		}
		return false
	case combSibling:
		for sib := previousElementSibling(n); sib != nil; sib = previousElementSibling(sib) {
			if cs.matchFrom(sib, prev) {
				return true
			}
		}
		return false
	}
	return false
}

// matchSelf tests the compound conditions of a single part against n.
func (part selectorPart) matchSelf(n *Node) bool {
	if n.Type != ElementNode {
		return false
	}
	if part.tag != "" && part.tag != "*" && part.tag != n.Data {
		return false
	}
	if part.id != "" {
		if v, ok := n.AttrValue("id"); !ok || v != part.id {
			return false
		}
	}
	for _, cl := range part.classes {
		if !n.hasClass(cl) {
			return false
		}
	}
	for _, am := range part.attrs {
		if !am.match(n) {
			return false
		}
	}
	for _, pm := range part.pseudos {
		if !pm.match(n) {
			return false
		}
	}
	return true
}

// match tests a single attribute matcher against n.
func (am attrMatcher) match(n *Node) bool {
	v, ok := n.AttrValue(am.key)
	if !ok {
		return false
	}
	want := am.val
	if am.ci {
		v = strings.ToLower(v)
		want = strings.ToLower(want)
	}
	switch am.op {
	case "":
		return true
	case "=":
		return v == want
	case "^=":
		return want != "" && strings.HasPrefix(v, want)
	case "$=":
		return want != "" && strings.HasSuffix(v, want)
	case "*=":
		return want != "" && strings.Contains(v, want)
	case "~=":
		if want == "" {
			return false
		}
		for _, f := range strings.Fields(v) {
			if f == want {
				return true
			}
		}
		return false
	case "|=":
		return v == want || strings.HasPrefix(v, want+"-")
	}
	return false
}

// match tests a pseudo-class against n.
func (pm pseudoMatcher) match(n *Node) bool {
	switch pm.name {
	case "first-child":
		return previousElementSibling(n) == nil
	case "last-child":
		return nextElementSibling(n) == nil
	case "only-child":
		return previousElementSibling(n) == nil && nextElementSibling(n) == nil
	case "first-of-type":
		return previousSiblingOfType(n) == nil
	case "last-of-type":
		return nextSiblingOfType(n) == nil
	case "only-of-type":
		return previousSiblingOfType(n) == nil && nextSiblingOfType(n) == nil
	case "nth-child":
		return matchNth(elementIndex(n), pm.a, pm.b)
	case "nth-last-child":
		return matchNth(elementIndexFromEnd(n), pm.a, pm.b)
	case "nth-of-type":
		return matchNth(typeIndex(n), pm.a, pm.b)
	case "nth-last-of-type":
		return matchNth(typeIndexFromEnd(n), pm.a, pm.b)
	case "empty":
		return len(n.elementChildren()) == 0 && !hasNonSpaceText(n)
	case "parent":
		return len(n.elementChildren()) > 0 || hasNonSpaceText(n)
	case "root":
		return n.Parent != nil && n.Parent.Type == DocumentNode
	case "not":
		return !pm.arg.matches(n)
	case "has":
		found := false
		for _, c := range n.Children {
			walkElements(c, func(e *Node) {
				if !found && pm.arg.matches(e) {
					found = true
				}
			})
		}
		return found
	case "contains":
		var b strings.Builder
		collectText(n, &b)
		return strings.Contains(b.String(), pm.text)
	case "input":
		switch n.Data {
		case "input", "textarea", "select", "button":
			return true
		}
		return false
	case "header":
		switch n.Data {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			return true
		}
		return false
	case "checked":
		if n.Data == "input" {
			t := strings.ToLower(attrOr(n, "type"))
			if t == "checkbox" || t == "radio" {
				_, ok := n.AttrValue("checked")
				return ok
			}
			return false
		}
		if n.Data == "option" {
			_, ok := n.AttrValue("selected")
			return ok
		}
		return false
	case "selected":
		if n.Data != "option" {
			return false
		}
		_, ok := n.AttrValue("selected")
		return ok
	case "disabled":
		_, ok := n.AttrValue("disabled")
		return ok && canBeDisabled(n)
	case "enabled":
		_, ok := n.AttrValue("disabled")
		return !ok && canBeDisabled(n)
	}
	return false
}

// canBeDisabled reports whether n is a control that supports the disabled
// state, used by :enabled and :disabled.
func canBeDisabled(n *Node) bool {
	switch n.Data {
	case "input", "button", "select", "textarea", "optgroup", "option", "fieldset":
		return true
	}
	return false
}

// previousSiblingOfType returns the nearest preceding element sibling of n with
// the same tag name, or nil.
func previousSiblingOfType(n *Node) *Node {
	for sib := previousElementSibling(n); sib != nil; sib = previousElementSibling(sib) {
		if sib.Data == n.Data {
			return sib
		}
	}
	return nil
}

// nextSiblingOfType returns the nearest following element sibling of n with the
// same tag name, or nil.
func nextSiblingOfType(n *Node) *Node {
	for sib := nextElementSibling(n); sib != nil; sib = nextElementSibling(sib) {
		if sib.Data == n.Data {
			return sib
		}
	}
	return nil
}

// typeIndex returns the 1-based position of n among its same-type element
// siblings.
func typeIndex(n *Node) int {
	if n.Parent == nil {
		return 1
	}
	idx := 0
	for _, c := range n.Parent.Children {
		if c.Type == ElementNode && c.Data == n.Data {
			idx++
			if c == n {
				return idx
			}
		}
	}
	return idx
}

// typeIndexFromEnd returns the 1-based position of n among its same-type element
// siblings, counting from the last one.
func typeIndexFromEnd(n *Node) int {
	if n.Parent == nil {
		return 1
	}
	var same []*Node
	for _, c := range n.Parent.Children {
		if c.Type == ElementNode && c.Data == n.Data {
			same = append(same, c)
		}
	}
	for i, c := range same {
		if c == n {
			return len(same) - i
		}
	}
	return 1
}

// matchNth reports whether a 1-based index matches an+b.
func matchNth(index, a, b int) bool {
	if index <= 0 {
		return false
	}
	if a == 0 {
		return index == b
	}
	// index = a*k + b for some integer k >= 0.
	diff := index - b
	if diff%a != 0 {
		return false
	}
	k := diff / a
	return k >= 0
}

// elementIndex returns the 1-based position of n among its element siblings.
func elementIndex(n *Node) int {
	if n.Parent == nil {
		return 1
	}
	idx := 0
	for _, c := range n.Parent.Children {
		if c.Type == ElementNode {
			idx++
			if c == n {
				return idx
			}
		}
	}
	return idx
}

// elementIndexFromEnd returns the 1-based position counting from the last
// element sibling.
func elementIndexFromEnd(n *Node) int {
	if n.Parent == nil {
		return 1
	}
	els := n.Parent.elementChildren()
	for i, c := range els {
		if c == n {
			return len(els) - i
		}
	}
	return 1
}

func previousElementSibling(n *Node) *Node {
	if n.Parent == nil {
		return nil
	}
	var prev *Node
	for _, c := range n.Parent.Children {
		if c == n {
			return prev
		}
		if c.Type == ElementNode {
			prev = c
		}
	}
	return prev
}

func nextElementSibling(n *Node) *Node {
	if n.Parent == nil {
		return nil
	}
	found := false
	for _, c := range n.Parent.Children {
		if found && c.Type == ElementNode {
			return c
		}
		if c == n {
			found = true
		}
	}
	return nil
}

func hasNonSpaceText(n *Node) bool {
	for _, c := range n.Children {
		if c.Type == TextNode && strings.TrimSpace(c.Data) != "" {
			return true
		}
	}
	return false
}
