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
}

// unescape decodes HTML character references (named, decimal &#NN; and
// hexadecimal &#xNN;) found in s. Unrecognized references are left intact.
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
		semi := strings.IndexByte(s[i:], ';')
		if semi < 0 || semi > 32 {
			b.WriteByte(s[i])
			i++
			continue
		}
		ref := s[i+1 : i+semi]
		if decoded, ok := decodeRef(ref); ok {
			b.WriteString(decoded)
			i += semi + 1
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// decodeRef decodes a single reference body (the text between '&' and ';').
func decodeRef(ref string) (string, bool) {
	if ref == "" {
		return "", false
	}
	if ref[0] == '#' {
		var n int64
		var err error
		if len(ref) > 1 && (ref[1] == 'x' || ref[1] == 'X') {
			n, err = strconv.ParseInt(ref[2:], 16, 32)
		} else {
			n, err = strconv.ParseInt(ref[1:], 10, 32)
		}
		if err != nil || n < 0 {
			return "", false
		}
		return string(rune(n)), true
	}
	if v, ok := namedEntities[ref]; ok {
		return v, true
	}
	return "", false
}

// escapeText escapes text content for serialization (& and < and >).
func escapeText(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

// escapeAttr escapes an attribute value for serialization inside double
// quotes.
func escapeAttr(s string) string {
	r := strings.NewReplacer("&", "&amp;", "\"", "&quot;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}
