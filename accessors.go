package cheerio

import (
	"strings"
)

// Attr returns the value of the named attribute of the first selected node and
// whether it exists.
func (s *Selection) Attr(name string) (string, bool) {
	if len(s.nodes) == 0 {
		return "", false
	}
	return s.nodes[0].AttrValue(name)
}

// AttrOr returns the value of the named attribute of the first node, or def if
// it is absent (or the selection is empty).
func (s *Selection) AttrOr(name, def string) string {
	if v, ok := s.Attr(name); ok {
		return v
	}
	return def
}

// SetAttr sets an attribute on every selected element and returns the receiver.
func (s *Selection) SetAttr(name, val string) *Selection {
	for _, n := range s.nodes {
		if n.Type == ElementNode {
			n.setAttr(name, val)
		}
	}
	return s
}

// RemoveAttr removes an attribute from every selected element.
func (s *Selection) RemoveAttr(name string) *Selection {
	for _, n := range s.nodes {
		if n.Type == ElementNode {
			n.removeAttr(name)
		}
	}
	return s
}

// HasClass reports whether any selected element carries the given class.
func (s *Selection) HasClass(class string) bool {
	for _, n := range s.nodes {
		if n.hasClass(class) {
			return true
		}
	}
	return false
}

// AddClass adds one or more space-separated classes to every selected element.
func (s *Selection) AddClass(classes string) *Selection {
	toAdd := strings.Fields(classes)
	for _, n := range s.nodes {
		if n.Type != ElementNode {
			continue
		}
		existing := n.classes()
		for _, c := range toAdd {
			if !containsStr(existing, c) {
				existing = append(existing, c)
			}
		}
		n.setAttr("class", strings.Join(existing, " "))
	}
	return s
}

// RemoveClass removes one or more space-separated classes from every selected
// element. Removing all classes leaves an empty class attribute.
func (s *Selection) RemoveClass(classes string) *Selection {
	toRemove := strings.Fields(classes)
	for _, n := range s.nodes {
		if n.Type != ElementNode {
			continue
		}
		var kept []string
		for _, c := range n.classes() {
			if !containsStr(toRemove, c) {
				kept = append(kept, c)
			}
		}
		n.setAttr("class", strings.Join(kept, " "))
	}
	return s
}

// ToggleClass toggles each space-separated class on every selected element.
func (s *Selection) ToggleClass(classes string) *Selection {
	toToggle := strings.Fields(classes)
	for _, n := range s.nodes {
		if n.Type != ElementNode {
			continue
		}
		existing := n.classes()
		for _, c := range toToggle {
			if containsStr(existing, c) {
				existing = removeStr(existing, c)
			} else {
				existing = append(existing, c)
			}
		}
		n.setAttr("class", strings.Join(existing, " "))
	}
	return s
}

// Text returns the concatenated text content of every selected node and its
// descendants, in document order.
func (s *Selection) Text() string {
	var b strings.Builder
	for _, n := range s.nodes {
		collectText(n, &b)
	}
	return b.String()
}

// SetText replaces the children of every selected element with a single text
// node.
func (s *Selection) SetText(text string) *Selection {
	for _, n := range s.nodes {
		if n.Type != ElementNode {
			continue
		}
		child := &Node{Type: TextNode, Data: text, Parent: n}
		n.Children = []*Node{child}
	}
	return s
}

// Html returns the inner HTML of the first selected node (its serialized
// children).
func (s *Selection) Html() string {
	if len(s.nodes) == 0 {
		return ""
	}
	var b strings.Builder
	for _, c := range s.nodes[0].Children {
		serialize(c, &b)
	}
	return b.String()
}

// OuterHtml returns the serialized HTML of the first selected node, including
// the node itself.
func (s *Selection) OuterHtml() string {
	if len(s.nodes) == 0 {
		return ""
	}
	var b strings.Builder
	serialize(s.nodes[0], &b)
	return b.String()
}

// Val returns the value of a form control: the "value" attribute for inputs,
// the selected option for selects, or the text content for textareas.
func (s *Selection) Val() string {
	if len(s.nodes) == 0 {
		return ""
	}
	n := s.nodes[0]
	switch n.Data {
	case "textarea":
		var b strings.Builder
		collectText(n, &b)
		return b.String()
	case "select":
		// First option with a selected attribute, else first option.
		var first *Node
		for _, opt := range n.elementChildren() {
			if opt.Data != "option" {
				continue
			}
			if first == nil {
				first = opt
			}
			if _, ok := opt.AttrValue("selected"); ok {
				return optionValue(opt)
			}
		}
		if first != nil {
			return optionValue(first)
		}
		return ""
	default:
		if v, ok := n.AttrValue("value"); ok {
			return v
		}
		return ""
	}
}

// TagName returns the tag name of the first selected element, or "".
func (s *Selection) TagName() string {
	if len(s.nodes) == 0 || s.nodes[0].Type != ElementNode {
		return ""
	}
	return s.nodes[0].Data
}

// optionValue returns an <option>'s value attribute or its text.
func optionValue(opt *Node) string {
	if v, ok := opt.AttrValue("value"); ok {
		return v
	}
	var b strings.Builder
	collectText(opt, &b)
	return strings.TrimSpace(b.String())
}

// collectText appends the text of n and its descendants to b.
func collectText(n *Node, b *strings.Builder) {
	switch n.Type {
	case TextNode:
		b.WriteString(n.Data)
	case ElementNode, DocumentNode:
		for _, c := range n.Children {
			collectText(c, b)
		}
	}
}

// serialize writes the HTML serialization of n to b.
func serialize(n *Node, b *strings.Builder) {
	switch n.Type {
	case TextNode:
		if n.Parent != nil && rawTextElements[n.Parent.Data] &&
			(n.Parent.Data == "script" || n.Parent.Data == "style") {
			b.WriteString(n.Data)
		} else {
			b.WriteString(escapeText(n.Data))
		}
	case CommentNode:
		b.WriteString("<!--")
		b.WriteString(n.Data)
		b.WriteString("-->")
	case DoctypeNode:
		b.WriteString("<!")
		b.WriteString(n.Data)
		b.WriteString(">")
	case ElementNode:
		b.WriteByte('<')
		b.WriteString(n.Data)
		for _, a := range n.Attr {
			b.WriteByte(' ')
			b.WriteString(a.Key)
			b.WriteString("=\"")
			b.WriteString(escapeAttr(a.Val))
			b.WriteByte('"')
		}
		b.WriteByte('>')
		if voidElements[n.Data] {
			return
		}
		for _, c := range n.Children {
			serialize(c, b)
		}
		b.WriteString("</")
		b.WriteString(n.Data)
		b.WriteByte('>')
	case DocumentNode:
		for _, c := range n.Children {
			serialize(c, b)
		}
	}
}

func containsStr(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func removeStr(list []string, v string) []string {
	out := list[:0]
	for _, s := range list {
		if s != v {
			out = append(out, s)
		}
	}
	return out
}
