package cheerio

import "strings"

// Document is a parsed HTML document. It is the entry point for querying and
// wraps the synthetic root node produced by the parser.
type Document struct {
	root *Node
}

// Load parses an HTML fragment or full document and returns a Document.
// Parsing never fails: malformed input is recovered as a browser would.
func Load(html string) *Document {
	return &Document{root: parse(html)}
}

// Root returns a Selection containing the document root node's element
// children (the top-level elements of the parsed input).
func (d *Document) Root() *Selection {
	return &Selection{doc: d, nodes: []*Node{d.root}}
}

// Find returns all elements in the document matching the selector.
func (d *Document) Find(selector string) *Selection {
	sel := parseSelector(selector)
	var out []*Node
	walkElements(d.root, func(n *Node) {
		if sel.matches(n) {
			out = append(out, n)
		}
	})
	return &Selection{doc: d, nodes: out}
}

// Selection is a set of matched DOM nodes with jQuery-style traversal and
// accessor methods. Selections are immutable with respect to their node set:
// traversal methods return new Selections. The order of nodes is document
// order and duplicates are removed.
type Selection struct {
	doc   *Document
	nodes []*Node
	prev  *Selection // the selection this one was derived from (for End/AddBack)
}

// newSelection builds a Selection, de-duplicating while preserving order.
func (d *Document) newSelection(nodes []*Node) *Selection {
	return &Selection{doc: d, nodes: unique(nodes)}
}

// Length returns the number of nodes in the selection.
func (s *Selection) Length() int { return len(s.nodes) }

// Nodes returns the underlying nodes. Callers must not mutate the slice.
func (s *Selection) Nodes() []*Node { return s.nodes }

// Find returns the descendants of the selected nodes that match the selector.
//
// A selector member that begins with a combinator is treated as relative to
// each selected node (its :scope): "> li" matches immediate element children,
// "+ .b" the immediately following sibling and "~ .b" any following sibling.
// Ordinary members continue to match descendants only.
func (s *Selection) Find(selector string) *Selection {
	list := parseSelector(selector)
	var normal, scoped selectorList
	for _, cs := range list {
		if cs.lead == combNone {
			normal = append(normal, cs)
		} else {
			scoped = append(scoped, cs)
		}
	}
	var out []*Node
	if len(normal) > 0 {
		for _, n := range s.nodes {
			for _, c := range n.Children {
				walkElements(c, func(e *Node) {
					if normal.matches(e) {
						out = append(out, e)
					}
				})
			}
		}
	}
	if len(scoped) > 0 && s.doc != nil {
		walkElements(s.doc.root, func(e *Node) {
			for _, cs := range scoped {
				for _, scope := range s.nodes {
					if cs.matchScoped(e, scope) {
						out = append(out, e)
						return
					}
				}
			}
		})
	}
	return s.pushStack(out)
}

// Children returns the immediate element children of the selected nodes,
// optionally filtered by a selector.
func (s *Selection) Children(selector ...string) *Selection {
	var out []*Node
	for _, n := range s.nodes {
		out = append(out, n.elementChildren()...)
	}
	return s.pushStack(out).maybeFilter(selector)
}

// Parent returns the direct parent of each selected node (elements only).
func (s *Selection) Parent(selector ...string) *Selection {
	var out []*Node
	for _, n := range s.nodes {
		if n.Parent != nil && n.Parent.Type == ElementNode {
			out = append(out, n.Parent)
		}
	}
	return s.pushStack(out).maybeFilter(selector)
}

// Parents returns all ancestors of each selected node, nearest first within
// each node's chain, optionally filtered by a selector.
func (s *Selection) Parents(selector ...string) *Selection {
	var out []*Node
	for _, n := range s.nodes {
		for a := n.Parent; a != nil && a.Type == ElementNode; a = a.Parent {
			out = append(out, a)
		}
	}
	return s.pushStack(out).maybeFilter(selector)
}

