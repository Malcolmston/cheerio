package cheerio

// ---------------------------------------------------------------------------
// Low-level node manipulation helpers
// ---------------------------------------------------------------------------

// removeNode detaches n from its parent, leaving n's own subtree intact.
func removeNode(n *Node) {
	p := n.Parent
	if p == nil {
		return
	}
	for i, c := range p.Children {
		if c == n {
			p.Children = append(p.Children[:i], p.Children[i+1:]...)
			break
		}
	}
	n.Parent = nil
}

// prependChild inserts c as the first child of n.
func (n *Node) prependChild(c *Node) {
	c.Parent = n
	n.Children = append([]*Node{c}, n.Children...)
}

// indexInParent returns the position of n within its parent's Children, or -1.
func indexInParent(n *Node) int {
	if n.Parent == nil {
		return -1
	}
	for i, c := range n.Parent.Children {
		if c == n {
			return i
		}
	}
	return -1
}

// insertBeforeNode inserts newNode immediately before ref among ref's siblings.
func insertBeforeNode(ref, newNode *Node) {
	p := ref.Parent
	if p == nil {
		return
	}
	i := indexInParent(ref)
	if i < 0 {
		return
	}
	newNode.Parent = p
	p.Children = append(p.Children[:i], append([]*Node{newNode}, p.Children[i:]...)...)
}

// insertAfterNode inserts newNode immediately after ref among ref's siblings.
func insertAfterNode(ref, newNode *Node) {
	p := ref.Parent
	if p == nil {
		return
	}
	i := indexInParent(ref)
	if i < 0 {
		return
	}
	newNode.Parent = p
	p.Children = append(p.Children[:i+1], append([]*Node{newNode}, p.Children[i+1:]...)...)
}

// cloneNode returns a deep copy of n with a nil parent.
func cloneNode(n *Node) *Node {
	c := &Node{Type: n.Type, Data: n.Data}
	if len(n.Attr) > 0 {
		c.Attr = make([]Attribute, len(n.Attr))
		copy(c.Attr, n.Attr)
	}
	for _, ch := range n.Children {
		cc := cloneNode(ch)
		cc.Parent = c
		c.Children = append(c.Children, cc)
	}
	return c
}

// cloneNodes returns deep copies of every node in nodes.
func cloneNodes(nodes []*Node) []*Node {
	out := make([]*Node, len(nodes))
	for i, n := range nodes {
		out[i] = cloneNode(n)
	}
	return out
}

// deepestElement descends the first element child chain of n and returns the
// innermost element (n itself if it has no element children). It is used to
// find the insertion point of a wrapper structure.
func deepestElement(n *Node) *Node {
	cur := n
	for {
		child := firstElementChild(cur)
		if child == nil {
			return cur
		}
		cur = child
	}
}

