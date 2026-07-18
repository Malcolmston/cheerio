package cheerio

// Parity tests encoding known-answer vectors taken directly from the upstream
// Node.js cheerio project (cheeriojs/cheerio) test suite, principally
// src/api/attributes.spec.ts, src/api/css.spec.ts and
// src/api/traversing.spec.ts together with src/__fixtures__/fixtures.ts.
//
// Each assertion mirrors an upstream expect(...) so the Go port's observable
// behavior can be checked against cheerio's own expected values. Where cheerio
// relies on a full-document parse (synthesizing <html>/<body>) or on JavaScript
// value semantics (arrays, JSON/number coercion) that a string-typed, pure
// standard-library Go API deliberately does not reproduce, the vector is
// adapted to the port's fragment model and documented inline.

import "testing"

// Upstream fixtures (verbatim from src/__fixtures__/fixtures.ts).
const (
	fxFruits = `<ul id="fruits">` +
		`<li class="apple">Apple</li>` +
		`<li class="orange">Orange</li>` +
		`<li class="pear">Pear</li>` +
		`</ul>`

	fxVegetables = `<ul id="vegetables">` +
		`<li class="carrot">Carrot</li>` +
		`<li class="sweetcorn">Sweetcorn</li>` +
		`</ul>`

	fxFood = `<ul id="food">` + fxFruits + fxVegetables + `</ul>`

	fxChocolates = `<ul id="chocolates">` +
		`<li class="linth" data-highlight="Lindor" data-origin="swiss">Linth</li>` +
		`<li class="frey" data-taste="sweet" data-best-collection="Mahony">Frey</li>` +
		`<li class="cailler">Cailler</li>` +
		`</ul>`

	fxInputs = `<button id="btn-value" value="button">Button</button>` +
		`<button id="btn-valueless">Button</button>` +
		`<select id="one"><option value="option_not_selected">Option not selected</option><option value="option_selected" selected>Option selected</option></select>` +
		`<select id="one-valueless"><option>Option not selected</option><option selected>Option selected</option></select>` +
		`<select id="one-html-entity"><option>Option not selected</option><option selected>Option &lt;selected&gt;</option></select>` +
		`<select id="one-nested"><option>Option not selected</option><option selected>Option <span>selected</span></option></select>` +
		`<input type="text" value="input_text" />` +
		`<input type="checkbox" name="checkbox_off" value="off" /><input type="checkbox" name="checkbox_on" value="on" checked />` +
		`<input type="checkbox" name="checkbox_valueless" />` +
		`<input type="radio" value="off" name="radio" /><input type="radio" name="radio" value="on" checked />` +
		`<input type="radio" name="radio_valueless" />` +
		`<select id="multi" multiple><option value="1">1</option><option value="2" selected>2</option><option value="3" selected>3</option><option value="4">4</option></select>`

	fxMixedText = `<a>1</a>TEXT<b>2</b>`
)

// TestParityCss mirrors src/api/css.spec.ts.
func TestParityCss(t *testing.T) {
	// (prop): should return a css property value.
	if got := Load(`<li style="hai: there">`).Find("li").Css("hai"); got != "there" {
		t.Errorf("css(hai) = %q, want %q", got, "there")
	}

	// (prop, val): should set a css property.
	d := Load(`<li style="margin: 0;"></li><li></li>`)
	d.Find("li").SetCss("color", "red")
	if got := d.Find("li").AttrOr("style", ""); got != "margin: 0; color: red;" {
		t.Errorf("set color: style = %q, want %q", got, "margin: 0; color: red;")
	}
	if got := d.Find("li").Eq(1).AttrOr("style", ""); got != "color: red;" {
		t.Errorf("set color eq(1): style = %q, want %q", got, "color: red;")
	}

	// (prop, ""): should unset a css property.
	u := Load(`<li style="padding: 1px; margin: 0;">`)
	u.Find("li").SetCss("padding", "")
	if got := u.Find("li").AttrOr("style", ""); got != "margin: 0;" {
		t.Errorf("unset padding: style = %q, want %q", got, "margin: 0;")
	}

	// (prop): should not mangle embedded urls.
	if got := Load(`<li style="background-image:url(http://example.com/img.png);">`).
		Find("li").Css("background-image"); got != "url(http://example.com/img.png)" {
		t.Errorf("css(background-image) = %q, want %q", got, "url(http://example.com/img.png)")
	}

	// parser: should allow any whitespace between declarations.
	ws := Load("<li style=\"one \t:\n 0;\n two \f\r:\v 1\">").Find("li")
	if got := ws.Css("one"); got != "0" {
		t.Errorf("whitespace css(one) = %q, want %q", got, "0")
	}
	if got := ws.Css("two"); got != "1" {
		t.Errorf("whitespace css(two) = %q, want %q", got, "1")
	}

	// parser: should add malformed values to previous field (#1134). GAP.
	if got := Load(`<button style="background-image: url(data:image/png;base64,iVBORw0KGgo)"></button>`).
		Find("button").Css("background-image"); got != "url(data:image/png;base64,iVBORw0KGgo)" {
		t.Errorf("css(#1134) = %q, want %q", got, "url(data:image/png;base64,iVBORw0KGgo)")
	}

	// (prop): should ignore blank properties (":#ccc;color:#aaa;" -> color only).
	if got := Load(`<li style=":#ccc;color:#aaa;">`).Find("li").Css("color"); got != "#aaa" {
		t.Errorf("blank-prop css(color) = %q, want %q", got, "#aaa")
	}
}

