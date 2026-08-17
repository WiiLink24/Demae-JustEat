package demae

import (
	"encoding/xml"
	"unicode/utf8"
)

// MaxFieldBytes is the longest value the channel will read out of an XML node.
// Up to 127 bytes plus the terminator uses a 128 byte stack buffer, anything
// longer goes through a heap alloc, it never null checks and it fails with an
// error screen. Tested on Wii hardware, 127 renders and 128 doesn't.
//
// See PR for the disassembly
const MaxFieldBytes = 127

// Show ellipse for better UX. "…" is 3 bytes in UTF-8, same as "...": 127 - 3 = 124 bytes left for content.
const ellipsis = "…"

// TruncateField cuts s down to MaxFieldBytes, on a rune boundary.
func TruncateField(s string) string {
	if len(s) <= MaxFieldBytes {
		return s
	}

	cut := MaxFieldBytes - len(ellipsis)
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}

	return s[:cut] + ellipsis
}

// MarshalXML writes the value as CDATA and truncates it on the way out.
func (c CDATA) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	value := c.Value
	if s, ok := value.(string); ok {
		value = TruncateField(s)
	}

	type cdata struct {
		Value any `xml:",cdata"`
	}

	return e.EncodeElement(cdata{Value: value}, start)
}
