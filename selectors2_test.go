package cheerio

import (
	"reflect"
	"testing"
)

func TestContainsPseudo(t *testing.T) {
	doc := Load(`<p>hello world</p><p>goodbye</p><p>say hello</p>`)
	if n := doc.Find("p:contains(hello)").Length(); n != 2 {
		t.Fatalf(":contains = %d want 2", n)
	}
	if n := doc.Find(`p:contains("goodbye")`).Length(); n != 1 {
		t.Fatalf(":contains quoted = %d want 1", n)
	}
}

func TestHasPseudo(t *testing.T) {
	doc := Load(`<div><p><a>x</a></p></div><div><p>y</p></div>`)
	if n := doc.Find("div:has(a)").Length(); n != 1 {
		t.Fatalf(":has = %d want 1", n)
	}
	if n := doc.Find("div:has(p)").Length(); n != 2 {
		t.Fatalf(":has(p) = %d want 2", n)
	}
}

func TestOfTypePseudos(t *testing.T) {
	doc := Load(`<div><p>1</p><span>a</span><p>2</p><span>b</span><p>3</p></div>`)
	if n := doc.Find("p:first-of-type").Length(); n != 1 {
		t.Fatalf("first-of-type = %d", n)
	}
	if got := doc.Find("p:first-of-type").Text(); got != "1" {
		t.Fatalf("first-of-type text = %q", got)
	}
	if got := doc.Find("p:last-of-type").Text(); got != "3" {
		t.Fatalf("last-of-type text = %q", got)
	}
	if got := doc.Find("p:nth-of-type(2)").Text(); got != "2" {
		t.Fatalf("nth-of-type(2) = %q", got)
	}
	if got := doc.Find("span:nth-last-of-type(1)").Text(); got != "b" {
		t.Fatalf("nth-last-of-type = %q", got)
	}
}

func TestOnlyOfType(t *testing.T) {
	doc := Load(`<div><p>a</p><span>only</span></div>`)
	if got := doc.Find("span:only-of-type").Text(); got != "only" {
		t.Fatalf("only-of-type = %q", got)
	}
	if n := doc.Find("p:only-of-type").Length(); n != 1 {
		t.Fatalf("only-of-type p = %d want 1", n)
	}
}

func TestRootPseudo(t *testing.T) {
	doc := Load(`<html><body><div></div></body></html>`)
	r := doc.Find(":root")
	if r.Length() != 1 || r.TagName() != "html" {
		t.Fatalf(":root = %d tag=%q", r.Length(), r.TagName())
	}
}

func TestFormPseudos(t *testing.T) {
	doc := Load(`<form>
		<input type="text" name="a">
		<input type="checkbox" checked>
		<textarea></textarea>
		<select><option>x</option><option selected>y</option></select>
		<button>b</button>
		<input type="text" disabled>
	</form>`)
	if n := doc.Find(":input").Length(); n != 6 {
		t.Fatalf(":input = %d want 6", n)
	}
	if n := doc.Find(":checked").Length(); n != 2 { // checkbox + option
		t.Fatalf(":checked = %d want 2", n)
	}
	if n := doc.Find(":selected").Length(); n != 1 {
		t.Fatalf(":selected = %d want 1", n)
	}
	if n := doc.Find(":disabled").Length(); n != 1 {
		t.Fatalf(":disabled = %d want 1", n)
	}
	if n := doc.Find("input:enabled").Length(); n != 2 { // text + checkbox, not disabled
		t.Fatalf(":enabled = %d want 2", n)
	}
}

func TestHeaderPseudo(t *testing.T) {
	doc := Load(`<h1>a</h1><h3>b</h3><p>c</p><h6>d</h6>`)
	if n := doc.Find(":header").Length(); n != 3 {
		t.Fatalf(":header = %d want 3", n)
	}
}

func TestParentPseudo(t *testing.T) {
	doc := Load(`<div><span>x</span></div><div></div>`)
	if n := doc.Find("div:parent").Length(); n != 1 {
		t.Fatalf(":parent = %d want 1", n)
	}
}

func TestAttrCaseInsensitive(t *testing.T) {
	doc := Load(`<input type="TEXT"><input type="text"><input type="Email">`)
	if n := doc.Find(`[type="text" i]`).Length(); n != 2 {
		t.Fatalf("case-insensitive = %d want 2", n)
	}
	if n := doc.Find(`[type="text"]`).Length(); n != 1 {
		t.Fatalf("case-sensitive = %d want 1", n)
	}
	if n := doc.Find(`[type^="e" i]`).Length(); n != 1 {
		t.Fatalf("prefix ci = %d want 1", n)
	}
}

