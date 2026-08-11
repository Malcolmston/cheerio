package cheerio

import (
	"reflect"
	"strings"
	"testing"
)

const sample = `<!DOCTYPE html>
<html>
<head><title>Test &amp; Demo</title></head>
<body>
  <div id="main" class="container">
    <h1>Heading</h1>
    <p class="intro first">Hello <b>world</b>&hellip;</p>
    <p class="intro">Second paragraph</p>
    <ul class="list">
      <li class="item odd" data-idx="1">One</li>
      <li class="item even" data-idx="2">Two</li>
      <li class="item odd" data-idx="3">Three</li>
      <li class="item even last" data-idx="4">Four</li>
    </ul>
    <a href="https://example.com/path" title="ext" rel="noopener external">Link</a>
    <img src="pic.png" alt="pic">
    <input type="text" name="q" value="query">
    <!-- a comment -->
  </div>
</body>
</html>`

func load(t *testing.T) *Document {
	t.Helper()
	return Load(sample)
}

func TestFindType(t *testing.T) {
	doc := load(t)
	if got := doc.Find("li").Length(); got != 4 {
		t.Fatalf("li count = %d, want 4", got)
	}
	if got := doc.Find("p").Length(); got != 2 {
		t.Fatalf("p count = %d, want 2", got)
	}
}

func TestFindUniversal(t *testing.T) {
	doc := load(t)
	if n := doc.Find("*").Length(); n < 10 {
		t.Fatalf("universal selected %d elements, want >=10", n)
	}
}

func TestFindID(t *testing.T) {
	doc := load(t)
	s := doc.Find("#main")
	if s.Length() != 1 {
		t.Fatalf("#main length = %d", s.Length())
	}
	if s.TagName() != "div" {
		t.Fatalf("#main tag = %q", s.TagName())
	}
}

func TestFindClass(t *testing.T) {
	doc := load(t)
	if n := doc.Find(".intro").Length(); n != 2 {
		t.Fatalf(".intro = %d, want 2", n)
	}
	if n := doc.Find(".item.odd").Length(); n != 2 {
		t.Fatalf(".item.odd = %d, want 2", n)
	}
}

func TestAttrSelectors(t *testing.T) {
	doc := load(t)
	cases := []struct {
		sel  string
		want int
	}{
		{"[data-idx]", 4},
		{`[data-idx="2"]`, 1},
		{`[href^="https"]`, 1},
		{`[href$="/path"]`, 1},
		{`[href*="example"]`, 1},
		{`[rel~="external"]`, 1},
		{`[type="text"]`, 1},
		{`[href|="https"]`, 0},
	}
	for _, c := range cases {
		if got := doc.Find(c.sel).Length(); got != c.want {
			t.Errorf("%s = %d, want %d", c.sel, got, c.want)
		}
	}
}

func TestAttrPipeEqual(t *testing.T) {
	doc := Load(`<a hreflang="en-US"></a><a hreflang="en"></a><a hreflang="fr"></a>`)
	if n := doc.Find(`[hreflang|="en"]`).Length(); n != 2 {
		t.Fatalf(`[hreflang|="en"] = %d, want 2`, n)
	}
}

func TestCombinators(t *testing.T) {
	doc := load(t)
	cases := []struct {
		sel  string
		want int
	}{
		{"ul li", 4},
		{"ul > li", 4},
		{"div p", 2},
		{"h1 + p", 1},
		{"p ~ ul", 1},
		{"li + li", 3},
		{"li ~ li", 3},
		{"#main > ul > li.odd", 2},
	}
	for _, c := range cases {
		if got := doc.Find(c.sel).Length(); got != c.want {
			t.Errorf("%s = %d, want %d", c.sel, got, c.want)
		}
	}
}

func TestCommaList(t *testing.T) {
	doc := load(t)
	if n := doc.Find("h1, img, input").Length(); n != 3 {
		t.Fatalf("comma list = %d, want 3", n)
	}
}