// TestParityAttr mirrors the .attr and .removeAttr blocks of attributes.spec.ts.
func TestParityAttr(t *testing.T) {
	doc := Load(fxFruits)

	// (invalid key): invalid attr should get undefined.
	if _, ok := doc.Find(".apple").Attr("lol"); ok {
		t.Error("attr(lol) should be absent")
	}
	// (valid key): valid attr should get value.
	if got := doc.Find(".apple").AttrOr("class", ""); got != "apple" {
		t.Errorf("attr(class) = %q, want %q", got, "apple")
	}
	// (valid key): valid attr should get name when boolean.
	if got := Load(`<input name=email autofocus>`).Find("input").AttrOr("autofocus", "<none>"); got != "" {
		t.Errorf("boolean attr(autofocus) = %q, want empty string (present)", got)
	}

	// (key, value): handle removed boolean attributes.
	apple := Load(fxFruits).Find(".apple")
	apple.SetAttr("autofocus", "autofocus")
	if got := apple.AttrOr("autofocus", "<none>"); got != "autofocus" {
		t.Errorf("attr(autofocus) after set = %q, want %q", got, "autofocus")
	}
	apple.RemoveAttr("autofocus")
	if _, ok := apple.Attr("autofocus"); ok {
		t.Error("attr(autofocus) after removeAttr should be absent")
	}

	// removeAttr: should remove a single attr.
	f := Load(fxFruits).Find("#fruits")
	f.RemoveAttr("id")
	if _, ok := f.Attr("id"); ok {
		t.Error("removeAttr(id) failed")
	}
}

// TestParityClasses mirrors the .hasClass, .addClass, .removeClass and
// .toggleClass blocks of attributes.spec.ts.
func TestParityClasses(t *testing.T) {
	withClass := func(c string) *Selection { return Load(`<div class="` + c + `"></div>`).Find("div") }

	// hasClass: (valid class) should return true.
	if !withClass("foo").HasClass("foo") {
		t.Error("hasClass(foo) on 'foo' = false")
	}
	if !withClass("foo bar").HasClass("foo") {
		t.Error("hasClass(foo) on 'foo bar' = false")
	}
	if !withClass("bar foo bar").HasClass("foo") {
		t.Error("hasClass(foo) on 'bar foo bar' = false")
	}
	// hasClass: (invalid class) should return false.
	if withClass("foo-bar").HasClass("foo") {
		t.Error("hasClass(foo) on 'foo-bar' = true")
	}
	if withClass("foo-bar").HasClass("foo-ba") {
		t.Error("hasClass(foo-ba) on 'foo-bar' = true")
	}
	// hasClass: (empty string argument) should return false.
	if withClass("foo").HasClass("") {
		t.Error("hasClass('') = true")
	}

	// addClass: (class class class) should add multiple classes.
	doc := Load(fxFruits)
	doc.Find(".apple").AddClass("fruit red tasty")
	for _, c := range []string{"apple", "fruit", "red", "tasty"} {
		if !doc.Find(".apple").HasClass(c) {
			t.Errorf("addClass: expected class %q present", c)
		}
	}

	// removeClass: (single class) should remove a single class.
	pear := Load(fxFruits).Find(".pear")
	pear.AddClass("fruit")
	pear.RemoveClass("fruit")
	if pear.HasClass("fruit") {
		t.Error("removeClass(fruit) left the class present")
	}
	if !pear.HasClass("pear") {
		t.Error("removeClass(fruit) removed unrelated class 'pear'")
	}

	// toggleClass: toggling adds then removes.
	tg := Load(`<div class="a"></div>`).Find("div")
	tg.ToggleClass("b")
	if !tg.HasClass("b") {
		t.Error("toggleClass(b) did not add")
	}
	tg.ToggleClass("b")
	if tg.HasClass("b") {
		t.Error("toggleClass(b) did not remove")
	}
}