// firstElementChild returns the first element child of n, or nil.
func firstElementChild(n *Node) *Node {
	for _, c := range n.Children {
		if c.Type == ElementNode {
			return c
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Content generation
// ---------------------------------------------------------------------------

// parseFragment parses an HTML fragment string and returns its top-level nodes,
// detached from the synthetic document root.
func parseFragment(html string) []*Node {
	doc := parse(html)
	out := make([]*Node, 0, len(doc.Children))
	for _, c := range doc.Children {
		c.Parent = nil
		out = append(out, c)
	}
	return out
}

// parseFragmentFirstElement parses html and returns its first top-level element
// node, or nil if the fragment has none.
func parseFragmentFirstElement(html string) *Node {
	for _, n := range parseFragment(html) {
		if n.Type == ElementNode {
			return n
		}
	}
	return nil
}

// manipulateNodes inserts nodes into each target of the selection using place.
// For every target but the last the nodes are cloned, so shared content is
// duplicated rather than moved; the last target receives the originals (after
// detaching them from any current parent), mirroring jQuery move semantics.
// When reverse is true the nodes are placed in reverse order (used by prepend
// and after so the resulting document order matches the input order).
func (s *Selection) manipulateNodes(nodes []*Node, reverse bool, place func(target, node *Node)) *Selection {
	if len(nodes) == 0 {
		return s
	}
	last := len(s.nodes) - 1
	for i, target := range s.nodes {
		var batch []*Node
		if i == last {
			batch = nodes
		} else {
			batch = cloneNodes(nodes)
		}
		if reverse {
			for j := len(batch) - 1; j >= 0; j-- {
				removeNode(batch[j])
				place(target, batch[j])
			}
		} else {
			for _, n := range batch {
				removeNode(n)
				place(target, n)
			}
		}
	}
	return s
}

// ---------------------------------------------------------------------------
// Append / Prepend
// ---------------------------------------------------------------------------

// Append parses each HTML string and appends the resulting nodes to the end of
// the children of every selected element.
func (s *Selection) Append(html ...string) *Selection {
	return s.AppendNodes(parseFragment(joinHTML(html))...)
}

// AppendNodes appends the given nodes to the end of the children of every
// selected element.
func (s *Selection) AppendNodes(nodes ...*Node) *Selection {
	return s.manipulateNodes(nodes, false, func(target, node *Node) {
		target.AppendChild(node)
	})
}

// AppendSelection appends the nodes of other to the end of the children of
// every selected element.
func (s *Selection) AppendSelection(other *Selection) *Selection {
	return s.AppendNodes(other.nodes...)
}

// Prepend parses each HTML string and inserts the resulting nodes at the start
// of the children of every selected element.
func (s *Selection) Prepend(html ...string) *Selection {
	return s.PrependNodes(parseFragment(joinHTML(html))...)
}

// PrependNodes inserts the given nodes at the start of the children of every
// selected element.
func (s *Selection) PrependNodes(nodes ...*Node) *Selection {
	return s.manipulateNodes(nodes, true, func(target, node *Node) {
		target.prependChild(node)
	})
}

// PrependSelection inserts the nodes of other at the start of the children of
// every selected element.
func (s *Selection) PrependSelection(other *Selection) *Selection {
	return s.PrependNodes(other.nodes...)
}

// ---------------------------------------------------------------------------
// Before / After
// ---------------------------------------------------------------------------

// Before parses each HTML string and inserts the resulting nodes before every
// selected node in the tree.
func (s *Selection) Before(html ...string) *Selection {
	return s.BeforeNodes(parseFragment(joinHTML(html))...)
}

// BeforeNodes inserts the given nodes before every selected node.
func (s *Selection) BeforeNodes(nodes ...*Node) *Selection {
	return s.manipulateNodes(nodes, false, func(target, node *Node) {
		insertBeforeNode(target, node)
	})
}

// BeforeSelection inserts the nodes of other before every selected node.
func (s *Selection) BeforeSelection(other *Selection) *Selection {
	return s.BeforeNodes(other.nodes...)
}

// After parses each HTML string and inserts the resulting nodes after every
// selected node in the tree.
func (s *Selection) After(html ...string) *Selection {
	return s.AfterNodes(parseFragment(joinHTML(html))...)
}

// AfterNodes inserts the given nodes after every selected node.
func (s *Selection) AfterNodes(nodes ...*Node) *Selection {
	return s.manipulateNodes(nodes, true, func(target, node *Node) {
		insertAfterNode(target, node)
	})
}

// AfterSelection inserts the nodes of other after every selected node.
func (s *Selection) AfterSelection(other *Selection) *Selection {
	return s.AfterNodes(other.nodes...)
}

// ---------------------------------------------------------------------------
// Inverse insertion (AppendTo / PrependTo / InsertBefore / InsertAfter)
// ---------------------------------------------------------------------------

// AppendTo appends the selected nodes to every element matching selector and
// returns the receiver.
func (s *Selection) AppendTo(selector string) *Selection {
	s.doc.Find(selector).AppendNodes(s.nodes...)
	return s
}

// PrependTo prepends the selected nodes to every element matching selector and
// returns the receiver.
func (s *Selection) PrependTo(selector string) *Selection {
	s.doc.Find(selector).PrependNodes(s.nodes...)
	return s
}

// InsertBefore inserts the selected nodes before every element matching
// selector and returns the receiver.
func (s *Selection) InsertBefore(selector string) *Selection {
	s.doc.Find(selector).BeforeNodes(s.nodes...)
	return s
}

// InsertAfter inserts the selected nodes after every element matching selector
// and returns the receiver.
func (s *Selection) InsertAfter(selector string) *Selection {
	s.doc.Find(selector).AfterNodes(s.nodes...)
	return s
}

// ---------------------------------------------------------------------------
// Replace / Remove / Empty / Clone
// ---------------------------------------------------------------------------

// ReplaceWith replaces every selected node with the nodes parsed from html and
// returns the receiver (whose nodes are now detached).
func (s *Selection) ReplaceWith(html ...string) *Selection {
	return s.ReplaceWithNodes(parseFragment(joinHTML(html))...)
}

// ReplaceWithSelection replaces every selected node with the nodes of other.
func (s *Selection) ReplaceWithSelection(other *Selection) *Selection {
	return s.ReplaceWithNodes(other.nodes...)
}

// ReplaceWithNodes replaces every selected node with the given nodes. For every
// target but the last the replacement nodes are cloned.
func (s *Selection) ReplaceWithNodes(nodes ...*Node) *Selection {
	last := len(s.nodes) - 1
	for i, target := range s.nodes {
		if target.Parent == nil {
			continue
		}
		var batch []*Node
		if i == last {
			batch = nodes
		} else {
			batch = cloneNodes(nodes)
		}
		for _, n := range batch {
			removeNode(n)
			insertBeforeNode(target, n)
		}
		removeNode(target)
	}
	return s
}

// Remove detaches every selected node (optionally filtered by selector) from
// the document and returns the receiver.
func (s *Selection) Remove(selector ...string) *Selection {
	target := s
	if len(selector) > 0 && selector[0] != "" {
		target = s.Filter(selector[0])
	}
	for _, n := range target.nodes {
		removeNode(n)
	}
	return s
}

// Empty removes all child nodes from every selected element and returns the
// receiver.
func (s *Selection) Empty() *Selection {
	for _, n := range s.nodes {
		if n.Type != ElementNode {
			continue
		}
		for _, c := range n.Children {
			c.Parent = nil
		}
		n.Children = nil
	}
	return s
}

// Clone returns a new Selection containing deep copies of the selected nodes.
// The clones are detached from the document.
func (s *Selection) Clone() *Selection {
	return s.pushStack(cloneNodes(s.nodes))
}

// ---------------------------------------------------------------------------
// Wrapping
// ---------------------------------------------------------------------------

// Wrap wraps every selected node in a clone of the structure described by html.
// Each node is placed inside the innermost descendant of its wrapper.
func (s *Selection) Wrap(html string) *Selection {
	for _, n := range s.nodes {
		if n.Parent == nil {
			continue
		}
		wrapper := parseFragmentFirstElement(html)
		if wrapper == nil {
			continue
		}
		inner := deepestElement(wrapper)
		insertBeforeNode(n, wrapper)
		removeNode(n)
		inner.AppendChild(n)
	}
	return s
}

// WrapAll wraps all selected nodes together in a single structure described by
// html, placed at the position of the first node.
func (s *Selection) WrapAll(html string) *Selection {
	if len(s.nodes) == 0 {
		return s
	}
	first := s.nodes[0]
	if first.Parent == nil {
		return s
	}
	wrapper := parseFragmentFirstElement(html)
	if wrapper == nil {
		return s
	}
	inner := deepestElement(wrapper)
	insertBeforeNode(first, wrapper)
	for _, n := range s.nodes {
		removeNode(n)
		inner.AppendChild(n)
	}
	return s
}

// WrapInner wraps the contents (child nodes) of every selected element in a
// clone of the structure described by html.
func (s *Selection) WrapInner(html string) *Selection {
	for _, n := range s.nodes {
		if n.Type != ElementNode {
			continue
		}
		wrapper := parseFragmentFirstElement(html)
		if wrapper == nil {
			continue
		}
		inner := deepestElement(wrapper)
		children := n.Children
		n.Children = nil
		n.AppendChild(wrapper)
		for _, c := range children {
			c.Parent = inner
			inner.Children = append(inner.Children, c)
		}
	}
	return s
}

// Unwrap removes the parent of every selected node, promoting the node (and its
// siblings) into the grandparent. Each distinct parent is unwrapped once.
func (s *Selection) Unwrap() *Selection {
	seen := make(map[*Node]bool)
	for _, n := range s.nodes {
		p := n.Parent
		if p == nil || p.Type != ElementNode || seen[p] {
			continue
		}
		gp := p.Parent
		if gp == nil {
			continue
		}
		seen[p] = true
		idx := indexInParent(p)
		if idx < 0 {
			continue
		}
		kids := p.Children
		for _, k := range kids {
			k.Parent = gp
		}
		tail := append([]*Node(nil), gp.Children[idx+1:]...)
		gp.Children = append(gp.Children[:idx], kids...)
		gp.Children = append(gp.Children, tail...)
		p.Parent = nil
		p.Children = nil
	}
	return s
}

// joinHTML concatenates HTML fragments so a variadic string argument behaves
// like a single fragment.
func joinHTML(parts []string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	default:
		total := 0
		for _, p := range parts {
			total += len(p)
		}
		b := make([]byte, 0, total)
		for _, p := range parts {
			b = append(b, p...)
		}
		return string(b)
	}
}
