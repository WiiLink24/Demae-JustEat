package demae

import (
	"encoding/base64"
	"encoding/hex"
	"math/rand"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/mitchellh/go-wordwrap"
)

// BoolToInt converts a boolean value to an integer.
// This is needed because Nintendo wants the integer, not the string literal.
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func RemoveInvalidCharacters(input string) string {
	result := make([]rune, 0, len(input))

	for i, r := range input {
		// Keep only printable ASCII and some specific Unicode ranges
		if r < 128 || unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsPunct(r) || unicode.IsSpace(r) {
			result = append(result, r)
		} else {
			// If it was invalid and the previous char was a space, we want to pop it.
			if i == 0 {
				continue
			}

			if result[len(result)-1] == ' ' {
				result = result[:len(result)-1]
			}
		}
	}

	return string(result)
}

func IDGenerator(size int, chars string) string {
	sb := strings.Builder{}
	sb.Grow(size) // Pre-allocate for efficiency

	for i := 0; i < size; i++ {
		randomIndex := rand.Intn(len(chars))
		sb.WriteByte(chars[randomIndex])
	}

	return sb.String()
}

func FloatToString(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func FormatDecimal(f float64, precision int) string {
	formatted := strconv.FormatFloat(f, 'f', precision, 64)
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")
	return formatted
}

func UUID() string {
	u, _ := uuid.NewUUID()
	return u.String()
}

func Wordwrap(text string, width uint, maxLines int) string {
	wrapped := wordwrap.WrapString(text, width)
	if maxLines == -1 {
		return wrapped
	}

	strippedWrapped := ""
	for i, s := range strings.Split(wrapped, "\n") {
		if i == maxLines {
			break
		}

		strippedWrapped += s + "\n"
	}

	// Remove last newline
	return strings.TrimRight(strippedWrapped, "\n")
}

func CompressUUID(uuid string) string {
	hexStr := strings.ReplaceAll(uuid, "-", "")
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return ""
	}

	encoded := base64.URLEncoding.EncodeToString(bytes)

	// Remove padding
	shortened := strings.TrimRight(encoded, "=")

	return shortened
}