// TestParityVal mirrors the .val block of attributes.spec.ts.
func TestParityVal(t *testing.T) {
	doc := Load(fxInputs)

	cases := []struct {
		sel, want, desc string
	}{
		{"#btn-value", "button", "button value"},
		{"select#one", "option_selected", "select selected value"},
		{"select#one-valueless", "Option selected", "select no-value -> text"},
		{"select#one-html-entity", "Option <selected>", "select entity-decoded text"},
		{"select#one-nested", "Option selected", "select nested text"},
		{`input[type="text"]`, "input_text", "text input value"},
		{`input[name="checkbox_on"]`, "on", "checked checkbox value"},
		{`input[name="checkbox_off"]`, "off", "unchecked checkbox value"},
		{`input[name="checkbox_valueless"]`, "on", "valueless checkbox -> on (GAP)"},
		{`input[name="radio_valueless"]`, "on", "valueless radio -> on (GAP)"},
	}
	for _, c := range cases {
		if got := doc.Find(c.sel).Val(); got != c.want {
			t.Errorf("val(%s) [%s] = %q, want %q", c.sel, c.desc, got, c.want)
		}
	}

	// (): on div should get undefined (empty in the Go string API).
	if got := Load(`<div>`).Find("div").Val(); got != "" {
		t.Errorf("val(div) = %q, want empty", got)
	}
	// on option should get value.
	if got := doc.Find("select#one option").Eq(0).Val(); got != "option_not_selected" {
		t.Errorf("val(first option) = %q, want %q", got, "option_not_selected")
	}
	// (value): on input text should set value.
	el := doc.Find(`input[type="text"]`).SetVal("test")
	if got := el.Val(); got != "test" {
		t.Errorf("setVal(test) -> %q, want %q", got, "test")
	}
	// Multiple-select first selected option value. Upstream returns the array
	// ['2','3']; the Go API is string-typed and returns the first selected
	// option's value, documented as a deliberate divergence.
	if got := Load(fxInputs).Find("select#multi").Val(); got != "2" {
		t.Errorf("val(multi) = %q, want %q (first selected)", got, "2")
	}
}

// TestParityData mirrors the .data block of attributes.spec.ts (string-level
// vectors; JavaScript type coercion to number/JSON/boolean is out of scope for
// the string-typed Go API and is not asserted).
func TestParityData(t *testing.T) {
	doc := Load(fxChocolates)

	// (): should get all data attributes initially declared in the markup.
	m := doc.Find(".linth").DataMap()
	if m["highlight"] != "Lindor" || m["origin"] != "swiss" || len(m) != 2 {
		t.Errorf("data() = %v, want {highlight:Lindor, origin:swiss}", m)
	}
	// (hyphen key): data attribute with hyphen should be camelized.
	fm := doc.Find(".frey").DataMap()
	if fm["taste"] != "sweet" || fm["bestCollection"] != "Mahony" {
		t.Errorf("data() frey = %v, want {taste:sweet, bestCollection:Mahony}", fm)
	}
	// (valid key): valid data attribute should get value.
	if v, _ := doc.Find(".linth").Data("highlight"); v != "Lindor" {
		t.Errorf("data(highlight) = %q, want %q", v, "Lindor")
	}
	// (invalid key): invalid data attribute should return undefined.
	if _, ok := doc.Find(".frey").Data("lol"); ok {
		t.Error("data(lol) should be absent")
	}
	// (): no data attribute should return an empty object.
	if got := doc.Find(".cailler").DataMap(); len(got) != 0 {
		t.Errorf("data() cailler = %v, want empty", got)
	}
	// (key): should translate camel-cased keys to hyphen-separated versions.
	el := Load(`<div data-foo-bar="a"></div>`).Find("div")
	if v, _ := el.Data("fooBar"); v != "a" {
		t.Errorf("data(fooBar) = %q, want %q", v, "a")
	}
	// (key, value): should set data attribute (readable via camelCase key).
	set := Load(`<div></div>`).Find("div").SetData("snack", "chocoletti")
	if v, _ := set.Data("snack"); v != "chocoletti" {
		t.Errorf("setData/data(snack) = %q, want %q", v, "chocoletti")
	}
}