func TestPseudoClasses(t *testing.T) {
	doc := load(t)
	cases := []struct {
		sel  string
		want int
	}{
		{"li:first-child", 1},
		{"li:last-child", 1},
		{"li:nth-child(2)", 1},
		{"li:nth-child(odd)", 2},
		{"li:nth-child(even)", 2},
		{"li:nth-child(2n)", 2},
		{"li:nth-child(2n+1)", 2},
		{"li:not(.odd)", 2},
		{"li:nth-last-child(1)", 1},
		{"p:first-child", 0},
	}
	for _, c := range cases {
		if got := doc.Find(c.sel).Length(); got != c.want {
			t.Errorf("%s = %d, want %d", c.sel, got, c.want)
		}
	}
}

func TestNthChildText(t *testing.T) {
	doc := load(t)
	if got := doc.Find("li:nth-child(3)").Text(); got != "Three" {
		t.Fatalf("nth-child(3) text = %q", got)
	}
}

func TestNotWithCompound(t *testing.T) {
	doc := load(t)
	if got := doc.Find("li:not(.last)").Length(); got != 3 {
		t.Fatalf("li:not(.last) = %d, want 3", got)
	}
}

func TestEmptyPseudo(t *testing.T) {
	// Per CSS/css-select, a whitespace-only element is NOT :empty (its text node
	// counts as content), so only the truly-empty span matches.
	doc := Load(`<div><span></span><span>x</span><span>  </span></div>`)
	if got := doc.Find("span:empty").Length(); got != 1 {
		t.Fatalf("span:empty = %d, want 1", got)
	}
}

func TestOnlyChild(t *testing.T) {
	doc := Load(`<div><p>solo</p></div><div><p>a</p><p>b</p></div>`)
	if got := doc.Find("p:only-child").Length(); got != 1 {
		t.Fatalf("p:only-child = %d, want 1", got)
	}
}

func TestText(t *testing.T) {
	doc := load(t)
	if got := doc.Find(".first").Text(); got != "Hello world…" {
		t.Fatalf("text = %q", got)
	}
}

func TestTitleEntityDecode(t *testing.T) {
	doc := load(t)
	if got := doc.Find("title").Text(); got != "Test & Demo" {
		t.Fatalf("title = %q", got)
	}
}

func TestAttrAccessors(t *testing.T) {
	doc := load(t)
	a := doc.Find("a")
	if v, ok := a.Attr("href"); !ok || v != "https://example.com/path" {
		t.Fatalf("href = %q ok=%v", v, ok)
	}
	if _, ok := a.Attr("nope"); ok {
		t.Fatal("unexpected attr")
	}
	if v := a.AttrOr("missing", "def"); v != "def" {
		t.Fatalf("AttrOr = %q", v)
	}
	if v := a.AttrOr("title", "def"); v != "ext" {
		t.Fatalf("AttrOr existing = %q", v)
	}
}

func TestVal(t *testing.T) {
	doc := load(t)
	if got := doc.Find("input").Val(); got != "query" {
		t.Fatalf("input val = %q", got)
	}
	doc2 := Load(`<textarea>line</textarea>`)
	if got := doc2.Find("textarea").Val(); got != "line" {
		t.Fatalf("textarea val = %q", got)
	}
	doc3 := Load(`<select><option value="a">A</option><option value="b" selected>B</option></select>`)
	if got := doc3.Find("select").Val(); got != "b" {
		t.Fatalf("select val = %q", got)
	}
	doc4 := Load(`<select><option>first</option><option>second</option></select>`)
	if got := doc4.Find("select").Val(); got != "first" {
		t.Fatalf("select default val = %q", got)
	}
}

func TestHtmlAndOuter(t *testing.T) {
	doc := Load(`<div class="x"><b>hi</b> there</div>`)
	d := doc.Find("div")
	if got := d.Html(); got != "<b>hi</b> there" {
		t.Fatalf("inner html = %q", got)
	}
	if got := d.OuterHtml(); got != `<div class="x"><b>hi</b> there</div>` {
		t.Fatalf("outer html = %q", got)
	}
}

