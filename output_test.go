package cheerio

import (
	"reflect"
	"strings"
	"testing"
)

func TestSelectionRender(t *testing.T) {
	doc := Load(`<ul><li>a</li><li>b</li></ul>`)
	var b strings.Builder
	if err := doc.Find("li").Render(&b); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if got, want := b.String(), `<li>a</li><li>b</li>`; got != want {
		t.Errorf("Render = %q, want %q", got, want)
	}
}

func TestSelectionToArray(t *testing.T) {
	doc := Load(`<b>1</b><b>2</b>`)
	s := doc.Find("b")
	arr := s.ToArray()
	if len(arr) != 2 {
		t.Fatalf("len = %d, want 2", len(arr))
	}
	arr[0] = nil // must not affect the selection's own slice
	if s.Nodes()[0] == nil {
		t.Errorf("ToArray did not return an independent copy")
	}
}

func TestSelectionAttrs(t *testing.T) {
	doc := Load(`<input type="text" name="q" value="hi">`)
	got := doc.Find("input").Attrs()
	want := map[string]string{"type": "text", "name": "q", "value": "hi"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Attrs = %v, want %v", got, want)
	}
	if len(Load(``).Root().Attrs()) != 0 {
		t.Errorf("Attrs of non-element should be empty")
	}
}

func TestSelectionRemoveData(t *testing.T) {
	doc := Load(`<div data-foo="1" data-bar-baz="2" id="keep">x</div>`)
	sel := doc.Find("div")
	sel.RemoveData("fooBar") // no such key, no-op
	sel.RemoveData("foo")
	if _, ok := sel.Data("foo"); ok {
		t.Errorf("data-foo should be removed")
	}
	if v, _ := sel.Data("barBaz"); v != "2" {
		t.Errorf("data-bar-baz should remain, got %q", v)
	}
	sel.RemoveData() // remove all
	if len(sel.DataMap()) != 0 {
		t.Errorf("all data should be removed, got %v", sel.DataMap())
	}
	if _, ok := sel.Attr("id"); !ok {
		t.Errorf("non-data attribute id should survive")
	}
}

func TestSelectionMapNodes(t *testing.T) {
	doc := Load(`<ul><li><a>1</a></li><li></li><li><a>2</a></li></ul>`)
	got := doc.Find("li").MapNodes(func(i int, n *Node) *Node {
		return firstElementChild(n) // nil for the empty <li>
	})
	if len(got) != 2 {
		t.Fatalf("MapNodes len = %d, want 2 (nils dropped)", len(got))
	}
	if got[0].TagName() != "a" || got[0].Text() != "1" {
		t.Errorf("MapNodes[0] = %q", got[0].OuterHTML())
	}
}

func TestDocumentHtmlText(t *testing.T) {
	doc := Load(`<!DOCTYPE html><h1>Hi</h1><p>Body</p>`)
	if got, want := doc.Html(), `<!DOCTYPE html><h1>Hi</h1><p>Body</p>`; got != want {
		t.Errorf("Html = %q, want %q", got, want)
	}
	if got, want := doc.Text(), "HiBody"; got != want {
		t.Errorf("Text = %q, want %q", got, want)
	}
	var b strings.Builder
	if err := doc.Render(&b); err != nil || b.String() != doc.Html() {
		t.Errorf("Render mismatch: err=%v", err)
	}
}
