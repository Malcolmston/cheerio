package cheerio

import "testing"

func TestCompileErrors(t *testing.T) {
	valid := []string{"div", "ul > li.a", "a[href^=/], span", "*|title", ":nth-child(2n+1)"}
	for _, sel := range valid {
		if _, err := Compile(sel); err != nil {
			t.Errorf("Compile(%q) unexpected error: %v", sel, err)
		}
	}
	// An empty or whitespace-only selector is valid (it matches nothing), matching
	// cheerio's $("") — only genuinely malformed selectors error.
	empty := []string{"", "   "}
	for _, sel := range empty {
		if _, err := Compile(sel); err != nil {
			t.Errorf("Compile(%q) unexpected error: %v", sel, err)
		}
	}
	invalid := []string{"a,", ",b", "[", ":totally-unknown"}
	for _, sel := range invalid {
		if _, err := Compile(sel); err == nil {
			t.Errorf("Compile(%q) expected error, got nil", sel)
		}
	}
}

func TestMustCompilePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("MustCompile(bad) did not panic")
		}
	}()
	MustCompile("[")
}

func TestMatcherMatchAndString(t *testing.T) {
	doc := Load(`<div><p class="a">x</p><p class="b">y</p></div>`)
	m := MustCompile("p.a")
	if m.String() != "p.a" {
		t.Errorf("String() = %q", m.String())
	}
	pa := doc.Find(".a").Get(0)
	pb := doc.Find(".b").Get(0)
	if !m.Match(pa) || m.Match(pb) {
		t.Errorf("Match wrong")
	}
}

func TestDocumentFindMatcher(t *testing.T) {
	doc := Load(`<ul><li class="x">1</li><li>2</li><li class="x">3</li></ul>`)
	m := MustCompile("li.x")
	got := doc.FindMatcher(m)
	if got.Length() != 2 {
		t.Fatalf("FindMatcher len = %d, want 2", got.Length())
	}
	if got.First().Text() != "1" || got.Last().Text() != "3" {
		t.Errorf("FindMatcher content wrong: %q..%q", got.First().Text(), got.Last().Text())
	}
	// Parity with the string form.
	if doc.FindMatcher(m).Length() != doc.Find("li.x").Length() {
		t.Errorf("FindMatcher disagrees with Find")
	}
}

func TestSelectionMatcherMethods(t *testing.T) {
	doc := Load(`<div><span class="a">1</span><span>2</span><b class="a">3</b></div>`)
	root := doc.Find("div")
	spanA := MustCompile("span.a")

	if root.FindMatcher(spanA).Length() != 1 {
		t.Errorf("Selection.FindMatcher len wrong")
	}

	allA := MustCompile(".a")
	filtered := doc.Find("span, b").FilterMatcher(allA)
	if filtered.Length() != 2 {
		t.Errorf("FilterMatcher len = %d, want 2", filtered.Length())
	}
	if !doc.Find("span").IsMatcher(allA) {
		t.Errorf("IsMatcher should be true")
	}
	if doc.Find("span").IsMatcher(MustCompile(".zzz")) {
		t.Errorf("IsMatcher should be false")
	}
}

func BenchmarkFindMatcherVsFind(b *testing.B) {
	doc := Load(`<div><ul><li class="x">a</li><li>b</li><li class="x">c</li></ul>` +
		`<section><p class="x">d</p></section></div>`)
	m := MustCompile("li.x, p.x")
	b.Run("compiled", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = doc.FindMatcher(m)
		}
	})
	b.Run("string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = doc.Find("li.x, p.x")
		}
	})
}