func TestSerializeVoidAndComment(t *testing.T) {
	doc := Load(`<p>hi<br>bye<!--c--><img src="a.png"></p>`)
	got := doc.Find("p").OuterHtml()
	want := `<p>hi<br>bye<!--c--><img src="a.png"></p>`
	if got != want {
		t.Fatalf("serialize = %q want %q", got, want)
	}
}

func TestClassMutation(t *testing.T) {
	doc := Load(`<div class="a b"></div>`)
	d := doc.Find("div")
	if !d.HasClass("a") {
		t.Fatal("expected class a")
	}
	d.AddClass("c a")
	if !d.HasClass("c") {
		t.Fatal("AddClass failed")
	}
	if got := d.AttrOr("class", ""); got != "a b c" {
		t.Fatalf("after AddClass class = %q", got)
	}
	d.RemoveClass("b")
	if d.HasClass("b") {
		t.Fatal("RemoveClass failed")
	}
	d.ToggleClass("a z") // remove a, add z
	if d.HasClass("a") || !d.HasClass("z") {
		t.Fatalf("ToggleClass failed: %q", d.AttrOr("class", ""))
	}
}

func TestSetAttrAndText(t *testing.T) {
	doc := Load(`<a href="old">x</a>`)
	a := doc.Find("a")
	a.SetAttr("href", "new").SetText("link")
	if got := a.AttrOr("href", ""); got != "new" {
		t.Fatalf("SetAttr = %q", got)
	}
	if got := a.Text(); got != "link" {
		t.Fatalf("SetText = %q", got)
	}
	a.RemoveAttr("href")
	if _, ok := a.Attr("href"); ok {
		t.Fatal("RemoveAttr failed")
	}
}

func TestTraversal(t *testing.T) {
	doc := load(t)

	// Children
	if n := doc.Find("ul").Children().Length(); n != 4 {
		t.Fatalf("ul children = %d", n)
	}
	if n := doc.Find("ul").Children(".odd").Length(); n != 2 {
		t.Fatalf("ul children.odd = %d", n)
	}

	// Parent
	if got := doc.Find("li").First().Parent().TagName(); got != "ul" {
		t.Fatalf("li parent = %q", got)
	}

	// Parents
	parents := doc.Find(".first b").Parents().Map(func(i int, n *Node) string { return n.Data })
	want := []string{"p", "div", "body", "html"}
	if !reflect.DeepEqual(parents, want) {
		t.Fatalf("parents = %v want %v", parents, want)
	}

	// Closest
	if got := doc.Find("li.last").Closest("ul").TagName(); got != "ul" {
		t.Fatalf("closest ul = %q", got)
	}
	if got := doc.Find("li.last").Closest(".item").TagName(); got != "li" {
		t.Fatalf("closest self = %q", got)
	}

	// Siblings
	if n := doc.Find("li.even.last").Siblings().Length(); n != 3 {
		t.Fatalf("siblings = %d", n)
	}

	// Next / Prev
	if got := doc.Find("li").First().Next().Text(); got != "Two" {
		t.Fatalf("next = %q", got)
	}
	if got := doc.Find("li.last").Prev().Text(); got != "Three" {
		t.Fatalf("prev = %q", got)
	}

	// NextAll / PrevAll
	if n := doc.Find("li").First().NextAll().Length(); n != 3 {
		t.Fatalf("nextall = %d", n)
	}
	if n := doc.Find("li.last").PrevAll().Length(); n != 3 {
		t.Fatalf("prevall = %d", n)
	}
}

