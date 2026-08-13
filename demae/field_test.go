package demae

import (
	"encoding/xml"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateField(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"under the limit", strings.Repeat("a", 100), 100},
		{"exactly at the limit", strings.Repeat("a", MaxFieldBytes), MaxFieldBytes},
		{"one over", strings.Repeat("a", MaxFieldBytes+1), MaxFieldBytes},
		{"empty", "", 0},
		{"umlauts", strings.Repeat("ä", 100), MaxFieldBytes - 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateField(tt.in)

			if len(got) != tt.want {
				t.Errorf("Expected %d bytes but got %d", tt.want, len(got))
			}

			if !utf8.ValidString(got) {
				t.Errorf("Expected valid UTF-8 but got %q", got)
			}
		})
	}
}

func TestCDATAMarshal(t *testing.T) {
	type field struct {
		XMLName xml.Name `xml:"shop"`
		Value   CDATA    `xml:"name"`
	}

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"short string", "Call a Pizza", "<![CDATA[Call a Pizza]]>"},
		{"integer", 1500, "<![CDATA[1500]]>"},
		{"float", 12.5, "<![CDATA[12.5]]>"},
		{"empty", "", "<name></name>"},
		{"too long", strings.Repeat("x", 500), "<![CDATA[" + strings.Repeat("x", MaxFieldBytes) + "]]>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := xml.Marshal(field{Value: CDATA{Value: tt.value}})
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			if !strings.Contains(string(out), tt.want) {
				t.Errorf("Expected %s but got %s", tt.want, out)
			}
		})
	}
}
