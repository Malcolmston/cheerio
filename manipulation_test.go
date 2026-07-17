package cheerio

import (
	"testing"
)

func TestAppendPrepend(t *testing.T) {
	doc := Load(`<ul><li>a</li></ul>`)
	doc.Find("ul").Append(`<li>b</li>`).Prepend(`<li>z</li>`)
	got := doc.Find("ul").Html()
	want := `<li>z</li><li>a</li><li>b</li>`
	if got != want {
		t.Fatalf("append/prepend = %q want %q", got, want)
	}
}

func TestAppendMultipleNodesOrder(t *testing.T) {
	doc := Load(`<div></div>`)
	doc.Find("div").Append(`<span>1</span><span>2</span>`)
	if got := doc.Find("div").Html(); got != `<span>1</span><span>2</span>` {
		t.Fatalf("append order = %q", got)
	}
	doc.Find("div").Empty()
	doc.Find("div").Prepend(`<span>1</span><span>2</span>`)
	if got := doc.Find("div").Html(); got != `<span>1</span><span>2</span>` {
		t.Fatalf("prepend order = %q", got)
	}
}

func TestAppendClonesForMultipleTargets(t *testing.T) {
	doc := Load(`<div class="box"></div><div class="box"></div>`)
	doc.Find(".box").Append(`<b>x</b>`)
	if n := doc.Find(".box > b").Length(); n != 2 {
		t.Fatalf("append to multiple targets = %d want 2", n)
	}
}

func TestBeforeAfter(t *testing.T) {
	doc := Load(`<div><p id="mid">mid</p></div>`)
	doc.Find("#mid").Before(`<p>before</p>`).After(`<p>after</p>`)
	got := doc.Find("div").Html()
	want := `<p>before</p><p id="mid">mid</p><p>after</p>`
	if got != want {
		t.Fatalf("before/after = %q want %q", got, want)
	}
}

func TestAfterMultipleOrder(t *testing.T) {
	doc := Load(`<div><a>x</a></div>`)
	doc.Find("a").After(`<b>1</b><b>2</b>`)
	if got := doc.Find("div").Html(); got != `<a>x</a><b>1</b><b>2</b>` {
		t.Fatalf("after order = %q", got)
	}
}

func TestReplaceWith(t *testing.T) {
	doc := Load(`<div><span class="old">x</span></div>`)
	doc.Find(".old").ReplaceWith(`<em>new</em>`)
	if doc.Find(".old").Length() != 0 {
		t.Fatal("old node should be gone")
	}
	if got := doc.Find("div").Html(); got != `<em>new</em>` {
		t.Fatalf("replaceWith = %q", got)
	}
}

func TestRemove(t *testing.T) {
	doc := Load(`<ul><li class="a">1</li><li class="b">2</li></ul>`)
	doc.Find("li.a").Remove()
	if got := doc.Find("ul").Html(); got != `<li class="b">2</li>` {
		t.Fatalf("remove = %q", got)
	}
	doc2 := Load(`<ul><li class="a">1</li><li class="b">2</li></ul>`)
	doc2.Find("li").Remove(".b")
	if got := doc2.Find("ul").Html(); got != `<li class="a">1</li>` {
		t.Fatalf("remove(selector) = %q", got)
	}
}

func TestEmpty(t *testing.T) {
	doc := Load(`<div><p>a</p><p>b</p></div>`)
	doc.Find("div").Empty()
	if got := doc.Find("div").Html(); got != "" {
		t.Fatalf("empty = %q", got)
	}
}

func TestClone(t *testing.T) {
	doc := Load(`<div><b class="c">hi</b></div>`)
	orig := doc.Find("b")
	clone := orig.Clone()
	clone.SetText("bye")
	// Original must be untouched.
	if got := orig.Text(); got != "hi" {
		t.Fatalf("clone mutated original: %q", got)
	}
	if got := clone.Text(); got != "bye" {
		t.Fatalf("clone text = %q", got)
	}
	// Clone is detached.
	if clone.Nodes()[0].Parent != nil {
		t.Fatal("clone should be detached")
	}
}

