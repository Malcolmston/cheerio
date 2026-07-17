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
