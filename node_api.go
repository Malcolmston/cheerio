package cheerio

import (
	"io"
	"strings"
)

// Text returns the concatenated character data of n and all of its
// descendants, in document order. It is the node-level counterpart of
// Selection.Text and mirrors the DOM textContent property.
func (n *Node) Text() string {
	var b strings.Builder
	collectText(n, &b)
	return b.String()
}

// InnerHTML returns the serialized HTML of n's children (everything inside the
// element, excluding its own start and end tags). It corresponds to the DOM
// innerHTML property and to Selection.Html.
func (n *Node) InnerHTML() string {
	var b strings.Builder
	for _, c := range n.Children {
		serialize(c, &b)
	}
	return b.String()
}

// OuterHTML returns the serialized HTML of n including its own tags. It
// corresponds to the DOM outerHTML property and to Selection.OuterHtml.
func (n *Node) OuterHTML() string {
	var b strings.Builder
	serialize(n, &b)
	return b.String()
}

// Render writes the serialized outer HTML of n to w. It returns any error
// reported by w. This is the streaming form of OuterHTML, useful for writing a
// subtree directly to an http.ResponseWriter or file.
func (n *Node) Render(w io.Writer) error {
	var b strings.Builder
	serialize(n, &b)
	_, err := io.WriteString(w, b.String())
	return err
}

// Clone returns a deep copy of n. The copy has a nil Parent and fully
// independent Children and Attr slices, so mutating either tree leaves the
// other unchanged.
func (n *Node) Clone() *Node {
	return cloneNode(n)
}

// FirstChild returns the first child of n (of any node type), or nil if n has
// no children.
func (n *Node) FirstChild() *Node {
	if len(n.Children) == 0 {
		return nil
	}
	return n.Children[0]
}

// LastChild returns the last child of n (of any node type), or nil if n has no
// children.
func (n *Node) LastChild() *Node {
	if len(n.Children) == 0 {
		return nil
	}
	return n.Children[len(n.Children)-1]
}

// NextSibling returns the sibling immediately following n in its parent's child
// list (of any node type, including text and comment nodes), or nil.
func (n *Node) NextSibling() *Node {
	if n.Parent == nil {
		return nil
	}
	i := indexInParent(n)
	if i < 0 || i+1 >= len(n.Parent.Children) {
		return nil
	}
	return n.Parent.Children[i+1]
}

// PrevSibling returns the sibling immediately preceding n in its parent's child
// list (of any node type, including text and comment nodes), or nil.
func (n *Node) PrevSibling() *Node {
	if n.Parent == nil {
		return nil
	}
	i := indexInParent(n)
	if i <= 0 {
		return nil
	}
	return n.Parent.Children[i-1]
}

// NextElementSibling returns the nearest following sibling of n that is an
// element, skipping text and comment nodes, or nil.
func (n *Node) NextElementSibling() *Node {
	return nextElementSibling(n)
}

// PrevElementSibling returns the nearest preceding sibling of n that is an
// element, skipping text and comment nodes, or nil.
func (n *Node) PrevElementSibling() *Node {
	return previousElementSibling(n)
}

// HasClass reports whether n is an element carrying the given class token in
// its class attribute.
func (n *Node) HasClass(class string) bool {
	return n.hasClass(class)
}

// TagName returns the lower-cased tag name of n if it is an element, otherwise
// the empty string.
func (n *Node) TagName() string {
	if n.Type != ElementNode {
		return ""
	}
	return n.Data
}

// SetAttr sets (or inserts) the named attribute on n, preserving the source
// order of existing attributes. The name is lower-cased. It has no effect on
// non-element nodes.
func (n *Node) SetAttr(key, val string) {
	if n.Type != ElementNode {
		return
	}
	n.setAttr(key, val)
}

// RemoveAttr deletes the named attribute from n if present. The name is matched
// case-insensitively.
func (n *Node) RemoveAttr(key string) {
	if n.Type != ElementNode {
		return
	}
	n.removeAttr(key)
}

// Matches reports whether n satisfies the given CSS selector. It compiles the
// selector on each call; use Compile with Matcher.Match to reuse a selector
// across many nodes.
func (n *Node) Matches(selector string) bool {
	return parseSelector(selector).matches(n)
}
