package cheerio

import (
	"net/url"
	"strings"
)

// FormField is a single successful form control name/value pair produced by
// SerializeArray.
type FormField struct {
	Name  string // the control's name attribute
	Value string // the control's submitted value
}

// SerializeArray collects the successful form controls within the selection
// and returns them as name/value pairs, following the HTML form-submission
// rules used by jQuery's .serializeArray():
//
//   - controls without a name, and disabled controls, are skipped;
//   - buttons and submit/reset/image/file inputs are skipped;
//   - checkboxes and radios contribute only when checked;
//   - a multiple-select contributes one entry per selected option;
//   - other selects contribute their current value.
//
// When a selected node is itself a form (or other container) its descendant
// controls are gathered; when it is a control it is considered directly.
func (s *Selection) SerializeArray() []FormField {
	var out []FormField
	seen := make(map[*Node]bool)
	for _, n := range s.nodes {
		collectFields(n, &out, seen)
	}
	return out
}

// Serialize returns the selection's successful form controls encoded as an
// application/x-www-form-urlencoded string (e.g. "a=1&b=2"), matching jQuery's
// .serialize().
func (s *Selection) Serialize() string {
	fields := s.SerializeArray()
	var b strings.Builder
	for i, f := range fields {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(url.QueryEscape(f.Name))
		b.WriteByte('=')
		b.WriteString(url.QueryEscape(f.Value))
	}
	return b.String()
}

// collectFields walks n and appends successful controls to out. If n is itself
// a control it is considered directly; otherwise its descendants are scanned.
func collectFields(n *Node, out *[]FormField, seen map[*Node]bool) {
	if n.Type != ElementNode {
		return
	}
	if isFormControl(n) {
		appendField(n, out, seen)
		return
	}
	walkElements(n, func(e *Node) {
		if isFormControl(e) {
			appendField(e, out, seen)
		}
	})
}

// isFormControl reports whether n is an input, select or textarea element.
func isFormControl(n *Node) bool {
	switch n.Data {
	case "input", "select", "textarea":
		return true
	}
	return false
}

// appendField adds n's successful value(s) to out, honoring the submission
// rules. Each control is processed at most once.
func appendField(n *Node, out *[]FormField, seen map[*Node]bool) {
	if seen[n] {
		return
	}
	seen[n] = true

	name, ok := n.AttrValue("name")
	if !ok || name == "" {
		return
	}
	if _, disabled := n.AttrValue("disabled"); disabled {
		return
	}

	switch n.Data {
	case "input":
		typ := strings.ToLower(attrOr(n, "type"))
		switch typ {
		case "submit", "reset", "button", "image", "file":
			return
		case "checkbox", "radio":
			if _, checked := n.AttrValue("checked"); !checked {
				return
			}
			val, has := n.AttrValue("value")
			if !has {
				val = "on"
			}
			*out = append(*out, FormField{Name: name, Value: val})
		default:
			*out = append(*out, FormField{Name: name, Value: attrOr(n, "value")})
		}
	case "textarea":
		var b strings.Builder
		collectText(n, &b)
		*out = append(*out, FormField{Name: name, Value: b.String()})
	case "select":
		_, multiple := n.AttrValue("multiple")
		var selected []*Node
		for _, opt := range n.elementChildren() {
			if opt.Data != "option" {
				continue
			}
			if _, ok := opt.AttrValue("selected"); ok {
				selected = append(selected, opt)
			}
		}
		if len(selected) == 0 && !multiple {
			// Default: first option is selected when none is marked.
			for _, opt := range n.elementChildren() {
				if opt.Data == "option" {
					selected = append(selected, opt)
					break
				}
			}
		}
		if multiple {
			for _, opt := range selected {
				*out = append(*out, FormField{Name: name, Value: optionValue(opt)})
			}
		} else if len(selected) > 0 {
			*out = append(*out, FormField{Name: name, Value: optionValue(selected[0])})
		}
	}
}