func TestAppendSelectionMoves(t *testing.T) {
	doc := Load(`<div id="src"><b>x</b></div><div id="dst"></div>`)
	src := doc.Find("#src b")
	doc.Find("#dst").AppendSelection(src)
	if doc.Find("#src b").Length() != 0 {
		t.Fatal("node should have moved out of src")
	}
	if got := doc.Find("#dst").Html(); got != `<b>x</b>` {
		t.Fatalf("appendSelection = %q", got)
	}
}

func TestAppendToPrependTo(t *testing.T) {
	doc := Load(`<div id="a"></div><p class="x">1</p><p class="x">2</p>`)
	doc.Find("p.x").AppendTo("#a")
	if n := doc.Find("#a > p").Length(); n != 2 {
		t.Fatalf("appendTo = %d", n)
	}
	doc2 := Load(`<div id="a"><i>keep</i></div><p class="x">1</p>`)
	doc2.Find("p.x").PrependTo("#a")
	if got := doc2.Find("#a").Html(); got != `<p class="x">1</p><i>keep</i>` {
		t.Fatalf("prependTo = %q", got)
	}
}

func TestInsertBeforeAfter(t *testing.T) {
	doc := Load(`<div><p id="t">t</p></div><span class="s">s</span>`)
	doc.Find(".s").InsertBefore("#t")
	if got := doc.Find("div").Html(); got != `<span class="s">s</span><p id="t">t</p>` {
		t.Fatalf("insertBefore = %q", got)
	}
	doc2 := Load(`<div><p id="t">t</p></div><span class="s">s</span>`)
	doc2.Find(".s").InsertAfter("#t")
	if got := doc2.Find("div").Html(); got != `<p id="t">t</p><span class="s">s</span>` {
		t.Fatalf("insertAfter = %q", got)
	}
}

func TestWrap(t *testing.T) {
	doc := Load(`<div><span>x</span></div>`)
	doc.Find("span").Wrap(`<p class="w"></p>`)
	if got := doc.Find("div").Html(); got != `<p class="w"><span>x</span></p>` {
		t.Fatalf("wrap = %q", got)
	}
}

func TestWrapNested(t *testing.T) {
	doc := Load(`<div><span>x</span></div>`)
	doc.Find("span").Wrap(`<a><b></b></a>`)
	if got := doc.Find("div").Html(); got != `<a><b><span>x</span></b></a>` {
		t.Fatalf("wrap nested = %q", got)
	}
}

func TestWrapAll(t *testing.T) {
	doc := Load(`<div><span class="i">1</span><span class="i">2</span></div>`)
	doc.Find(".i").WrapAll(`<p class="w"></p>`)
	if got := doc.Find("div").Html(); got != `<p class="w"><span class="i">1</span><span class="i">2</span></p>` {
		t.Fatalf("wrapAll = %q", got)
	}
}

func TestWrapInner(t *testing.T) {
	doc := Load(`<div class="d">a<b>x</b></div>`)
	doc.Find(".d").WrapInner(`<section></section>`)
	if got := doc.Find(".d").Html(); got != `<section>a<b>x</b></section>` {
		t.Fatalf("wrapInner = %q", got)
	}
}

func TestUnwrap(t *testing.T) {
	doc := Load(`<div><p class="w"><span>x</span></p></div>`)
	doc.Find("span").Unwrap()
	if got := doc.Find("div").Html(); got != `<span>x</span>` {
		t.Fatalf("unwrap = %q", got)
	}
}

func TestSetHtml(t *testing.T) {
	doc := Load(`<div>old</div>`)
	doc.Find("div").SetHtml(`<b>new</b>`)
	if got := doc.Find("div").Html(); got != `<b>new</b>` {
		t.Fatalf("setHtml = %q", got)
	}
}

func TestToString(t *testing.T) {
	doc := Load(`<ul><li>a</li><li>b</li></ul>`)
	got := doc.Find("li").ToString()
	if got != `<li>a</li><li>b</li>` {
		t.Fatalf("toString = %q", got)
	}
}

