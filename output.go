package cheerio

import (
	"io"
	"strings"
)

// Render writes the serialized outer HTML of every selected node, in document
// order, to w. It returns any error reported by w. Render is the streaming form
// of ToString, avoiding an intermediate string when writing to an
// http.ResponseWriter, a file, or any other io.Writer.
func (s *Selection) Render(w io.Writer) error {
	var b strings.Builder
	for _, n := range s.nodes {
		serialize(n, &b)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// ToArray returns a copy of the selected nodes as a slice. Unlike Nodes, the
// returned slice is safe to mutate; it mirrors jQuery's .toArray().
func (s *Selection) ToArray() []*Node {
	return append([]*Node(nil), s.nodes...)
}

// Attrs returns every attribute of the first selected element as a map keyed by
// attribute name. It returns an empty (non-nil) map when the selection is empty
// or the first node is not an element. This is the no-argument form of jQuery's
// .attr() that returns the whole attribute set.
func (s *Selection) Attrs() map[string]string {
	out := make(map[string]string)
	if len(s.nodes) == 0 {
		return out
	}
	for _, a := range s.nodes[0].Attr {
		out[a.Key] = a.Val
	}
	return out
}

// RemoveData removes data-* attributes from every selected element and returns
// the receiver. With no arguments it removes all data-* attributes; otherwise
// only the named keys are removed. Keys may be given in camelCase ("fooBar") or
// dashed ("foo-bar") form.
func (s *Selection) RemoveData(keys ...string) *Selection {
	for _, n := range s.nodes {
		if n.Type != ElementNode {
			continue
		}
		if len(keys) == 0 {
			kept := n.Attr[:0]
			for _, a := range n.Attr {
				if !strings.HasPrefix(a.Key, "data-") {
					kept = append(kept, a)
				}
			}
			n.Attr = kept
			continue
		}
		for _, k := range keys {
			n.removeAttr("data-" + camelToDash(k))
		}
	}
	return s
}

// MapNodes applies fn to each selected node in order and returns the collected
// non-nil results as a slice of nodes. It is the node-returning companion of
// Map (which collects strings), matching jQuery's .map() when the callback
// yields elements.
func (s *Selection) MapNodes(fn func(i int, n *Node) *Node) []*Node {
	out := make([]*Node, 0, len(s.nodes))
	for i, n := range s.nodes {
		if r := fn(i, n); r != nil {
			out = append(out, r)
		}
	}
	return out
}

// Html returns the serialized HTML of the whole document, including doctype,
// comments and top-level elements, in document order. It mirrors the static
// cheerio $.html() call on a loaded document.
func (d *Document) Html() string {
	var b strings.Builder
	for _, c := range d.root.Children {
		serialize(c, &b)
	}
	return b.String()
}

// Text returns the concatenated text content of the entire document, in
// document order. It mirrors the static cheerio $.text() call.
func (d *Document) Text() string {
	var b strings.Builder
	collectText(d.root, &b)
	return b.String()
}

// Render writes the document's serialized HTML (see Html) to w and returns any
// error reported by w.
func (d *Document) Render(w io.Writer) error {
	_, err := io.WriteString(w, d.Html())
	return err
}
