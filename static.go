package cheerio

import "strings"

// ParseHTML parses an HTML fragment and returns its top-level nodes, detached
// from any document root (each returned node has a nil Parent). It mirrors the
// static cheerio.parseHTML helper and is convenient for building content to
// hand to the manipulation methods (AppendNodes, BeforeNodes, ...).
//
// Parsing never fails: malformed input is recovered the way a browser would.
// An empty (or whitespace-only markup with no nodes) input returns nil, matching
// cheerio's $.parseHTML which returns null for an empty string.
func ParseHTML(html string) []*Node {
	if html == "" {
		return nil
	}
	return parseFragment(html)
}

// Contains reports whether contained is a descendant of container (the two must
// be distinct nodes). It mirrors the static cheerio.contains helper and the DOM
// Node.contains method restricted to strict descendants.
func Contains(container, contained *Node) bool {
	if container == nil || contained == nil || container == contained {
		return false
	}
	for p := contained.Parent; p != nil; p = p.Parent {
		if p == container {
			return true
		}
	}
	return false
}

// Merge appends the nodes of tail to head and returns the combined slice,
// mirroring the static cheerio.merge helper (which is jQuery's $.merge). Like
// upstream it concatenates without de-duplicating: a node present in both
// inputs appears twice. The input slices are not modified.
func Merge(head []*Node, tail ...*Node) []*Node {
	combined := make([]*Node, 0, len(head)+len(tail))
	combined = append(combined, head...)
	combined = append(combined, tail...)
	return combined
}

// RenderNodes returns the serialized outer HTML of every supplied node,
// concatenated in argument order. It mirrors the static cheerio.html helper
// applied to a set of nodes.
func RenderNodes(nodes ...*Node) string {
	var b strings.Builder
	for _, n := range nodes {
		serialize(n, &b)
	}
	return b.String()
}

// TextOf returns the concatenated text content of every supplied node and its
// descendants, in argument order. It mirrors the static cheerio.text helper.
func TextOf(nodes ...*Node) string {
	var b strings.Builder
	for _, n := range nodes {
		collectText(n, &b)
	}
	return b.String()
}