// TestParityTraversing mirrors representative vectors from traversing.spec.ts.
// Parent chains use cheerio's fragment model (no synthesized <html>/<body>),
// which the Go port shares, so parent counts reflect the fragment tree.
func TestParityTraversing(t *testing.T) {
	doc := Load(fxFruits)

	// find: (single) should find one descendant.
	if got := doc.Find("#fruits").Find(".apple").AttrOr("class", ""); got != "apple" {
		t.Errorf("find(.apple) class = %q, want %q", got, "apple")
	}
	// find: (many) should find all matching descendants.
	if got := doc.Find("#fruits").Find("li").Length(); got != 3 {
		t.Errorf("find(li) len = %d, want 3", got)
	}
	// find: should query descendants only (no self match).
	if got := doc.Find("#fruits").Find("ul").Length(); got != 0 {
		t.Errorf("find(ul) len = %d, want 0", got)
	}
	// find: should lowercase selectors.
	if got := doc.Find("#fruits").Find("LI").Length(); got != 3 {
		t.Errorf("find(LI) len = %d, want 3", got)
	}

	// find: should query immediate descendant only ("> bar"). GAP.
	q := Load(`<foo><bar><bar></bar><bar></bar></bar></foo>`)
	if got := q.Find("foo").Find("> bar").Length(); got != 1 {
		t.Errorf("find('> bar') len = %d, want 1", got)
	}
	// find: should find siblings via leading combinators. GAP.
	p := Load(`<p class=a></p><p class=b></p>`)
	if got := p.Find(".a").Find("+.b").Length(); got != 1 {
		t.Errorf("find('+.b') len = %d, want 1", got)
	}
	if got := p.Find(".a").Find("~.b").Length(); got != 1 {
		t.Errorf("find('~.b') len = %d, want 1", got)
	}
	if got := p.Find(".a").Find("+.a").Length(); got != 0 {
		t.Errorf("find('+.a') len = %d, want 0", got)
	}
	if got := p.Find(".a").Find("~.a").Length(); got != 0 {
		t.Errorf("find('~.a') len = %d, want 0", got)
	}

	// children: () should get all children (elements only).
	if got := doc.Find("ul").Children().Length(); got != 3 {
		t.Errorf("children() len = %d, want 3", got)
	}
	// children: () should skip text nodes.
	if got := Load(fxMixedText).Root().Children().Length(); got != 2 {
		t.Errorf("children() of mixedText len = %d, want 2", got)
	}
	// children: (selector) should return children matching selector.
	if got := doc.Find("ul").Children(".orange").AttrOr("class", ""); got != "orange" {
		t.Errorf("children(.orange) class = %q, want %q", got, "orange")
	}

	// next: () should return next element.
	if got := doc.Find(".orange").Next().AttrOr("class", ""); got != "pear" {
		t.Errorf("next() class = %q, want %q", got, "pear")
	}
	// next: (no next) should return empty for last child.
	if got := doc.Find(".pear").Next().Length(); got != 0 {
		t.Errorf("next() of last len = %d, want 0", got)
	}
	// nextAll: () should return all following siblings.
	if got := doc.Find(".apple").NextAll().Length(); got != 2 {
		t.Errorf("nextAll() len = %d, want 2", got)
	}
	// nextAll: (selector) should filter.
	if got := doc.Find(".apple").NextAll(".pear").Length(); got != 1 {
		t.Errorf("nextAll(.pear) len = %d, want 1", got)
	}

	// siblings: () should get all siblings.
	if got := doc.Find(".orange").Siblings().Length(); got != 2 {
		t.Errorf("siblings() len = %d, want 2", got)
	}
	// siblings: (selector) should filter.
	if got := doc.Find(".orange").Siblings(".apple").Length(); got != 1 {
		t.Errorf("siblings(.apple) len = %d, want 1", got)
	}

	// parents: (selector) should get matching parents in logical order.
	food := Load(fxFood)
	uls := food.Find(".orange").Parents("ul")
	if uls.Length() != 2 {
		t.Fatalf("parents(ul) len = %d, want 2", uls.Length())
	}
	if got := uls.Eq(0).AttrOr("id", ""); got != "fruits" {
		t.Errorf("parents(ul)[0] id = %q, want %q", got, "fruits")
	}
	if got := uls.Eq(1).AttrOr("id", ""); got != "food" {
		t.Errorf("parents(ul)[1] id = %q, want %q", got, "food")
	}

	// parent: should return the direct parent.
	if got := food.Find(".orange").Parent().AttrOr("id", ""); got != "fruits" {
		t.Errorf("parent() id = %q, want %q", got, "fruits")
	}
	// closest: nearest ancestor-or-self matching the selector.
	if got := food.Find(".orange").Closest("ul").AttrOr("id", ""); got != "fruits" {
		t.Errorf("closest(ul) id = %q, want %q", got, "fruits")
	}
}