func TestSetVal(t *testing.T) {
	doc := Load(`<input type="text" value="a">`)
	doc.Find("input").SetVal("b")
	if got := doc.Find("input").Val(); got != "b" {
		t.Fatalf("setVal input = %q", got)
	}
	doc2 := Load(`<textarea>a</textarea>`)
	doc2.Find("textarea").SetVal("hello")
	if got := doc2.Find("textarea").Val(); got != "hello" {
		t.Fatalf("setVal textarea = %q", got)
	}
	doc3 := Load(`<select><option value="a">A</option><option value="b">B</option></select>`)
	doc3.Find("select").SetVal("b")
	if got := doc3.Find("select").Val(); got != "b" {
		t.Fatalf("setVal select = %q", got)
	}
}

func TestSetAttrs(t *testing.T) {
	doc := Load(`<a>x</a>`)
	doc.Find("a").SetAttrs(map[string]string{"href": "/go", "title": "t"})
	if v, _ := doc.Find("a").Attr("href"); v != "/go" {
		t.Fatalf("setAttrs href = %q", v)
	}
	if v, _ := doc.Find("a").Attr("title"); v != "t" {
		t.Fatalf("setAttrs title = %q", v)
	}
}

func TestProp(t *testing.T) {
	doc := Load(`<input type="checkbox" checked><div id="d">hi</div>`)
	if v, ok := doc.Find("input").Prop("checked"); !ok || v != "true" {
		t.Fatalf("prop checked = %q ok=%v", v, ok)
	}
	if v, _ := doc.Find("div").Prop("tagName"); v != "DIV" {
		t.Fatalf("prop tagName = %q", v)
	}
	if v, _ := doc.Find("#d").Prop("textContent"); v != "hi" {
		t.Fatalf("prop textContent = %q", v)
	}
	// SetProp boolean.
	doc.Find("input").SetProp("checked", "false")
	if _, ok := doc.Find("input").Attr("checked"); ok {
		t.Fatal("SetProp checked=false should remove attr")
	}
	doc.Find("input").SetProp("checked", "true")
	if _, ok := doc.Find("input").Attr("checked"); !ok {
		t.Fatal("SetProp checked=true should add attr")
	}
}

func TestCssGetSet(t *testing.T) {
	doc := Load(`<div style="color: red; margin: 0"></div>`)
	d := doc.Find("div")
	if got := d.Css("color"); got != "red" {
		t.Fatalf("css get = %q", got)
	}
	d.SetCss("color", "blue")
	if got := d.Css("color"); got != "blue" {
		t.Fatalf("css set = %q", got)
	}
	// Adding a new property keeps existing ones.
	d.SetCss("padding", "1px")
	if got := d.Css("padding"); got != "1px" {
		t.Fatalf("css add = %q", got)
	}
	if got := d.Css("margin"); got != "0" {
		t.Fatalf("css preserved = %q", got)
	}
	// Removing.
	d.SetCss("color", "")
	if got := d.Css("color"); got != "" {
		t.Fatalf("css remove = %q", got)
	}
}

func TestCssSerializeDeterministic(t *testing.T) {
	doc := Load(`<div></div>`)
	d := doc.Find("div")
	d.SetCss("color", "red")
	d.SetCss("font-size", "10px")
	style, _ := d.Nodes()[0].AttrValue("style")
	if style != "color: red; font-size: 10px;" {
		t.Fatalf("style serialize = %q", style)
	}
}

func TestDataGetSet(t *testing.T) {
	doc := Load(`<div data-foo-bar="1" data-x="y"></div>`)
	d := doc.Find("div")
	if v, ok := d.Data("fooBar"); !ok || v != "1" {
		t.Fatalf("data camelCase = %q ok=%v", v, ok)
	}
	if v, ok := d.Data("foo-bar"); !ok || v != "1" {
		t.Fatalf("data dashed = %q ok=%v", v, ok)
	}
	d.SetData("newKey", "z")
	if v, _ := d.Attr("data-new-key"); v != "z" {
		t.Fatalf("setData attr = %q", v)
	}
	m := d.DataMap()
	if m["fooBar"] != "1" || m["x"] != "y" || m["newKey"] != "z" {
		t.Fatalf("dataMap = %v", m)
	}
	if got := d.DataOr("missing", "def"); got != "def" {
		t.Fatalf("dataOr = %q", got)
	}
}

