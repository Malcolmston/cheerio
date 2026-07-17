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
| Simple         | `type`, `*`, `#id`, `.class`, namespace `svg\|rect`, `*\|title`                                                                       |
| Attribute      | `[a]`, `[a=v]`, `[a^=v]`, `[a$=v]`, `[a*=v]`, `[a~=v]`, `[a\|=v]`, case-insensitive `[a=v i]`                                          |
| Combinators    | descendant (space), child `>`, adjacent sibling `+`, general sibling `~`                                                              |
| Lists          | comma-separated groups `a, b, c`                                                                                                      |
| Structural     | `:first-child`, `:last-child`, `:only-child`, `:first-of-type`, `:last-of-type`, `:only-of-type`, `:empty`, `:root`, `:parent`, `:nth-child(n)`, `:nth-last-child(n)`, `:nth-of-type(n)`, `:nth-last-of-type(n)` (`an+b`, `odd`, `even`) |
| Functional     | `:not(sel)`, `:has(sel)`, `:contains(text)`                                                                                           |
| Forms/content  | `:input`, `:checked`, `:selected`, `:enabled`, `:disabled`, `:header`                                                                |

Compound selectors combine all of the above, e.g.
`ul.menu > li:first-child a[href^="/"]`.

## Selection API

`Document.Find` returns a chainable `*Selection` — an ordered, de-duplicated
set of matched nodes.

**Traversal:** `Find`, `Children`, `Contents`, `Parent`, `Parents`,
`ParentsUntil`, `Closest`, `Siblings`, `Next`, `Prev`, `NextAll`, `PrevAll`,
`NextUntil`, `PrevUntil`, `Filter`, `FilterFunc`, `Not`, `Is`, `Has`, `Add`,
`AddSelection`, `AddNodes`, `AddBack`, `End`, `Eq`, `First`, `Last`, `Even`,
`Odd`, `Slice`, `Index`, `IndexSelector`, `IndexOfNode`, `Each`, `Map`, `Get`,
`Length`, `Nodes`, `Root`.

**Accessors:** `Attr`, `AttrOr`, `Prop`, `Text`, `Html` (inner), `OuterHtml`,
`ToString`, `HasClass`, `Css`, `Data`, `DataOr`, `DataMap`, `Val`, `TagName`.

**Manipulation:** `SetAttr`, `SetAttrs`, `RemoveAttr`, `AddClass`,
`RemoveClass`, `ToggleClass`, `SetText`, `SetHtml`, `SetVal`, `SetProp`,
`SetCss`, `SetCssMap`, `SetData`, `Append`, `Prepend`, `Before`, `After`,
`AppendTo`, `PrependTo`, `InsertBefore`, `InsertAfter`, `ReplaceWith`,
`Remove`, `Empty`, `Clone`, `Wrap`, `WrapAll`, `WrapInner`, `Unwrap`. Each
content method also has `…Nodes` and `…Selection` forms.

**Forms:** `SerializeArray`, `Serialize`.

## Determinism

Identical input always yields identical output, and Selections preserve
document order.

## License

MIT
