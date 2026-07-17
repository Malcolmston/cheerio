// Library content for the cheerio documentation site. Mirrors the shape used by
// the malcolmston/go landing site's data.ts so the sibling sites stay in sync.
export interface Lib {
  id: string; name: string; icon: string; accent: string; pkg: string; node: string;
  repo: string; docs: string; tagline: string; blurb: string; tags: string[];
  features: string[]; node_code: string; go_code: string; integrate: string;
}

export const NODE_ACCENT = '#8cc84b';

export const CHEERIO: Lib = {
  id:"cheerio", name:"cheerio", icon:'<i class="fa-solid fa-sitemap"></i>', accent:"#ff7a45",
  pkg:"github.com/malcolmston/cheerio", node:"cheeriojs/cheerio",
  repo:"https://github.com/malcolmston/cheerio", docs:"https://malcolmston.github.io/cheerio/",
  tagline:"HTML parsing and jQuery-style traversal for Go.",
  blurb:"A dependency-free, standard-library-only Go take on the Node.js cheerio package: tolerant HTML "+
    "parsing plus a chainable, jQuery-style traversal API. Everything is built from scratch — its own HTML "+
    "tokenizer and tree-construction parser, a CSS selector engine, and the Selection API — with no cgo and "+
    "no third-party code, not even golang.org/x/net. Load never fails: malformed markup is recovered the way "+
    "a browser would into a *Node tree, then Find returns an ordered, de-duplicated Selection you can filter, "+
    "traverse, read and lightly mutate. The import path is github.com/malcolmston/cheerio and the package is "+
    "named cheerio.",
  tags:["Load()","CSS selectors","jQuery-style API","own tokenizer","tree builder",":nth-child","combinators","zero deps"],
  features:[
    "Tolerant parsing — <code>Load</code> recovers malformed HTML into a <code>*Node</code> tree (<code>DocumentNode</code>/<code>ElementNode</code>/<code>TextNode</code>/<code>CommentNode</code>/<code>DoctypeNode</code>) using its own tokenizer + tree builder, no <code>golang.org/x/net</code>",
    "CSS selector engine — type/<code>*</code>/<code>#id</code>/<code>.class</code>, attribute matches (<code>[a^=v]</code>, <code>[a$=v]</code>, <code>[a*=v]</code>, <code>[a~=v]</code>, <code>[a|=v]</code>), combinators <code>&gt;</code> <code>+</code> <code>~</code>, and pseudo-classes <code>:nth-child(an+b)</code>/<code>:not</code>",
    "Chainable traversal — <code>Find</code>, <code>Children</code>, <code>Parent</code>, <code>Parents</code>, <code>Closest</code>, <code>Siblings</code>, <code>Next</code>/<code>Prev</code>, <code>Filter</code>, <code>Not</code>, <code>Is</code>, <code>Has</code>, <code>Eq</code>, <code>First</code>/<code>Last</code>, <code>Each</code>, <code>Map</code>",
    "Accessors — <code>Attr</code>/<code>AttrOr</code>, <code>Text</code>, <code>Html</code> (inner), <code>OuterHtml</code>, <code>HasClass</code>, <code>Val</code>, <code>TagName</code>",
    "Light mutation — <code>SetAttr</code>, <code>RemoveAttr</code>, <code>AddClass</code>, <code>RemoveClass</code>, <code>ToggleClass</code>, <code>SetText</code>",
    "Browser-like recovery — void elements take no children, self-closing syntax is honored, implicit-close rules fire for <code>&lt;p&gt;</code>/<code>&lt;li&gt;</code>/<code>&lt;tr&gt;</code>, raw-text <code>&lt;script&gt;</code>/<code>&lt;style&gt;</code>/<code>&lt;textarea&gt;</code> capture verbatim, and named/decimal/hex character references are decoded",
    "Deterministic — identical input yields identical output and Selections preserve document order; <code>Val</code> resolves <code>&lt;input&gt;</code>/<code>&lt;select&gt;</code>/<code>&lt;textarea&gt;</code> values",
    "Zero dependencies — pure Go standard library, no cgo, nothing to audit but the toolchain"
  ],
  node_code:
`import * as cheerio from 'cheerio';

const $ = cheerio.load(\`
  <ul id="fruit">
    <li class="a">Apple</li>
    <li class="b">Banana</li>
  </ul>
  <a href="/go">Go</a>\`);

// Text of the first list item.
console.log($('#fruit li').first().text()); // Apple

// Iterate over matched elements.
$('li').each((i, el) => console.log(i, $(el).text()));

// Attribute access.
console.log($('a').attr('href')); // /go`,
  go_code:
`import "github.com/malcolmston/cheerio"

doc := cheerio.Load(\`
  <ul id="fruit">
    <li class="a">Apple</li>
    <li class="b">Banana</li>
  </ul>
  <a href="/go">Go</a>\`)

// Text of the first list item.
fmt.Println(doc.Find("#fruit li").First().Text()) // Apple

// Iterate over matched elements.
doc.Find("li").Each(func(i int, n *cheerio.Node) {
    fmt.Println(i, n.Children[0].Data)
})

// Attribute access plus a pseudo-class.
href, _ := doc.Find("a").Attr("href")
fmt.Println(href, doc.Find("li:nth-child(odd)").Length()) // /go 1`,
  integrate:
`<span class="tok-c">// Load never fails: malformed markup is recovered the way a browser would.</span>
doc := cheerio.Load(page)

<span class="tok-c">// Combinators + attribute selectors — collect every off-site nav link.</span>
var external []string
doc.Find("nav a[href^=\\"http\\"]").Each(func(i int, n *cheerio.Node) {
    if href, ok := n.AttrValue("href"); ok {
        external = append(external, href)
    }
})

<span class="tok-c">// Pseudo-classes + light mutation: stripe the odd, non-header table rows.</span>
doc.Find("table tr:nth-child(odd):not(.head)").AddClass("zebra")

<span class="tok-c">// Traverse up the tree, then read text and serialized HTML back out.</span>
price := doc.Find(".product .price").First()
card := price.Closest(".product")
fmt.Println(card.Find("h2").Text(), price.Text(), card.OuterHtml())`
};
