package cheerio

// parser builds a DOM tree from a token stream using an open-element stack.
// It implements a pragmatic subset of the HTML tree-construction rules:
// void elements never receive children, raw-text elements capture their
// content verbatim, and certain elements implicitly close open siblings
// (e.g. <p>, <li>, table cells) so that common malformed markup parses the
// way a browser would.
type parser struct {
	tok   *tokenizer
	doc   *Node
	stack []*Node // open element stack; stack[0] is the document
}

// impliedEndTags maps a tag that, when opened, should implicitly close a set
// of currently-open elements before being inserted.
var closesOnOpen = map[string][]string{
	"li":       {"li"},
	"dt":       {"dt", "dd"},
	"dd":       {"dt", "dd"},
	"option":   {"option"},
	"optgroup": {"option", "optgroup"},
	"p":        {"p"},
	"tr":       {"tr", "td", "th"},
	"td":       {"td", "th"},
	"th":       {"td", "th"},
	"thead":    {"thead", "tbody", "tfoot"},
	"tbody":    {"thead", "tbody", "tfoot"},
	"tfoot":    {"thead", "tbody", "tfoot"},
}

// blockTagsThatClosePinline: opening a block-level element closes an open <p>.
var closesParagraph = map[string]bool{
	"address": true, "article": true, "aside": true, "blockquote": true,
	"div": true, "dl": true, "fieldset": true, "figure": true, "footer": true,
	"form": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true,
	"h6": true, "header": true, "hr": true, "main": true, "nav": true,
	"ol": true, "p": true, "pre": true, "section": true, "table": true,
	"ul": true,
}

// parse tokenizes and builds the document.
func parse(html string) *Node {
	doc := &Node{Type: DocumentNode}
	p := &parser{
		tok:   newTokenizer(html),
		doc:   doc,
		stack: []*Node{doc},
	}
	p.run()
	return doc
}

// current returns the top of the open-element stack.
func (p *parser) current() *Node {
	return p.stack[len(p.stack)-1]
}

func (p *parser) push(n *Node) {
	p.stack = append(p.stack, n)
}

func (p *parser) pop() {
	if len(p.stack) > 1 {
		p.stack = p.stack[:len(p.stack)-1]
	}
}

// run consumes the whole token stream.
func (p *parser) run() {
	for {
		tok, ok := p.tok.next()
		if !ok {
			return
		}
		switch tok.kind {
		case tokenText:
			p.handleText(tok.data)
		case tokenComment:
			p.current().AppendChild(&Node{Type: CommentNode, Data: tok.data})
		case tokenDoctype:
			p.doc.AppendChild(&Node{Type: DoctypeNode, Data: tok.data})
		case tokenStartTag:
			p.handleStartTag(tok)
		case tokenEndTag:
			p.handleEndTag(tok.data)
		}
	}
}

// handleText appends character data, decoding entities.
func (p *parser) handleText(data string) {
	if data == "" {
		return
	}
	p.current().AppendChild(&Node{Type: TextNode, Data: unescape(data)})
}

// handleStartTag performs implicit closes, inserts the element, and manages
// void and raw-text elements.
func (p *parser) handleStartTag(tok token) {
	name := tok.data

	// A <col> start tag is only meaningful inside a table/colgroup; the HTML5
	// tree builder ignores it anywhere else (it never reaches "in body").
	if name == "col" && !p.inColContext() {
		return
	}

	p.applyImplicitCloses(name)
	p.applyImpliedTableSection(name)

	el := &Node{Type: ElementNode, Data: name, Attr: tok.attr}

	// Foster parenting: a non-table element opened while the current node is a
	// <table> is moved out and inserted immediately before the table, rather than
	// nested inside it (HTML5 "in table" insertion mode).
	if p.current().Data == "table" && !tableScoped[name] {
		p.fosterParent(el)
	} else {
		p.current().AppendChild(el)
	}

	// The self-closing slash is only honoured for void elements; on ordinary HTML
	// elements ("<span/>") it is ignored, so the element stays open.
	if voidElements[name] {
		return
	}

	if rawTextElements[name] {
		text := p.tok.readRawText(name)
		if text != "" {
			// title/textarea content is treated as decoded text; script and
			// style are kept verbatim.
			if name == "title" || name == "textarea" {
				el.AppendChild(&Node{Type: TextNode, Data: unescape(text)})
			} else {
				el.AppendChild(&Node{Type: TextNode, Data: text})
			}
		}
		return
	}

	p.push(el)
}

