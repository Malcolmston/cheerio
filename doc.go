// Package cheerio is a dependency-free, standard-library-only HTML parsing and
// jQuery-style traversal library for Go, inspired by the Node.js "cheerio"
// package.
//
// It provides three cooperating layers, all implemented from scratch without
// golang.org/x/net or any other third-party code:
//
//   - a hand-written HTML tokenizer and tree-construction parser;
//   - a CSS selector engine covering the selectors used in everyday scraping;
//   - a chainable, jQuery-style Selection API for querying and lightly
//     mutating the resulting DOM.
//
// # Parsing
//
// Load parses an HTML string into a Document. Parsing is tolerant: void
// elements (br, img, hr, input, meta, link, ...) never receive children,
// self-closing syntax is honored, comments and doctypes are preserved, and
// unclosed or improperly nested tags are recovered using implicit-close rules
// similar to a browser (for example an open <p> is closed when a block element
// or a new <p> begins, and <li>, <td>, <tr> and friends close open siblings).
// Raw-text elements (script, style, textarea, title) capture their content
// verbatim. HTML character references — named (&amp;), decimal (&#38;) and
// hexadecimal (&#x26;) — are decoded in text and attribute values.
//
// The parse tree is made of Node values. Each Node has a Type (DocumentNode,
// ElementNode, TextNode, CommentNode or DoctypeNode), a Data field (the
// lower-cased tag name for elements, otherwise the raw content), an ordered
// slice of Attribute values, and Parent/Children links.
//
// # Selectors
//
// The selector engine supports:
//
//   - type selectors (div), the universal selector (*), #id and .class;
//   - namespace-qualified type selectors (svg|rect, *|title), matched on the
//     local name;
//   - attribute selectors [attr], [attr=v], [attr^=v], [attr$=v], [attr*=v],
//     [attr~=v] and [attr|=v], with an optional case-insensitivity flag
//     ([attr=v i]);
//   - selector lists separated by commas;
//   - combinators: descendant (space), child (>), adjacent sibling (+) and
//     general sibling (~);
//   - structural pseudo-classes :first-child, :last-child, :only-child,
//     :first-of-type, :last-of-type, :only-of-type, :empty, :root, :parent,
//     :nth-child(n), :nth-last-child(n), :nth-of-type(n) and
//     :nth-last-of-type(n) (each with the an+b microsyntax plus the odd/even
//     keywords);
//   - functional pseudo-classes :not(selector), :has(selector) and
//     :contains(text);
//   - form and content pseudo-classes :input, :checked, :selected, :enabled,
//     :disabled and :header.
//
// Compound selectors combine any of the above (for example
// "ul.menu > li:first-child a[href^=\"/\"]").
//
// # Selection API
//
// Document.Find runs a selector over the whole document and returns a
// Selection: an ordered, de-duplicated set of matched nodes. Selections are
// chainable.
//
// Traversal: Find, Children, Contents, Parent, Parents, ParentsUntil, Closest,
// Siblings, Next, Prev, NextAll, PrevAll, NextUntil, PrevUntil, Filter,
// FilterFunc, Not, Is, Has, Add, AddSelection, AddNodes, AddBack, End, Eq,
// First, Last, Even, Odd, Slice, Index, IndexSelector, IndexOfNode, Each, Map,
// Get and Length.
//
// Accessors: Attr, AttrOr, Prop, Text, Html (inner HTML), OuterHtml, ToString,
// HasClass, Css, Data, DataOr, DataMap, Val and TagName.
//
// Manipulation: SetAttr, SetAttrs, RemoveAttr, AddClass, RemoveClass,
// ToggleClass, SetText, SetHtml, SetVal, SetProp, SetCss, SetCssMap, SetData,
// Append, Prepend, Before, After, AppendTo, PrependTo, InsertBefore,
// InsertAfter, ReplaceWith, Remove, Empty, Clone, Wrap, WrapAll, WrapInner and
// Unwrap. Each content method has string, node and Selection accepting forms.
//
// Forms: SerializeArray returns the successful form controls of a selection as
// name/value pairs, and Serialize encodes them as an
// application/x-www-form-urlencoded string.
//
// # Example
//
//	doc := cheerio.Load(`<ul><li class="a">One</li><li>Two</li></ul>`)
//	doc.Find("li.a").Each(func(i int, n *cheerio.Node) {
//		// ...
//	})
//	fmt.Println(doc.Find("li").First().Text()) // One
//
// The library is deterministic: identical input always yields identical
// output, and Selections preserve document order.
package cheerio
