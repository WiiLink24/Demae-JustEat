package demae

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/mitchellh/go-wordwrap"
	"math/rand"
	"strconv"
	"strings"
	"unicode"
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

func UUID() string {
	u, _ := uuid.DefaultGenerator.NewV4()
	return u.String()
}

func RandIntWRange(min, max int) int {
	return min + int(rand.Int63n(int64(max-min+1)))
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

func DecompressUUID(shortened string) string {
	// Step 1: Add padding back (Base64 needs length %4 == 0)
	padding := len(shortened) % 4
	if padding > 0 {
		shortened += strings.Repeat("=", 4-padding)
	}

	// Step 2: Decode Base64 URL-safe to bytes
	bytes, err := base64.URLEncoding.DecodeString(shortened)
	if err != nil {
		return ""
	}

	// Step 3: Convert bytes back to hex string
	hexStr := hex.EncodeToString(bytes)

	// Step 4: Reformat into UUID (add hyphens)
	if len(hexStr) != 32 {
		return ""
	}
	uuid := fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		hexStr[:8],
		hexStr[8:12],
		hexStr[12:16],
		hexStr[16:20],
		hexStr[20:],
	)

	return uuid
}