// Closest returns, for each selected node, the nearest ancestor-or-self that
// matches the selector.
func (s *Selection) Closest(selector string) *Selection {
	sel := parseSelector(selector)
	var out []*Node
	for _, n := range s.nodes {
		for a := n; a != nil && a.Type == ElementNode; a = a.Parent {
			if sel.matches(a) {
				out = append(out, a)
				break
			}
		}
	}
	return s.pushStack(out)
}

// Siblings returns the element siblings of each selected node (excluding the
// node itself), optionally filtered by a selector.
func (s *Selection) Siblings(selector ...string) *Selection {
	var out []*Node
	for _, n := range s.nodes {
		if n.Parent == nil {
			continue
		}
		for _, c := range n.Parent.elementChildren() {
			if c != n {
				out = append(out, c)
			}
		}
	}
	return s.pushStack(out).maybeFilter(selector)
}

// Next returns the immediately following element sibling of each node.
func (s *Selection) Next(selector ...string) *Selection {
	var out []*Node
	for _, n := range s.nodes {
		if sib := nextElementSibling(n); sib != nil {
			out = append(out, sib)
		}
	}
	return s.pushStack(out).maybeFilter(selector)
}

// Prev returns the immediately preceding element sibling of each node.
func (s *Selection) Prev(selector ...string) *Selection {
	var out []*Node
	for _, n := range s.nodes {
		if sib := previousElementSibling(n); sib != nil {
			out = append(out, sib)
		}
	}
	return s.pushStack(out).maybeFilter(selector)
}

// NextAll returns all following element siblings of each node.
func (s *Selection) NextAll(selector ...string) *Selection {
	var out []*Node
	for _, n := range s.nodes {
		for sib := nextElementSibling(n); sib != nil; sib = nextElementSibling(sib) {
			out = append(out, sib)
		}
	}
	return s.pushStack(out).maybeFilter(selector)
}

// PrevAll returns all preceding element siblings of each node.
func (s *Selection) PrevAll(selector ...string) *Selection {
	var out []*Node
	for _, n := range s.nodes {
		for sib := previousElementSibling(n); sib != nil; sib = previousElementSibling(sib) {
			out = append(out, sib)
		}
	}
	return s.pushStack(out).maybeFilter(selector)
}

// Filter returns the subset of selected nodes that match the selector.
func (s *Selection) Filter(selector string) *Selection {
	sel := parseSelector(selector)
	var out []*Node
	for _, n := range s.nodes {
		if sel.matches(n) {
			out = append(out, n)
		}
	}
	return s.pushStack(out)
}

// FilterFunc returns the subset for which the predicate returns true. The
// index is the position within the current selection.
func (s *Selection) FilterFunc(pred func(i int, n *Node) bool) *Selection {
	var out []*Node
	for i, n := range s.nodes {
		if pred(i, n) {
			out = append(out, n)
		}
	}
	return s.pushStack(out)
}

// Not returns the selected nodes that do NOT match the selector.
func (s *Selection) Not(selector string) *Selection {
	sel := parseSelector(selector)
	var out []*Node
	for _, n := range s.nodes {
		if !sel.matches(n) {
			out = append(out, n)
		}
	}
	return s.pushStack(out)
}

// Is reports whether at least one selected node matches the selector.
func (s *Selection) Is(selector string) bool {
	sel := parseSelector(selector)
	for _, n := range s.nodes {
		if sel.matches(n) {
			return true
		}
	}
	return false
}

// Has returns the selected nodes that have a descendant matching the selector.
func (s *Selection) Has(selector string) *Selection {
	sel := parseSelector(selector)
	var out []*Node
	for _, n := range s.nodes {
		found := false
		for _, c := range n.Children {
			walkElements(c, func(e *Node) {
				if !found && sel.matches(e) {
					found = true
				}
			})
		}
		if found {
			out = append(out, n)
		}
	}
	return s.pushStack(out)
}