func TestSerializeArray(t *testing.T) {
	doc := Load(`<form>
		<input type="text" name="a" value="1">
		<input type="checkbox" name="b" value="yes" checked>
		<input type="checkbox" name="c" value="no">
		<input type="submit" name="s" value="go">
		<input type="text" name="d" value="4" disabled>
		<select name="sel"><option value="o1">1</option><option value="o2" selected>2</option></select>
		<textarea name="ta">hello</textarea>
	</form>`)
	fields := doc.Find("form").SerializeArray()
	got := map[string]string{}
	for _, f := range fields {
		got[f.Name] = f.Value
	}
	if got["a"] != "1" {
		t.Fatalf("field a = %q", got["a"])
	}
	if got["b"] != "yes" {
		t.Fatalf("checked checkbox b = %q", got["b"])
	}
	if _, ok := got["c"]; ok {
		t.Fatal("unchecked checkbox c should be absent")
	}
	if _, ok := got["s"]; ok {
		t.Fatal("submit should be absent")
	}
	if _, ok := got["d"]; ok {
		t.Fatal("disabled should be absent")
	}
	if got["sel"] != "o2" {
		t.Fatalf("select = %q", got["sel"])
	}
	if got["ta"] != "hello" {
		t.Fatalf("textarea = %q", got["ta"])
	}
}

func TestSerialize(t *testing.T) {
	doc := Load(`<form><input type="text" name="q" value="a b"><input type="text" name="p" value="x&y"></form>`)
	got := doc.Find("form").Serialize()
	if got != "q=a+b&p=x%26y" {
		t.Fatalf("serialize = %q", got)
	}
}

func TestSerializeMultipleSelect(t *testing.T) {
	doc := Load(`<select name="m" multiple><option value="1" selected>1</option><option value="2" selected>2</option><option value="3">3</option></select>`)
	got := doc.Find("select").Serialize()
	if got != "m=1&m=2" {
		t.Fatalf("multiple select serialize = %q", got)
	}
}

func TestSelectionVariants(t *testing.T) {
	doc := Load(`<div id="dst"><i>keep</i></div><b class="a">1</b><b class="a">2</b>`)
	doc.Find("#dst").PrependSelection(doc.Find("b.a").First())
	if got := doc.Find("#dst").Html(); got != `<b class="a">1</b><i>keep</i>` {
		t.Fatalf("prependSelection = %q", got)
	}

	doc2 := Load(`<div><p id="t">t</p></div><span class="s">s</span>`)
	doc2.Find("#t").BeforeSelection(doc2.Find(".s"))
	if got := doc2.Find("div").Html(); got != `<span class="s">s</span><p id="t">t</p>` {
		t.Fatalf("beforeSelection = %q", got)
	}

	doc3 := Load(`<div><p id="t">t</p></div><span class="s">s</span>`)
	doc3.Find("#t").AfterSelection(doc3.Find(".s"))
	if got := doc3.Find("div").Html(); got != `<p id="t">t</p><span class="s">s</span>` {
		t.Fatalf("afterSelection = %q", got)
	}

	doc4 := Load(`<div><p class="old">x</p></div><em class="new">n</em>`)
	doc4.Find(".old").ReplaceWithSelection(doc4.Find(".new"))
	if got := doc4.Find("div").Html(); got != `<em class="new">n</em>` {
		t.Fatalf("replaceWithSelection = %q", got)
	}
}

func TestSetCssMap(t *testing.T) {
	doc := Load(`<div></div>`)
	d := doc.Find("div")
	d.SetCssMap(map[string]string{"color": "red"})
	if got := d.Css("color"); got != "red" {
		t.Fatalf("setCssMap = %q", got)
	}
}

func TestJoinHTMLVariadic(t *testing.T) {
	doc := Load(`<div></div>`)
	doc.Find("div").Append(`<a>1</a>`, `<a>2</a>`, `<a>3</a>`)
	if n := doc.Find("div > a").Length(); n != 3 {
		t.Fatalf("variadic append = %d want 3", n)
	}
}
