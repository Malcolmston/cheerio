package cheerio_test

import (
	"fmt"

	"github.com/malcolmston/cheerio"
)

func Example() {
	const html = `
	<html><body>
	  <ul id="fruit">
	    <li class="a">Apple</li>
	    <li class="b">Banana</li>
	    <li class="c">Cherry</li>
	  </ul>
	  <a href="https://example.com/go">Go</a>
	</body></html>`

	doc := cheerio.Load(html)

	// Text of the first list item.
	fmt.Println(doc.Find("#fruit li").First().Text())

	// Count of list items.
	fmt.Println(doc.Find("li").Length())

	// Attribute access.
	href, _ := doc.Find("a").Attr("href")
	fmt.Println(href)

	// Selecting by class and combinator.
	fmt.Println(doc.Find("li.b + li").Text())

	// Output:
	// Apple
	// 3
	// https://example.com/go
	// Cherry
}

func ExampleSelection_Map() {
	doc := cheerio.Load(`<ol><li>one</li><li>two</li><li>three</li></ol>`)
	items := doc.Find("li").Map(func(i int, n *cheerio.Node) string {
		return fmt.Sprintf("%d:%s", i, n.Children[0].Data)
	})
	fmt.Println(items)
	// Output:
	// [0:one 1:two 2:three]
}

func ExampleSelection_Append() {
	doc := cheerio.Load(`<ul><li>a</li></ul>`)
	doc.Find("ul").Append(`<li>b</li>`).Prepend(`<li>z</li>`)
	fmt.Println(doc.Find("ul").Html())
	// Output:
	// <li>z</li><li>a</li><li>b</li>
}

func ExampleSelection_Wrap() {
	doc := cheerio.Load(`<div><span>x</span></div>`)
	doc.Find("span").Wrap(`<b class="hl"></b>`)
	fmt.Println(doc.Find("div").Html())
	// Output:
	// <b class="hl"><span>x</span></b>
}

func ExampleSelection_SerializeArray() {
	doc := cheerio.Load(`<form>
		<input type="text" name="user" value="ada">
		<input type="checkbox" name="admin" value="yes" checked>
	</form>`)
	fmt.Println(doc.Find("form").Serialize())
	// Output:
	// user=ada&admin=yes
}