func TestNamespaceTag(t *testing.T) {
	doc := Load(`<svg><rect></rect><circle></circle></svg>`)
	if n := doc.Find("svg|rect").Length(); n != 1 {
		t.Fatalf("namespace tag = %d want 1", n)
	}
	if n := doc.Find("*|circle").Length(); n != 1 {
		t.Fatalf("universal namespace = %d want 1", n)
	}
}

func TestContents(t *testing.T) {
	doc := Load(`<div>text<b>bold</b><!--c--></div>`)
	c := doc.Find("div").Contents()
	if c.Length() != 3 {
		t.Fatalf("contents = %d want 3", c.Length())
	}
}

func TestIndex(t *testing.T) {
	doc := Load(`<ul><li>a</li><li class="t">b</li><li>c</li></ul>`)
	if got := doc.Find("li.t").Index(); got != 1 {
		t.Fatalf("index = %d want 1", got)
	}
	if got := doc.Find("li.t").IndexSelector("li"); got != 1 {
		t.Fatalf("indexSelector = %d want 1", got)
	}
	li := doc.Find("li")
	if got := li.IndexOfNode(li.Get(2)); got != 2 {
		t.Fatalf("indexOfNode = %d want 2", got)
	}
}

func TestNextPrevUntil(t *testing.T) {
	doc := Load(`<ul><li class="s">1</li><li>2</li><li>3</li><li class="e">4</li><li>5</li></ul>`)
	if n := doc.Find("li.s").NextUntil(".e").Length(); n != 2 { // 2,3
		t.Fatalf("nextUntil = %d want 2", n)
	}
	if n := doc.Find("li.e").PrevUntil(".s").Length(); n != 2 { // 3,2
		t.Fatalf("prevUntil = %d want 2", n)
	}
}

func TestParentsUntil(t *testing.T) {
	doc := Load(`<div class="top"><section><article><p><b class="x">y</b></p></article></section></div>`)
	got := doc.Find(".x").ParentsUntil(".top").Map(func(i int, n *Node) string { return n.Data })
	want := []string{"p", "article", "section"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parentsUntil = %v want %v", got, want)
	}
}

func TestEvenOdd(t *testing.T) {
	doc := Load(`<ul><li>0</li><li>1</li><li>2</li><li>3</li></ul>`)
	if got := doc.Find("li").Even().Map(func(i int, n *Node) string { return nodeText(n) }); !reflect.DeepEqual(got, []string{"0", "2"}) {
		t.Fatalf("even = %v", got)
	}
	if got := doc.Find("li").Odd().Map(func(i int, n *Node) string { return nodeText(n) }); !reflect.DeepEqual(got, []string{"1", "3"}) {
		t.Fatalf("odd = %v", got)
	}
}

func TestEndAndAddBack(t *testing.T) {
	doc := Load(`<div><ul><li>a</li><li>b</li></ul></div>`)
	// End returns previous stage.
	ul := doc.Find("ul")
	if ul.Find("li").End() != ul {
		// prev should be the ul selection object
		if ul.Find("li").End().TagName() != "ul" {
			t.Fatal("End should return ul selection")
		}
	}
	// AddBack merges li with ul.
	n := doc.Find("ul").Find("li").AddBack().Length()
	if n != 3 { // ul + 2 li
		t.Fatalf("addBack = %d want 3", n)
	}
}

func TestAddVariants(t *testing.T) {
	doc := Load(`<h1>h</h1><p>p</p><span>s</span>`)
	base := doc.Find("h1")
	if n := base.AddSelection(doc.Find("p")).Length(); n != 2 {
		t.Fatalf("addSelection = %d", n)
	}
	if n := base.AddNodes(doc.Find("span").Get(0)).Length(); n != 2 {
		t.Fatalf("addNodes = %d", n)
	}
	// Document order preserved.
	got := doc.Find("span").AddSelection(doc.Find("h1")).Map(func(i int, n *Node) string { return n.Data })
	want := []string{"h1", "span"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("add order = %v want %v", got, want)
	}
}

func TestImplicitTbody(t *testing.T) {
	doc := Load(`<table><tr><td>a</td></tr></table>`)
	if n := doc.Find("tbody").Length(); n != 1 {
		t.Fatalf("implicit tbody = %d want 1", n)
	}
	if n := doc.Find("table > tbody > tr > td").Length(); n != 1 {
		t.Fatalf("tbody nesting = %d want 1", n)
	}
}
