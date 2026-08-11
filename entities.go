package cheerio

import (
	"strconv"
	"strings"
)

// namedEntities maps the common HTML named character references to their
// replacement text. This is intentionally a pragmatic subset covering the
// entities seen in ordinary documents rather than the full WHATWG table.
var namedEntities = map[string]string{
	"amp":    "&",
	"lt":     "<",
	"gt":     ">",
	"quot":   "\"",
	"apos":   "'",
	"nbsp":   " ",
	"copy":   "©",
	"reg":    "®",
	"trade":  "™",
	"hellip": "…",
	"mdash":  "—",
	"ndash":  "–",
	"lsquo":  "‘",
	"rsquo":  "’",
	"ldquo":  "“",
	"rdquo":  "”",
	"times":  "×",
	"divide": "÷",
	"deg":    "°",
	"euro":   "€",
	"pound":  "£",
	"cent":   "¢",
	"yen":    "¥",
	"middot": "·",
	"laquo":  "«",
	"raquo":  "»",
	"eacute": "é",
	"egrave": "è",
	"agrave": "à",
	"ccedil": "ç",
	"uuml":   "ü",
	"ouml":   "ö",
	"not":    "¬",
	"para":   "¶",
	"sect":   "§",
	"iexcl":  "¡",
	"iquest": "¿",
	"micro":  "µ",
	"plusmn": "±",
	"sup2":   "²",
	"sup3":   "³",
	"frac12": "½",
	"frac14": "¼",
	"frac34": "¾",
	"szlig":  "ß",
	"aacute": "á",
	"iacute": "í",
	"oacute": "ó",
	"uacute": "ú",
	"ntilde": "ñ",
}

// legacyEntities are the named character references that the HTML5 tokenizer
// still decodes when they are NOT followed by a semicolon (the "without a
// semicolon" column of the named-character-reference table). Everything else
// requires the trailing ';'. This set lets "&amp" decode to "&" and lets a
// longest-prefix match such as "&notanentity;" resolve "&not" to "¬" and leave
// "anentity;" literal, matching the spec's tokenizer behaviour.
var legacyEntities = map[string]bool{
	"amp": true, "lt": true, "gt": true, "quot": true, "nbsp": true,
	"copy": true, "reg": true, "not": true, "para": true, "sect": true,
	"deg": true, "middot": true, "laquo": true, "raquo": true, "iexcl": true,
	"iquest": true, "micro": true, "plusmn": true, "sup2": true, "sup3": true,
	"frac12": true, "frac14": true, "frac34": true, "szlig": true,
	"times": true, "divide": true, "cent": true, "pound": true, "yen": true,
	"eacute": true, "egrave": true, "agrave": true, "ccedil": true,
	"uuml": true, "ouml": true, "aacute": true, "iacute": true,
	"oacute": true, "uacute": true, "ntilde": true,
}

// maxEntityNameLen is the longest key in namedEntities; the decoder never needs
// to scan more than this many characters looking for a named reference.
var maxEntityNameLen = func() int {
	m := 0
	for k := range namedEntities {
		if len(k) > m {
			m = len(k)
		}
	}
	return m
}()

// unescape decodes HTML character references (named, decimal &#NN; and
// hexadecimal &#xNN;) found in s, following the HTML5 tokenizer rules closely
// enough for real-world documents:
//
//   - numeric references (&#38; / &#x26;) decode with or without a trailing ';';
//   - a named reference decodes by longest prefix — "&notanentity;" resolves the
//     leading "&not" to "¬" and leaves "anentity;" literal;
//   - a name that is not terminated by ';' decodes only if it is a legacy
//     (semicolon-optional) entity, so "&amp" becomes "&" but "&copyx" does not.
//
// Unrecognized references are left intact.
func unescape(s string) string {
	if !strings.ContainsRune(s, '&') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] != '&' {
			b.WriteByte(s[i])
			i++
			continue
		}
		if decoded, adv, ok := decodeEntityAt(s, i); ok {
			b.WriteString(decoded)
			i += adv
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// decodeEntityAt attempts to decode a single character reference beginning at
// s[i] (which must be '&'). It returns the decoded text, the number of input
// bytes consumed (including the '&'), and whether a reference was recognised.
func decodeEntityAt(s string, i int) (string, int, bool) {
	// Numeric reference: &#DD; or &#xHH; (semicolon optional).
	if i+1 < len(s) && s[i+1] == '#' {
		j := i + 2
		hex := false
		if j < len(s) && (s[j] == 'x' || s[j] == 'X') {
			hex = true
			j++
		}
		start := j
		for j < len(s) && isHexOrDigit(s[j], hex) {
			j++
		}
		if j == start {
			return "", 0, false
		}
		var n int64
		var err error
		if hex {
			n, err = strconv.ParseInt(s[start:j], 16, 32)
		} else {
			n, err = strconv.ParseInt(s[start:j], 10, 32)
		}
		if err != nil || n < 0 {
			return "", 0, false
		}
		if j < len(s) && s[j] == ';' {
			j++
		}
		return string(rune(n)), j - i, true
	}

	// Named reference: try the longest prefix of the following letters/digits
	// that names a known entity.
	j := i + 1
	limit := i + 1 + maxEntityNameLen
	if limit > len(s) {
		limit = len(s)
	}
	for j < limit && isEntityNameChar(s[j]) {
		j++
	}
	for end := j; end > i+1; end-- {
		name := s[i+1 : end]
		v, ok := namedEntities[name]
		if !ok {
			continue
		}
		if end < len(s) && s[end] == ';' {
			return v, end - i + 1, true
		}
		if legacyEntities[name] {
			return v, end - i, true
		}
	}
	return "", 0, false
}

func isHexOrDigit(b byte, hex bool) bool {
	if b >= '0' && b <= '9' {
		return true
	}
	if hex {
		return (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
	}
	return false
}

func isEntityNameChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

// escapeText escapes text content for serialization. It mirrors the entities
// library used by cheerio's dom-serializer: only &, <, > and the non-breaking
// space are escaped; the double quote is left raw in text content.
func escapeText(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\u00A0", "&nbsp;")
	return r.Replace(s)
}

// escapeAttr escapes an attribute value for serialization inside double quotes.
// It mirrors cheerio/dom-serializer: & and " and the non-breaking space are
// escaped, but < and > are emitted raw (an attribute value never terminates a
// tag, so browsers do not require them to be escaped and upstream leaves them
// untouched).
func escapeAttr(s string) string {
	r := strings.NewReplacer("&", "&amp;", "\"", "&quot;", "\u00A0", "&nbsp;")
	return r.Replace(s)
}