func TestFilterNotIsHas(t *testing.T) {
	doc := load(t)
	if n := doc.Find("li").Filter(".odd").Length(); n != 2 {
		t.Fatalf("filter = %d", n)
	}
	if n := doc.Find("li").Not(".odd").Length(); n != 2 {
		t.Fatalf("not = %d", n)
	}
	if !doc.Find("li").Is(".last") {
		t.Fatal("Is failed")
	}
	if doc.Find("li").Is(".nonexistent") {
		t.Fatal("Is false positive")
	}
	if n := doc.Find("ul").Has("li.last").Length(); n != 1 {
		t.Fatalf("has = %d", n)
	}
	if n := doc.Find("ul").Has("table").Length(); n != 0 {
		t.Fatalf("has none = %d", n)
	}
}

func TestFilterFunc(t *testing.T) {
	doc := load(t)
	got := doc.Find("li").FilterFunc(func(i int, n *Node) bool {
		return i%2 == 0
	}).Length()
	if got != 2 {
		t.Fatalf("filterfunc = %d", got)
	}
}

func TestEqFirstLastSliceGet(t *testing.T) {
	doc := load(t)
	li := doc.Find("li")
	if got := li.Eq(1).Text(); got != "Two" {
		t.Fatalf("eq1 = %q", got)
	}
	if got := li.Eq(-1).Text(); got != "Four" {
		t.Fatalf("eq-1 = %q", got)
	}
	if li.Eq(99).Length() != 0 {
		t.Fatal("eq oob should be empty")
	}
	if got := li.First().Text(); got != "One" {
		t.Fatalf("first = %q", got)
	}
	if got := li.Last().Text(); got != "Four" {
		t.Fatalf("last = %q", got)
	}
	if n := li.Slice(1, 3).Length(); n != 2 {
		t.Fatalf("slice = %d", n)
	}
	if li.Get(0) == nil || li.Get(99) != nil {
		t.Fatal("Get bounds")
	}
	if li.Get(-1).Data != "li" {
		t.Fatal("Get negative")
	}
}

func TestEachAndMap(t *testing.T) {
	doc := load(t)
	var count int
	doc.Find("li").Each(func(i int, n *Node) { count++ })
	if count != 4 {
		t.Fatalf("each count = %d", count)
	}
	texts := doc.Find("li").Map(func(i int, n *Node) string { return strings.TrimSpace(nodeText(n)) })
	want := []string{"One", "Two", "Three", "Four"}
	if !reflect.DeepEqual(texts, want) {
		t.Fatalf("map = %v want %v", texts, want)
	}
}

func TestAddAndRoot(t *testing.T) {
	doc := load(t)
	combined := doc.Find("h1").Add("img")
	if combined.Length() != 2 {
		t.Fatalf("add = %d", combined.Length())
	}
	if doc.Root().Children().Length() == 0 {
		t.Fatal("root children empty")
	}
	if doc.Find("li").End().Children().Length() == 0 {
		t.Fatal("End() should return root")
	}
}

func TestSelectionFindNested(t *testing.T) {
	doc := load(t)
	if n := doc.Find("#main").Find("li").Length(); n != 4 {
		t.Fatalf("nested find = %d", n)
	}
	// Find should only search descendants, not self.
	if n := doc.Find("ul").Find("ul").Length(); n != 0 {
		t.Fatalf("find self excluded = %d", n)
	}
}

func nodeText(n *Node) string {
	var b strings.Builder
	collectText(n, &b)
	return b.String()
}

// --- Parser / malformed HTML tests ---

func TestVoidElements(t *testing.T) {
	doc := Load(`<div><br><hr><img src="x"><p>after</p></div>`)
	if n := doc.Find("br, hr, img").Length(); n != 3 {
		t.Fatalf("void count = %d", n)
	}
	// Void elements must have no children; <p> is a sibling not a child.
	if doc.Find("br").First().Nodes()[0].Children != nil {
		t.Fatal("br should have no children")
	}
	if doc.Find("div > p").Length() != 1 {
		t.Fatal("p should be child of div")
	}
}

