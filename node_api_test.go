package cheerio

import (
	"strings"
	"testing"
)

func TestNodeTextInnerOuter(t *testing.T) {
	doc := Load(`<div id="d"><p class="a">Hi <b>there</b></p></div>`)
	n := doc.Find("#d").Get(0)
	if got, want := n.Text(), "Hi there"; got != want {
		t.Errorf("Text() = %q, want %q", got, want)
	}
	if got, want := n.InnerHTML(), `<p class="a">Hi <b>there</b></p>`; got != want {
		t.Errorf("InnerHTML() = %q, want %q", got, want)
	}
	if got, want := n.OuterHTML(), `<div id="d"><p class="a">Hi <b>there</b></p></div>`; got != want {
		t.Errorf("OuterHTML() = %q, want %q", got, want)
	}
}

func TestNodeRender(t *testing.T) {
	doc := Load(`<span>x&amp;y</span>`)
	n := doc.Find("span").Get(0)
	var b strings.Builder
	if err := n.Render(&b); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if got, want := b.String(), `<span>x&amp;y</span>`; got != want {
		t.Errorf("Render = %q, want %q", got, want)
	}
}

func TestNodeCloneIndependent(t *testing.T) {
	doc := Load(`<ul><li>a</li></ul>`)
	orig := doc.Find("li").Get(0)
	cp := orig.Clone()
	if cp.Parent != nil {
		t.Errorf("clone Parent = %v, want nil", cp.Parent)
	}
	cp.SetAttr("class", "x")
	if _, ok := orig.AttrValue("class"); ok {
		t.Errorf("mutating clone changed original")
	}
	if cp.OuterHTML() != `<li class="x">a</li>` {
		t.Errorf("clone OuterHTML = %q", cp.OuterHTML())
	}
}

func TestNodeChildrenAndSiblings(t *testing.T) {
	doc := Load(`<ul><li id="a">a</li>x<li id="b">b</li><li id="c">c</li></ul>`)
	ul := doc.Find("ul").Get(0)
	if ul.FirstChild().TagName() != "li" {
		t.Errorf("FirstChild tag = %q", ul.FirstChild().TagName())
	}
	if id, _ := ul.LastChild().AttrValue("id"); id != "c" {
		t.Errorf("LastChild id = %q, want c", id)
	}
	a := doc.Find("#a").Get(0)
	// NextSibling is the text node "x".
	if ns := a.NextSibling(); ns == nil || ns.Type != TextNode || ns.Data != "x" {
		t.Errorf("NextSibling = %+v, want text x", ns)
	}
	if id, _ := a.NextElementSibling().AttrValue("id"); id != "b" {
		t.Errorf("NextElementSibling id = %q, want b", id)
	}
	c := doc.Find("#c").Get(0)
	if id, _ := c.PrevElementSibling().AttrValue("id"); id != "b" {
		t.Errorf("PrevElementSibling id = %q, want b", id)
	}
	if a.PrevSibling() != nil {
		t.Errorf("PrevSibling of first = %v, want nil", a.PrevSibling())
	}
}

func TestNodeClassAttrMatches(t *testing.T) {
	doc := Load(`<p class="foo bar" data-x="1">hi</p>`)
	p := doc.Find("p").Get(0)
	if !p.HasClass("bar") || p.HasClass("baz") {
		t.Errorf("HasClass wrong")
	}
	if p.TagName() != "p" {
		t.Errorf("TagName = %q", p.TagName())
	}
	p.SetAttr("id", "z")
	if v, _ := p.AttrValue("id"); v != "z" {
		t.Errorf("SetAttr id = %q", v)
	}
	p.RemoveAttr("data-x")
	if _, ok := p.AttrValue("data-x"); ok {
		t.Errorf("RemoveAttr failed")
	}
	cases := []struct {
		sel  string
		want bool
	}{
		{"p.foo", true},
		{"p#z", true},
		{"p.baz", false},
		{"[id=z]", true},
		{"div", false},
	}
	for _, tc := range cases {
		if got := p.Matches(tc.sel); got != tc.want {
			t.Errorf("Matches(%q) = %v, want %v", tc.sel, got, tc.want)
		}
	}
}

func BenchmarkNodeOuterHTML(b *testing.B) {
	doc := Load(`<div><p class="a">Hi <b>there</b> friend</p><p>second</p></div>`)
	n := doc.Find("div").Get(0)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = n.OuterHTML()
	}
}
