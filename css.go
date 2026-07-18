package cheerio

import "strings"

// styleDecl is a single CSS declaration parsed from an inline style attribute.
type styleDecl struct {
	prop string
	val  string
}

// parseStyle parses the value of a style attribute into ordered declarations.
// Property names are lower-cased and trimmed; empty or malformed declarations
// are skipped. Order is preserved so serialization is deterministic.
func parseStyle(style string) []styleDecl {
	var out []styleDecl
	for _, part := range strings.Split(style, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		colon := strings.IndexByte(part, ':')
		if colon < 0 {
			// A fragment without a colon is a malformed continuation of the
			// previous declaration whose value contained a semicolon (for
			// example a data: URI). Reattach it to the previous value so the
			// declaration round-trips intact (cheeriojs/cheerio#1134).
			if len(out) > 0 {
				out[len(out)-1].val += ";" + part
			}
			continue
		}
		prop := strings.ToLower(strings.TrimSpace(part[:colon]))
		val := strings.TrimSpace(part[colon+1:])
		if prop == "" {
			continue
		}
		out = append(out, styleDecl{prop: prop, val: val})
	}
	return out
}

// serializeStyle renders declarations back into a style attribute value.
func serializeStyle(decls []styleDecl) string {
	var b strings.Builder
	for i, d := range decls {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(d.prop)
		b.WriteString(": ")
		b.WriteString(d.val)
		b.WriteString(";")
	}
	return b.String()
}

// setStyleProp sets or removes a single declaration in decls, preserving order,
// and returns the updated slice. An empty value removes the property.
func setStyleProp(decls []styleDecl, prop, val string) []styleDecl {
	prop = strings.ToLower(strings.TrimSpace(prop))
	for i := range decls {
		if decls[i].prop == prop {
			if val == "" {
				return append(decls[:i], decls[i+1:]...)
			}
			decls[i].val = val
			return decls
		}
	}
	if val == "" {
		return decls
	}
	return append(decls, styleDecl{prop: prop, val: val})
}

// Css returns the value of the named inline CSS property of the first selected
// element, or "" if the property (or the selection) is absent.
func (s *Selection) Css(prop string) string {
	if len(s.nodes) == 0 {
		return ""
	}
	style, _ := s.nodes[0].AttrValue("style")
	prop = strings.ToLower(strings.TrimSpace(prop))
	for _, d := range parseStyle(style) {
		if d.prop == prop {
			return d.val
		}
	}
	return ""
}

// SetCss sets a single inline CSS property on every selected element and
// returns the receiver. Setting an empty value removes the property.
func (s *Selection) SetCss(prop, val string) *Selection {
	for _, n := range s.nodes {
		if n.Type != ElementNode {
			continue
		}
		decls := parseStyle(attrOr(n, "style"))
		decls = setStyleProp(decls, prop, val)
		writeStyle(n, decls)
	}
	return s
}

// SetCssMap sets every property/value pair of styles on all selected elements
// and returns the receiver. Because Go maps have no order, callers that require
// a specific declaration order should use SetCss for each property.
func (s *Selection) SetCssMap(styles map[string]string) *Selection {
	for _, n := range s.nodes {
		if n.Type != ElementNode {
			continue
		}
		decls := parseStyle(attrOr(n, "style"))
		for prop, val := range styles {
			decls = setStyleProp(decls, prop, val)
		}
		writeStyle(n, decls)
	}
	return s
}

// writeStyle sets the style attribute of n from decls, removing the attribute
// entirely when there are no declarations left.
func writeStyle(n *Node, decls []styleDecl) {
	if len(decls) == 0 {
		n.removeAttr("style")
		return
	}
	n.setAttr("style", serializeStyle(decls))
}

// attrOr returns the value of the named attribute of n, or "".
func attrOr(n *Node, key string) string {
	v, _ := n.AttrValue(key)
	return v
}
