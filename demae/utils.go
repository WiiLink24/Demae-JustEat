package demae

import (
	"github.com/gofrs/uuid"
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

	for _, r := range input {
		// Keep only printable ASCII and some specific Unicode ranges
		if r < 128 || unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsPunct(r) || unicode.IsSpace(r) {
			result = append(result, r)
		} else {
			// If it was invalid and the previous char was a space, we want to pop it.
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