// Eq returns a Selection containing only the node at index i (negative counts
// from the end). An out-of-range index yields an empty Selection.
func (s *Selection) Eq(i int) *Selection {
	if i < 0 {
		i += len(s.nodes)
	}
	if i < 0 || i >= len(s.nodes) {
		return &Selection{doc: s.doc, prev: s}
	}
	return &Selection{doc: s.doc, nodes: []*Node{s.nodes[i]}, prev: s}
}

// First returns the first node as a Selection.
func (s *Selection) First() *Selection { return s.Eq(0) }

// Last returns the last node as a Selection.
func (s *Selection) Last() *Selection { return s.Eq(len(s.nodes) - 1) }

// Slice returns the nodes in [start, end) as a Selection. Negative indices
// count from the end; out-of-range bounds are clamped.
func (s *Selection) Slice(start, end int) *Selection {
	n := len(s.nodes)
	if start < 0 {
		start += n
	}
	if end < 0 {
		end += n
	}
	if start < 0 {
		start = 0
	}
	if end > n {
		end = n
	}
	if start >= end {
		return &Selection{doc: s.doc, prev: s}
	}
	return &Selection{doc: s.doc, nodes: append([]*Node(nil), s.nodes[start:end]...), prev: s}
}

// Each invokes fn for every selected node in order and returns the receiver
// for chaining.
func (s *Selection) Each(fn func(i int, n *Node)) *Selection {
	for i, n := range s.nodes {
		fn(i, n)
	}
	return s
}

// Map applies fn to each node and returns the collected results.
func (s *Selection) Map(fn func(i int, n *Node) string) []string {
	out := make([]string, 0, len(s.nodes))
	for i, n := range s.nodes {
		out = append(out, fn(i, n))
	}
	return out
}

// Get returns the node at index i (negative counts from the end), or nil.
func (s *Selection) Get(i int) *Node {
	if i < 0 {
		i += len(s.nodes)
	}
	if i < 0 || i >= len(s.nodes) {
		return nil
	}
	return s.nodes[i]
}

// Add returns a new Selection combining the receiver with nodes matching the
// selector across the whole document.
func (s *Selection) Add(selector string) *Selection {
	extra := s.doc.Find(selector).nodes
	combined := append(append([]*Node(nil), s.nodes...), extra...)
	return s.pushStack(orderedUnique(s.doc, combined))
}

// End returns the Selection that this one was derived from, restoring the
// previous stage of a traversal chain (jQuery's .end()). If the receiver was
// not derived from another Selection (for example a top-level Document.Find),
// End returns the document root selection.
func (s *Selection) End() *Selection {
	if s.prev != nil {
		return s.prev
	}
	return s.doc.Root()
}

// AddBack adds the previous set of matched elements (the Selection this one was
// derived from) to the current set, optionally filtered by a selector. It is
// the inverse companion to End: where End discards the current stage, AddBack
// merges it with the prior one.
func (s *Selection) AddBack(selector ...string) *Selection {
	if s.prev == nil {
		return s
	}
	prev := s.prev
	if len(selector) > 0 && strings.TrimSpace(selector[0]) != "" {
		prev = prev.Filter(selector[0])
	}
	combined := append(append([]*Node(nil), s.nodes...), prev.nodes...)
	return s.pushStack(orderedUnique(s.doc, combined))
}

// maybeFilter applies an optional trailing selector argument.
func (s *Selection) maybeFilter(selector []string) *Selection {
	if len(selector) == 0 || strings.TrimSpace(selector[0]) == "" {
		return s
	}
	return s.Filter(selector[0])
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// walkElements invokes fn for n (if an element) and every element descendant,
// in document order.
func walkElements(n *Node, fn func(*Node)) {
	if n.Type == ElementNode {
		fn(n)
	}
	for _, c := range n.Children {
		walkElements(c, fn)
	}
}

// unique removes duplicate nodes while preserving first-seen order.
func unique(nodes []*Node) []*Node {
	if len(nodes) < 2 {
		return nodes
	}
	seen := make(map[*Node]bool, len(nodes))
	out := nodes[:0:0]
	for _, n := range nodes {
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}