func TestImplicitParagraphClose(t *testing.T) {
	doc := Load(`<div><p>one<p>two<p>three</div>`)
	if n := doc.Find("p").Length(); n != 3 {
		t.Fatalf("implicit p = %d", n)
	}
	// They must be siblings, not nested.
	if n := doc.Find("p p").Length(); n != 0 {
		t.Fatalf("p should not nest, got %d", n)
	}
}

func TestImplicitListClose(t *testing.T) {
	doc := Load(`<ul><li>a<li>b<li>c</ul>`)
	if n := doc.Find("li").Length(); n != 3 {
		t.Fatalf("implicit li = %d", n)
	}
	if n := doc.Find("li li").Length(); n != 0 {
		t.Fatalf("li nested = %d", n)
	}
}

func TestImplicitTableRows(t *testing.T) {
	doc := Load(`<table><tr><td>a<td>b<tr><td>c</table>`)
	if n := doc.Find("tr").Length(); n != 2 {
		t.Fatalf("tr = %d", n)
	}
	if n := doc.Find("td").Length(); n != 3 {
		t.Fatalf("td = %d", n)
	}
}

func TestUnclosedTags(t *testing.T) {
	doc := Load(`<div><span>text`)
	if doc.Find("span").Text() != "text" {
		t.Fatal("unclosed span text lost")
	}
	if doc.Find("div span").Length() != 1 {
		t.Fatal("unclosed nesting broken")
	}
}

func TestStrayEndTags(t *testing.T) {
	doc := Load(`<p>hi</b></p></div>done`)
	if got := doc.Find("p").Text(); got != "hi" {
		t.Fatalf("stray end tags: %q", got)
	}
}

func TestAttributeQuotingVariants(t *testing.T) {
	doc := Load(`<a id='single' class="double" data-x=unquoted checked>y</a>`)
	a := doc.Find("a")
	if v, _ := a.Attr("id"); v != "single" {
		t.Fatalf("single quote = %q", v)
	}
	if v, _ := a.Attr("class"); v != "double" {
		t.Fatalf("double quote = %q", v)
	}
	if v, _ := a.Attr("data-x"); v != "unquoted" {
		t.Fatalf("unquoted = %q", v)
	}
	if v, ok := a.Attr("checked"); !ok || v != "" {
		t.Fatalf("valueless = %q ok=%v", v, ok)
	}
}

func TestSelfClosing(t *testing.T) {
	// The self-closing slash is ignored on non-void HTML elements (including
	// unknown/custom elements), matching the HTML5 parser and cheerio: the
	// custom element stays open and the following span nests inside it.
	doc := Load(`<div><custom-el/><span>x</span></div>`)
	if doc.Find("div > span").Length() != 0 {
		t.Fatal("self-closing slash should be ignored: span must nest in custom-el")
	}
	if doc.Find("custom-el > span").Length() != 1 {
		t.Fatal("span should be a child of the still-open custom element")
	}
}

func TestComments(t *testing.T) {
	doc := Load(`<div><!-- hello --><p>x</p></div>`)
	if doc.Find("p").Length() != 1 {
		t.Fatal("comment broke parsing")
	}
	// Comment content preserved in serialization.
	if !strings.Contains(doc.Find("div").Html(), "<!-- hello -->") {
		t.Fatal("comment not preserved")
	}
}

func TestScriptRawText(t *testing.T) {
	doc := Load(`<script>var x = 1 < 2 && 3 > 2;</script><p>after</p>`)
	if got := doc.Find("script").Text(); got != "var x = 1 < 2 && 3 > 2;" {
		t.Fatalf("script raw = %q", got)
	}
	if doc.Find("p").Length() != 1 {
		t.Fatal("content after script lost")
	}
}

func TestStyleRawText(t *testing.T) {
	doc := Load(`<style>a > b { color: red }</style><div>x</div>`)
	if !strings.Contains(doc.Find("style").Text(), "a > b") {
		t.Fatal("style content lost")
	}
	if doc.Find("div").Length() != 1 {
		t.Fatal("content after style lost")
	}
}

