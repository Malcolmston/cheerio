package cheerio

import "strings"

// camelToDash converts a camelCase data key into its dashed attribute suffix,
// e.g. "fooBar" becomes "foo-bar". Existing dashes are preserved.
func camelToDash(key string) string {
	var b strings.Builder
	for i := 0; i < len(key); i++ {
		c := key[i]
		if c >= 'A' && c <= 'Z' {
			b.WriteByte('-')
			b.WriteByte(c - 'A' + 'a')
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}

// dashToCamel converts a dashed data attribute suffix into a camelCase key,
// e.g. "foo-bar" becomes "fooBar".
func dashToCamel(key string) string {
	var b strings.Builder
	upper := false
	for i := 0; i < len(key); i++ {
		c := key[i]
		if c == '-' {
			upper = true
			continue
		}
		if upper && c >= 'a' && c <= 'z' {
			b.WriteByte(c - 'a' + 'A')
		} else {
			b.WriteByte(c)
		}
		upper = false
	}
	return b.String()
}

// Data returns the value of the data-* attribute for key on the first selected
// element and whether it exists. The key may be given in camelCase ("fooBar")
// or dashed ("foo-bar") form; both map to the data-foo-bar attribute.
func (s *Selection) Data(key string) (string, bool) {
	if len(s.nodes) == 0 {
		return "", false
	}
	return s.nodes[0].AttrValue("data-" + camelToDash(key))
}

// DataOr returns the data-* value for key on the first selected element, or def
// if it is absent.
func (s *Selection) DataOr(key, def string) string {
	if v, ok := s.Data(key); ok {
		return v
	}
	return def
}

// SetData sets the data-* attribute for key on every selected element and
// returns the receiver. The key is normalized to dashed form.
func (s *Selection) SetData(key, val string) *Selection {
	attr := "data-" + camelToDash(key)
	for _, n := range s.nodes {
		if n.Type == ElementNode {
			n.setAttr(attr, val)
		}
	}
	return s
}

// DataMap returns all data-* attributes of the first selected element as a map
// keyed by camelCase name (data-foo-bar becomes "fooBar"). It returns an empty
// map when the selection is empty or has no data attributes.
func (s *Selection) DataMap() map[string]string {
	out := make(map[string]string)
	if len(s.nodes) == 0 {
		return out
	}
	for _, a := range s.nodes[0].Attr {
		if strings.HasPrefix(a.Key, "data-") {
			out[dashToCamel(strings.TrimPrefix(a.Key, "data-"))] = a.Val
		}
	}
	return out
}
