package cheerio

// pushStack builds a derived Selection from nodes (de-duplicated, preserving
// first-seen order) and records the receiver as its previous selection so that
// End and AddBack can walk back up a traversal chain.
func (s *Selection) pushStack(nodes []*Node) *Selection {
	ns := s.doc.newSelection(nodes)
	ns.prev = s
	return ns
}

// Contents returns the children of each selected node, including text and
// comment nodes (not only elements). This mirrors jQuery's .contents().
func (s *Selection) Contents() *Selection {
	var out []*Node
	for _, n := range s.nodes {
		out = append(out, n.Children...)
	}
	return s.pushStack(out)
}

// Index returns the position of the first selected node among its element
// siblings (0-based), or -1 if the selection is empty. It is the zero-argument
// form of jQuery's .index().
func (s *Selection) Index() int {
	if len(s.nodes) == 0 {
		return -1
	}
	return elementIndex(s.nodes[0]) - 1
}

// IndexSelector returns the 0-based position of the first selected node within
// the set of elements matching selector across the document, or -1 if it is not
// found. It corresponds to jQuery's .index(selector).
func (s *Selection) IndexSelector(selector string) int {
	if len(s.nodes) == 0 {
		return -1
	}
	target := s.nodes[0]
	for i, n := range s.doc.Find(selector).nodes {
		if n == target {
			return i
		}
	}
	return -1
}

// IndexOfNode returns the 0-based position of node within the current
// selection, or -1 if it is not present. It corresponds to jQuery's
// .index(element).
func (s *Selection) IndexOfNode(node *Node) int {
	for i, n := range s.nodes {
		if n == node {
			return i
		}
	}
	return -1
}

// AddSelection returns a new Selection combining the receiver with the nodes of
// other, in document order with duplicates removed.
func (s *Selection) AddSelection(other *Selection) *Selection {
	combined := append(append([]*Node(nil), s.nodes...), other.nodes...)
	return s.pushStack(orderedUnique(s.doc, combined))
}

// AddNodes returns a new Selection combining the receiver with the given nodes,
// in document order with duplicates removed.
func (s *Selection) AddNodes(nodes ...*Node) *Selection {
	combined := append(append([]*Node(nil), s.nodes...), nodes...)
	return s.pushStack(orderedUnique(s.doc, combined))
}

// NextUntil returns the following element siblings of each selected node, up to
// (but not including) the first sibling that matches selector. An optional
// second selector filters the collected elements.
func (s *Selection) NextUntil(selector string, filter ...string) *Selection {
	stop := parseSelector(selector)
	var out []*Node
	for _, n := range s.nodes {
		for sib := nextElementSibling(n); sib != nil; sib = nextElementSibling(sib) {
			if stop.matches(sib) {
				break
			}
			out = append(out, sib)
		}
	}
	return s.pushStack(out).maybeFilter(filter)
}

// PrevUntil returns the preceding element siblings of each selected node, up to
// (but not including) the first sibling that matches selector. An optional
// second selector filters the collected elements.
func (s *Selection) PrevUntil(selector string, filter ...string) *Selection {
	stop := parseSelector(selector)
	var out []*Node
	for _, n := range s.nodes {
		for sib := previousElementSibling(n); sib != nil; sib = previousElementSibling(sib) {
			if stop.matches(sib) {
				break
			}
			out = append(out, sib)
		}
	}
	return s.pushStack(out).maybeFilter(filter)
}

// ParentsUntil returns the ancestors of each selected node, nearest first, up
// to (but not including) the first ancestor that matches selector. An optional
// second selector filters the collected elements.
func (s *Selection) ParentsUntil(selector string, filter ...string) *Selection {
	stop := parseSelector(selector)
	var out []*Node
	for _, n := range s.nodes {
		for a := n.Parent; a != nil && a.Type == ElementNode; a = a.Parent {
			if stop.matches(a) {
				break
			}
			out = append(out, a)
		}
	}
	return s.pushStack(out).maybeFilter(filter)
}

// Even returns the even-indexed nodes of the selection (0, 2, 4, ...).
func (s *Selection) Even() *Selection {
	var out []*Node
	for i, n := range s.nodes {
		if i%2 == 0 {
			out = append(out, n)
		}
	}
	return s.pushStack(out)
}

// Odd returns the odd-indexed nodes of the selection (1, 3, 5, ...).
func (s *Selection) Odd() *Selection {
	var out []*Node
	for i, n := range s.nodes {
		if i%2 == 1 {
			out = append(out, n)
		}
	}
	return s.pushStack(out)
}

// orderedUnique removes duplicate nodes and returns them sorted in document
// order relative to doc. Nodes not reachable from the document root retain a
// stable relative order after those that are.
func orderedUnique(doc *Document, nodes []*Node) []*Node {
	pos := make(map[*Node]int)
	idx := 0
	var walk func(n *Node)
	walk = func(n *Node) {
		pos[n] = idx
		idx++
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(doc.root)

	seen := make(map[*Node]bool, len(nodes))
	out := make([]*Node, 0, len(nodes))
	for _, n := range nodes {
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	// Stable insertion sort by document position (small selections).
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && docLess(pos, out[j], out[j-1]); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// docLess reports whether a precedes b in the position map. Unknown nodes sort
// after known ones.
func docLess(pos map[*Node]int, a, b *Node) bool {
	pa, oka := pos[a]
	pb, okb := pos[b]
	switch {
	case oka && okb:
		return pa < pb
	case oka:
		return true
	case okb:
		return false
	default:
		return false
	}
}
