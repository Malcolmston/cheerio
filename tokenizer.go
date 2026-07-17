package cheerio

import "strings"

// tokenKind enumerates the token categories emitted by the tokenizer.
type tokenKind int

const (
	tokenStartTag tokenKind = iota
	tokenEndTag
	tokenText
	tokenComment
	tokenDoctype
)

// token is a single lexical unit produced by the tokenizer.
type token struct {
	kind        tokenKind
	data        string      // tag name (lower-cased) or text/comment/doctype content
	attr        []Attribute // start-tag attributes in source order
	selfClosing bool        // start tag written as <.../>
}

// tokenizer converts a raw HTML string into a stream of tokens. It is a
// hand-written scanner that tolerates a good deal of malformed input.
type tokenizer struct {
	s   string
	pos int
}

func newTokenizer(s string) *tokenizer {
	return &tokenizer{s: s}
}

// next returns the next token and true, or a zero token and false at EOF.
func (t *tokenizer) next() (token, bool) {
	if t.pos >= len(t.s) {
		return token{}, false
	}
	if t.s[t.pos] == '<' {
		if tok, ok := t.readTag(); ok {
			return tok, true
		}
		// Not a valid tag opener; treat '<' as literal text.
	}
	return t.readText(), true
}

// readText consumes characters up to the next '<' (or EOF). When the current
// character is a '<' that did not form a valid tag, it is consumed as a
// literal so the scanner always makes progress.
func (t *tokenizer) readText() token {
	start := t.pos
	if t.pos < len(t.s) && t.s[t.pos] == '<' {
		t.pos++ // consume the stray '<' as literal text
	}
	for t.pos < len(t.s) && t.s[t.pos] != '<' {
		t.pos++
	}
	return token{kind: tokenText, data: t.s[start:t.pos]}
}

// readTag attempts to parse a tag, comment or doctype starting at '<'. If the
// input does not form a recognizable construct it returns ok=false and leaves
// the position unchanged so the caller can emit literal text.
func (t *tokenizer) readTag() (token, bool) {
	rest := t.s[t.pos:]

	// Comment.
	if strings.HasPrefix(rest, "<!--") {
		end := strings.Index(t.s[t.pos+4:], "-->")
		if end < 0 {
			// Unterminated comment: consume to EOF.
			data := t.s[t.pos+4:]
			t.pos = len(t.s)
			return token{kind: tokenComment, data: data}, true
		}
		data := t.s[t.pos+4 : t.pos+4+end]
		t.pos = t.pos + 4 + end + 3
		return token{kind: tokenComment, data: data}, true
	}

	// Doctype (or other bogus <! declaration).
	if strings.HasPrefix(rest, "<!") {
		end := strings.IndexByte(rest, '>')
		if end < 0 {
			data := rest[2:]
			t.pos = len(t.s)
			return token{kind: tokenDoctype, data: strings.TrimSpace(data)}, true
		}
		data := rest[2:end]
		t.pos += end + 1
		return token{kind: tokenDoctype, data: strings.TrimSpace(data)}, true
	}

	// End tag.
	if strings.HasPrefix(rest, "</") {
		i := 2
		for i < len(rest) && rest[i] != '>' {
			i++
		}
		if i >= len(rest) {
			return token{}, false
		}
		name := strings.TrimSpace(rest[2:i])
		t.pos += i + 1
		return token{kind: tokenEndTag, data: strings.ToLower(name)}, true
	}

	// Start tag: must begin with a letter.
	if len(rest) < 2 || !isTagNameStart(rest[1]) {
		return token{}, false
	}
	return t.readStartTag()
}

// readStartTag parses "<name attr... >" including quoted, unquoted and
// valueless attributes.
func (t *tokenizer) readStartTag() (token, bool) {
	i := 1 // skip '<'
	rest := t.s[t.pos:]
	nameStart := i
	for i < len(rest) && isTagNameChar(rest[i]) {
		i++
	}
	name := strings.ToLower(rest[nameStart:i])

	var attrs []Attribute
	selfClosing := false

	for i < len(rest) {
		// Skip whitespace between attributes.
		for i < len(rest) && isSpace(rest[i]) {
			i++
		}
		if i >= len(rest) {
			// Unterminated tag; consume the rest.
			t.pos = len(t.s)
			return token{kind: tokenStartTag, data: name, attr: attrs}, true
		}
		if rest[i] == '>' {
			i++
			break
		}
		if rest[i] == '/' {
			selfClosing = true
			i++
			continue
		}

		// Attribute name.
		attrStart := i
		for i < len(rest) && !isSpace(rest[i]) && rest[i] != '=' && rest[i] != '>' && rest[i] != '/' {
			i++
		}
		attrName := strings.ToLower(rest[attrStart:i])
		if attrName == "" {
			// Stray character; skip it to make progress.
			i++
			continue
		}

		// Optional value.
		for i < len(rest) && isSpace(rest[i]) {
			i++
		}
		attrVal := ""
		if i < len(rest) && rest[i] == '=' {
			i++
			for i < len(rest) && isSpace(rest[i]) {
				i++
			}
			if i < len(rest) && (rest[i] == '"' || rest[i] == '\'') {
				quote := rest[i]
				i++
				valStart := i
				for i < len(rest) && rest[i] != quote {
					i++
				}
				attrVal = rest[valStart:i]
				if i < len(rest) {
					i++ // consume closing quote
				}
			} else {
				valStart := i
				for i < len(rest) && !isSpace(rest[i]) && rest[i] != '>' {
					i++
				}
				attrVal = rest[valStart:i]
			}
		}
		attrs = appendAttr(attrs, attrName, unescape(attrVal))
	}

	t.pos += i
	tok := token{kind: tokenStartTag, data: name, attr: attrs, selfClosing: selfClosing}
	return tok, true
}

// readRawText reads the raw content of a raw-text element up to its closing
// tag (case-insensitive) and returns the text plus whether a close was found.
func (t *tokenizer) readRawText(name string) string {
	closeTag := "</" + name
	lower := strings.ToLower(t.s[t.pos:])
	idx := strings.Index(lower, closeTag)
	if idx < 0 {
		text := t.s[t.pos:]
		t.pos = len(t.s)
		return text
	}
	text := t.s[t.pos : t.pos+idx]
	// Advance past the closing tag's '>'.
	rest := t.s[t.pos+idx:]
	gt := strings.IndexByte(rest, '>')
	if gt < 0 {
		t.pos = len(t.s)
	} else {
		t.pos = t.pos + idx + gt + 1
	}
	return text
}

// appendAttr adds an attribute unless the key already exists (first wins,
// matching browser behavior for duplicate attributes).
func appendAttr(attrs []Attribute, key, val string) []Attribute {
	for _, a := range attrs {
		if a.Key == key {
			return attrs
		}
	}
	return append(attrs, Attribute{Key: key, Val: val})
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f'
}

func isTagNameStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isTagNameChar(b byte) bool {
	return isTagNameStart(b) || (b >= '0' && b <= '9') || b == '-' || b == ':' || b == '_'
}
