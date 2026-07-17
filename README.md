# cheerio

HTML parsing and jQuery-style traversal for Go — a dependency-free,
standard-library-only library inspired by the Node.js
[cheerio](https://cheerio.js.org/) package.

It ships its own HTML tokenizer, tree-construction parser, and CSS selector
engine. There are **no third-party dependencies** (not even
`golang.org/x/net`) and no cgo.

## Install

```sh
go get github.com/malcolmston/cheerio
```

Requires Go 1.24 or newer.

## Quick start

```go
package main

import (
	"fmt"

	"github.com/malcolmston/cheerio"
)

func main() {
	doc := cheerio.Load(`
		<ul id="fruit">
			<li class="a">Apple</li>
			<li class="b">Banana</li>
			<li class="c">Cherry</li>
		</ul>
		<a href="https://example.com/go">Go</a>`)

	// Text of the first list item.
	fmt.Println(doc.Find("#fruit li").First().Text()) // Apple

	// Iterate over matched elements.
	doc.Find("li").Each(func(i int, n *cheerio.Node) {
		fmt.Printf("%d: %s\n", i, n.Children[0].Data)
	})

	// Attribute access.
	href, _ := doc.Find("a").Attr("href")
	fmt.Println(href) // https://example.com/go

	// Combinators and pseudo-classes.
	fmt.Println(doc.Find("li.b + li").Text())           // Cherry
	fmt.Println(doc.Find("li:nth-child(odd)").Length()) // 2
}
```

## Parsing

`Load` never fails. Malformed input is recovered the way a browser would:

- void elements (`br`, `img`, `hr`, `input`, `meta`, `link`, …) never take
  children;
- self-closing syntax (`<el/>`) is honored;
- comments and doctypes are preserved;
- unclosed and improperly nested tags are recovered with implicit-close rules
  (an open `<p>` is closed by a following block element or `<p>`; `<li>`,
  `<tr>`, `<td>`, `<option>` and friends close open siblings);
- raw-text elements (`script`, `style`, `textarea`, `title`) capture their
  content verbatim;
- HTML character references — named (`&amp;`), decimal (`&#38;`) and
  hexadecimal (`&#x26;`) — are decoded in text and attribute values.

The tree is made of `*Node` values (`DocumentNode`, `ElementNode`, `TextNode`,
`CommentNode`, `DoctypeNode`) with ordered `Attribute`s and `Parent`/`Children`
links.

## Selectors

| Category       | Supported                                                                                                                              |
|----------------|--------------------------------------------------------------------------------------------------------------------------------------|
| Simple         | `type`, `*`, `#id`, `.class`                                                                                                          |
| Attribute      | `[a]`, `[a=v]`, `[a^=v]`, `[a$=v]`, `[a*=v]`, `[a~=v]`, `[a\|=v]`                                                                     |
| Combinators    | descendant (space), child `>`, adjacent sibling `+`, general sibling `~`                                                              |
| Lists          | comma-separated groups `a, b, c`                                                                                                      |
| Pseudo-classes | `:first-child`, `:last-child`, `:only-child`, `:empty`, `:nth-child(n)` (with `an+b`, `odd`, `even`), `:nth-last-child(...)`, `:not(sel)` |

Compound selectors combine all of the above, e.g.
`ul.menu > li:first-child a[href^="/"]`.

## Selection API

`Document.Find` returns a chainable `*Selection` — an ordered, de-duplicated
set of matched nodes.

**Traversal:** `Find`, `Children`, `Parent`, `Parents`, `Closest`, `Siblings`,
`Next`, `Prev`, `NextAll`, `PrevAll`, `Filter`, `FilterFunc`, `Not`, `Is`,
`Has`, `Add`, `Eq`, `First`, `Last`, `Slice`, `Each`, `Map`, `Get`, `Length`,
`Nodes`, `Root`, `End`.

**Accessors:** `Attr`, `AttrOr`, `Text`, `Html` (inner), `OuterHtml`,
`HasClass`, `Val`, `TagName`.

**Mutation (light):** `SetAttr`, `RemoveAttr`, `AddClass`, `RemoveClass`,
`ToggleClass`, `SetText`.

## Determinism

Identical input always yields identical output, and Selections preserve
document order.

## License

MIT
