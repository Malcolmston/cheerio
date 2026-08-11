package cheerio

import "testing"

func TestParseHTML(t *testing.T) {
	nodes := ParseHTML(`<p>one</p><!--c--><p>two</p>`)
	var elems int
	for _, n := range nodes {
		if n.Parent != nil {
			t.Errorf("node %q has non-nil parent", n.Data)
		}
		if n.Type == ElementNode {
			elems++
		}
	}
	if elems != 2 {
		t.Errorf("got %d elements, want 2", elems)
	}
	if RenderNodes(nodes...) != `<p>one</p><!--c--><p>two</p>` {
		t.Errorf("RenderNodes round-trip = %q", RenderNodes(nodes...))
	}
}

func TestContains(t *testing.T) {
	doc := Load(`<div id="outer"><section><p id="inner">x</p></section></div>`)
	outer := doc.Find("#outer").Get(0)
	inner := doc.Find("#inner").Get(0)
	sec := doc.Find("section").Get(0)
	if !Contains(outer, inner) {
		t.Errorf("outer should contain inner")
	}
	if !Contains(sec, inner) {
		t.Errorf("section should contain inner")
	}
	if Contains(inner, outer) {
		t.Errorf("inner should not contain outer")
	}
	if Contains(outer, outer) {
		t.Errorf("node should not contain itself")
	}
	if Contains(nil, inner) || Contains(outer, nil) {
		t.Errorf("nil handling wrong")
	}
}

func TestMerge(t *testing.T) {
	doc := Load(`<a id="1"></a><a id="2"></a><a id="3"></a>`)
	all := doc.Find("a").Nodes()
	head := all[:1]
	got := Merge(head, all[1], all[2], all[0]) // includes a duplicate of head[0]
	// Like jQuery/cheerio $.merge, Merge concatenates without de-duplicating.
	if len(got) != 4 {
		t.Fatalf("Merge len = %d, want 4 (concatenated)", len(got))
	}
	wantIDs := []string{"1", "2", "3", "1"}
	for i, n := range got {
		if id, _ := n.AttrValue("id"); id != wantIDs[i] {
			t.Errorf("Merge[%d] id = %q, want %q", i, id, wantIDs[i])
		}
	}
}

func TestTextOf(t *testing.T) {
	doc := Load(`<h1>Title</h1><p>Body <b>text</b></p>`)
	nodes := doc.Find("h1, p").Nodes()
	if got, want := TextOf(nodes...), "TitleBody text"; got != want {
		t.Errorf("TextOf = %q, want %q", got, want)
	}
}
