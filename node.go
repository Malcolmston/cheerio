package cheerio

import (
	"strings"
)

// NodeType identifies the kind of a Node in the DOM tree.
type NodeType int

const (
	// DocumentNode is the synthetic root node produced by Load.
	DocumentNode NodeType = iota
	// ElementNode is an HTML element such as <div> or <p>.
	ElementNode
	// TextNode is a run of character data.
	TextNode
	// CommentNode is an HTML comment (<!-- ... -->).
	CommentNode
	// DoctypeNode is a <!DOCTYPE ...> declaration.
	DoctypeNode
)

// Attribute is a single HTML attribute. Attributes preserve their source
// order within an element.
type Attribute struct {
	Key string
	Val string
}

// Node is a single node in the parsed DOM tree. Elements carry a lower-cased
// Data (tag name) and an ordered slice of attributes; text, comment and
// doctype nodes carry their content in Data.
type Node struct {
	Type     NodeType
	Data     string      // tag name for elements, text content for text/comment/doctype
	Attr     []Attribute // ordered attributes (elements only)
	Parent   *Node
	Children []*Node

	// dataCache holds values set via Selection.SetData. Like cheerio, .data(k,v)
	// stores the value in an in-memory cache (keyed by the dashed suffix, e.g.
	// "foo-bar") and does NOT write a data-* attribute to the DOM; the getter
	// overlays this cache on top of the element's data-* attributes.
	dataCache map[string]string
}

// voidElements are HTML elements that never have children and are closed
// implicitly by the parser.
var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "command": true,
	"embed": true, "hr": true, "img": true, "input": true, "keygen": true,
	"link": true, "meta": true, "param": true, "source": true, "track": true,
	"wbr": true,
}

// rawTextElements have content that is not parsed as markup.
var rawTextElements = map[string]bool{
	"script": true, "style": true, "textarea": true, "title": true,
}

// AppendChild adds c as the last child of n and sets c's parent.
func (n *Node) AppendChild(c *Node) {
	c.Parent = n
	n.Children = append(n.Children, c)
}

// elementChildren returns the element children of n in order.
func (n *Node) elementChildren() []*Node {
	var out []*Node
	for _, c := range n.Children {
		if c.Type == ElementNode {
			out = append(out, c)
		}
	}
	return out
}

// AttrValue returns the value of the named attribute and whether it exists.
// Attribute names are matched case-insensitively.
func (n *Node) AttrValue(key string) (string, bool) {
	key = strings.ToLower(key)
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val, true
		}
	}
	return "", false
}

// setAttr sets (or inserts) an attribute, preserving order for existing keys.
func (n *Node) setAttr(key, val string) {
	key = strings.ToLower(key)
	for i := range n.Attr {
		if n.Attr[i].Key == key {
			n.Attr[i].Val = val
			return
		}
	}
	n.Attr = append(n.Attr, Attribute{Key: key, Val: val})
}

// removeAttr deletes an attribute if present.
func (n *Node) removeAttr(key string) {
	key = strings.ToLower(key)
	out := n.Attr[:0]
	for _, a := range n.Attr {
		if a.Key != key {
			out = append(out, a)
		}
	}
	n.Attr = out
}

// classes returns the set of class tokens on the element.
func (n *Node) classes() []string {
	v, ok := n.AttrValue("class")
	if !ok {
		return nil
	}
	return strings.Fields(v)
}

// hasClass reports whether the element carries the given class token.
func (n *Node) hasClass(class string) bool {
	for _, c := range n.classes() {
		if c == class {
			return true
		}
	}
	return false
}