func TestEntityDecoding(t *testing.T) {
	doc := Load(`<p>&lt;tag&gt; &amp; &#65; &#x42; &copy;</p>`)
	if got := doc.Find("p").Text(); got != "<tag> & A B ©" {
		t.Fatalf("entities = %q", got)
	}
}

func TestAttrEntityDecoding(t *testing.T) {
	doc := Load(`<a href="a?x=1&amp;y=2">l</a>`)
	if v, _ := doc.Find("a").Attr("href"); v != "a?x=1&y=2" {
		t.Fatalf("attr entity = %q", v)
	}
}

func TestDoctype(t *testing.T) {
	doc := Load(`<!DOCTYPE html><html><body>x</body></html>`)
	if doc.Find("body").Text() != "x" {
		t.Fatal("doctype broke parse")
	}
}

func TestEmptyAndGarbage(t *testing.T) {
	if Load("").Find("div").Length() != 0 {
		t.Fatal("empty doc")
	}
	if Load("just text").Root().Children().Length() != 0 {
		t.Fatal("text-only doc should have no elements")
	}
	if Load("< not a tag >").Find("*").Length() != 0 {
		t.Fatal("garbage should produce no elements")
	}
}

func TestDuplicateAttributes(t *testing.T) {
	doc := Load(`<a id="first" id="second">x</a>`)
	if v, _ := doc.Find("a").Attr("id"); v != "first" {
		t.Fatalf("dup attr = %q, want first", v)
	}
}

func TestCaseInsensitiveTags(t *testing.T) {
	doc := Load(`<DIV CLASS="X"><P>hi</P></DIV>`)
	if doc.Find("div.X").Length() != 1 {
		t.Fatal("uppercase tag/attr")
	}
	if doc.Find("p").Text() != "hi" {
		t.Fatal("uppercase p")
	}
}

func TestNthChildComplex(t *testing.T) {
	doc := Load(`<ul>` + strings.Repeat("<li>x</li>", 10) + `</ul>`)
	if n := doc.Find("li:nth-child(3n)").Length(); n != 3 { // 3,6,9
		t.Fatalf("3n = %d, want 3", n)
	}
	if n := doc.Find("li:nth-child(3n+1)").Length(); n != 4 { // 1,4,7,10
		t.Fatalf("3n+1 = %d, want 4", n)
	}
	if n := doc.Find("li:nth-child(-n+3)").Length(); n != 3 { // 1,2,3
		t.Fatalf("-n+3 = %d, want 3", n)
	}
}

func TestSelectorWhitespaceTolerance(t *testing.T) {
	doc := load(t)
	if doc.Find("  ul   >   li.odd  ").Length() != 2 {
		t.Fatal("whitespace tolerance")
	}
	if doc.Find("h1 ,  img").Length() != 2 {
		t.Fatal("comma whitespace")
	}
}

func TestInvalidSelectors(t *testing.T) {
	doc := load(t)
	for _, sel := range []string{"", "   ", ">", "#", ".", "[", "::", ":unknown", ":nth-child(abc)"} {
		if got := doc.Find(sel).Length(); got != 0 {
			t.Errorf("invalid selector %q matched %d", sel, got)
		}
	}
}

func TestNodeAppendAndTypes(t *testing.T) {
	doc := load(t)
	root := doc.Root().Nodes()[0]
	if root.Type != DocumentNode {
		t.Fatal("root should be document node")
	}
	// Ensure parent links are set.
	li := doc.Find("li").First().Nodes()[0]
	if li.Parent == nil || li.Parent.Data != "ul" {
		t.Fatal("parent link broken")
	}
}

func TestDeterministicOrder(t *testing.T) {
	doc := load(t)
	a := doc.Find("li").Map(func(i int, n *Node) string { return strings.TrimSpace(nodeText(n)) })
	b := doc.Find("li").Map(func(i int, n *Node) string { return strings.TrimSpace(nodeText(n)) })
	if !reflect.DeepEqual(a, b) {
		t.Fatal("non-deterministic order")
	}
}
