package cheerio

import (
	"errors"
	"fmt"
	"strings"
)

// Matcher is a compiled CSS selector. Compiling a selector once and reusing the
// resulting Matcher across many nodes or documents avoids the cost of parsing
// the selector string on every call, which matters in tight scraping loops.
//
// A Matcher is immutable after construction and safe for concurrent use by
// multiple goroutines.
type Matcher struct {
	source string
	list   selectorList
}

// Compile parses selector into a reusable Matcher. It returns an error if
// selector is empty or cannot be parsed. Unlike the string-accepting methods
// (which silently treat an unparseable selector as matching nothing), Compile
// reports malformed selectors so they can be caught at startup.
func Compile(selector string) (*Matcher, error) {
	list, err := compileSelector(selector)
	if err != nil {
		return nil, err
	}
	return &Matcher{source: selector, list: list}, nil
}

// MustCompile is like Compile but panics if selector cannot be parsed. It is
// intended for package-level Matcher variables initialized with known-good
// selectors.
func MustCompile(selector string) *Matcher {
	m, err := Compile(selector)
	if err != nil {
		panic(err)
	}
	return m
}

// Match reports whether n satisfies the compiled selector.
func (m *Matcher) Match(n *Node) bool {
	return m.list.matches(n)
}

// String returns the original selector text the Matcher was compiled from.
func (m *Matcher) String() string {
	return m.source
}

// FindMatcher returns all elements in the document that satisfy the compiled
// matcher. It is the pre-compiled equivalent of Find.
func (d *Document) FindMatcher(m *Matcher) *Selection {
	var out []*Node
	walkElements(d.root, func(n *Node) {
		if m.list.matches(n) {
			out = append(out, n)
		}
	})
	return &Selection{doc: d, nodes: out}
}

// FindMatcher returns the descendants of the selected nodes that satisfy the
// compiled matcher. It is the pre-compiled equivalent of Find.
func (s *Selection) FindMatcher(m *Matcher) *Selection {
	var out []*Node
	for _, n := range s.nodes {
		for _, c := range n.Children {
			walkElements(c, func(e *Node) {
				if m.list.matches(e) {
					out = append(out, e)
				}
			})
		}
	}
	return s.pushStack(out)
}

// FilterMatcher returns the subset of selected nodes that satisfy the compiled
// matcher. It is the pre-compiled equivalent of Filter.
func (s *Selection) FilterMatcher(m *Matcher) *Selection {
	var out []*Node
	for _, n := range s.nodes {
		if m.list.matches(n) {
			out = append(out, n)
		}
	}
	return s.pushStack(out)
}

// IsMatcher reports whether at least one selected node satisfies the compiled
// matcher. It is the pre-compiled equivalent of Is.
func (s *Selection) IsMatcher(m *Matcher) bool {
	for _, n := range s.nodes {
		if m.list.matches(n) {
			return true
		}
	}
	return false
}

// compileSelector parses a selector string strictly, returning an error when
// the input is empty or any comma-separated group fails to parse. It reuses the
// same grammar as parseSelector but reports failures instead of dropping them.
func compileSelector(s string) (selectorList, error) {
	if strings.TrimSpace(s) == "" {
		return nil, errors.New("cheerio: empty selector")
	}
	var list selectorList
	for _, group := range splitTopLevel(s, ',') {
		group = strings.TrimSpace(group)
		if group == "" {
			return nil, fmt.Errorf("cheerio: empty selector group in %q", s)
		}
		cs, ok := parseComplex(group)
		if !ok {
			return nil, fmt.Errorf("cheerio: invalid selector %q", group)
		}
		list = append(list, cs)
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("cheerio: invalid selector %q", s)
	}
	return list, nil
}