// tableScoped is the set of elements that legitimately live directly inside a
// <table> and therefore are not foster-parented out of it.
var tableScoped = map[string]bool{
	"caption": true, "colgroup": true, "col": true, "thead": true,
	"tbody": true, "tfoot": true, "tr": true, "td": true, "th": true,
	"script": true, "style": true, "template": true, "form": true,
}

// inColContext reports whether the open-element stack currently has a table or
// colgroup, the only contexts in which a <col> start tag is retained.
func (p *parser) inColContext() bool {
	for i := len(p.stack) - 1; i >= 1; i-- {
		switch p.stack[i].Data {
		case "table", "colgroup":
			return true
		}
	}
	return false
}

// fosterParent inserts el immediately before the current <table> in the table's
// parent, implementing the HTML5 foster-parenting rule for stray content in
// table context. el's Parent is updated accordingly.
func (p *parser) fosterParent(el *Node) {
	table := p.current()
	parent := table.Parent
	if parent == nil {
		table.AppendChild(el)
		return
	}
	idx := len(parent.Children)
	for i, c := range parent.Children {
		if c == table {
			idx = i
			break
		}
	}
	parent.Children = append(parent.Children, nil)
	copy(parent.Children[idx+1:], parent.Children[idx:])
	parent.Children[idx] = el
	el.Parent = parent
}

// applyImpliedTableSection inserts an implicit <tbody> when a <tr> is opened
// directly inside a <table>, mirroring the HTML tree-construction rule that
// gives table rows an implied section parent. A <td>/<th> opened directly in a
// <table> likewise gets an implied <tbody><tr>.
func (p *parser) applyImpliedTableSection(name string) {
	switch name {
	case "tr":
		if p.current().Data == "table" {
			p.openImplied("tbody")
		}
	case "td", "th":
		if p.current().Data == "table" {
			p.openImplied("tbody")
			p.openImplied("tr")
		} else if p.current().Data == "tbody" || p.current().Data == "thead" || p.current().Data == "tfoot" {
			p.openImplied("tr")
		}
	}
}

// openImplied creates and pushes a synthetic element with the given name.
func (p *parser) openImplied(name string) {
	el := &Node{Type: ElementNode, Data: name}
	p.current().AppendChild(el)
	p.push(el)
}

// applyImplicitCloses pops open elements that the newly-opened tag closes.
func (p *parser) applyImplicitCloses(name string) {
	if targets, ok := closesOnOpen[name]; ok {
		for setContains(targets, p.current().Data) {
			p.pop()
		}
	}
	if closesParagraph[name] && p.hasOpen("p") {
		// Close the open <p> if it is the current node.
		if p.current().Data == "p" {
			p.pop()
		}
	}
}

// handleEndTag pops the stack up to and including the matching open element.
// Unmatched end tags are ignored.
func (p *parser) handleEndTag(name string) {
	if voidElements[name] {
		return
	}
	if !p.hasOpen(name) {
		// A </p> with no <p> open synthesises an empty <p></p> (HTML5 "act as if
		// a <p> start tag had been seen"), but only once the current node already
		// contains element content — matching parse5's observed behaviour, where a
		// stray </p> after only text (or nothing) is dropped.
		if name == "p" && p.currentHasElementChild() {
			p.current().AppendChild(&Node{Type: ElementNode, Data: "p"})
		}
		return
	}
	for len(p.stack) > 1 {
		n := p.current()
		p.pop()
		if n.Data == name {
			return
		}
	}
}

// currentHasElementChild reports whether the current node already has at least
// one element child.
func (p *parser) currentHasElementChild() bool {
	for _, c := range p.current().Children {
		if c.Type == ElementNode {
			return true
		}
	}
	return false
}

// hasOpen reports whether an element with the given tag is on the open stack.
func (p *parser) hasOpen(name string) bool {
	for i := len(p.stack) - 1; i >= 1; i-- {
		if p.stack[i].Data == name {
			return true
		}
	}
	return false
}

func setContains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}
