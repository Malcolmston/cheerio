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
//   - attribute selectors [attr], [attr=v], [attr^=v], [attr$=v], [attr*=v],
//     [attr~=v] and [attr|=v];
//   - selector lists separated by commas;
//   - combinators: descendant (space), child (>), adjacent sibling (+) and
//     general sibling (~);
//   - pseudo-classes :first-child, :last-child, :only-child, :empty,
//     :nth-child(n) with the an+b microsyntax plus the odd/even keywords,
//     :nth-last-child(...) and :not(selector).
//
// Compound selectors combine any of the above (for example
// "ul.menu > li:first-child a[href^=\"/\"]").
//
// # Selection API
//
// Document.Find runs a selector over the whole document and returns a
// Selection: an ordered, de-duplicated set of matched nodes. Selections are
// chainable. Traversal methods include Find, Children, Parent, Parents,
// Closest, Siblings, Next, Prev, NextAll, PrevAll, Filter, FilterFunc, Not,
// Is, Has, Add, Eq, First, Last, Slice, Each, Map, Get and Length.
//
// Accessors include Attr/AttrOr, Text, Html (inner HTML), OuterHtml, HasClass,
// Val and TagName. Light mutation is supported through SetAttr, RemoveAttr,
// AddClass, RemoveClass, ToggleClass and SetText.
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