// TestParityFilterIsHasEq mirrors the .filter, .not, .is, .has, .eq, .first and
// .last blocks of traversing.spec.ts.
func TestParityFilterIsHasEq(t *testing.T) {
	doc := Load(fxFruits)

	if got := doc.Find("li").Filter(".orange").Length(); got != 1 {
		t.Errorf("filter(.orange) len = %d, want 1", got)
	}
	if got := doc.Find("li").Not(".apple").Length(); got != 2 {
		t.Errorf("not(.apple) len = %d, want 2", got)
	}
	if !doc.Find("ul").Is("#fruits") {
		t.Error("is(#fruits) = false, want true")
	}
	if doc.Find("ul").Is(".nope") {
		t.Error("is(.nope) = true, want false")
	}
	if got := doc.Find("ul").Has(".apple").Length(); got != 1 {
		t.Errorf("has(.apple) len = %d, want 1", got)
	}
	if got := doc.Find("li").Eq(1).Text(); got != "Orange" {
		t.Errorf("eq(1).text() = %q, want %q", got, "Orange")
	}
	if got := doc.Find("li").First().Text(); got != "Apple" {
		t.Errorf("first().text() = %q, want %q", got, "Apple")
	}
	if got := doc.Find("li").Last().Text(); got != "Pear" {
		t.Errorf("last().text() = %q, want %q", got, "Pear")
	}
}

// TestParityNthAndPseudos encodes structural pseudo-class vectors used across
// the upstream selector tests.
func TestParityNthAndPseudos(t *testing.T) {
	doc := Load(fxFruits)

	if got := doc.Find("li:nth-child(2)").Text(); got != "Orange" {
		t.Errorf("li:nth-child(2) = %q, want %q", got, "Orange")
	}
	if got := doc.Find("li:first-child").Text(); got != "Apple" {
		t.Errorf("li:first-child = %q, want %q", got, "Apple")
	}
	if got := doc.Find("li:last-child").Text(); got != "Pear" {
		t.Errorf("li:last-child = %q, want %q", got, "Pear")
	}
	// :nth-child(odd) selects the 1st and 3rd list items.
	if got := doc.Find("li:nth-child(odd)").Length(); got != 2 {
		t.Errorf("li:nth-child(odd) len = %d, want 2", got)
	}
}

// TestParityHtmlText mirrors static-rendering vectors: inner html, text and
// entity decoding.
func TestParityHtmlText(t *testing.T) {
	doc := Load(fxFruits)

	if got := doc.Find("#fruits").Text(); got != "AppleOrangePear" {
		t.Errorf("text() = %q, want %q", got, "AppleOrangePear")
	}
	wantHTML := `<li class="apple">Apple</li><li class="orange">Orange</li><li class="pear">Pear</li>`
	if got := doc.Find("#fruits").Html(); got != wantHTML {
		t.Errorf("html() = %q, want %q", got, wantHTML)
	}
	// Entity decoding in text content.
	if got := Load(`<p>a &amp; b &lt; c</p>`).Find("p").Text(); got != "a & b < c" {
		t.Errorf("text() entities = %q, want %q", got, "a & b < c")
	}
}
