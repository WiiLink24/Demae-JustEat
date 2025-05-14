package demae

import (
	"testing"
)

func TestRemoveInvalidCharacters(t *testing.T) {
	str := "Hot Wings Meal: 6 pc 🔥"

	if RemoveInvalidCharacters(str) != "Hot Wings Meal: 6 pc" {
		t.Error("incorrect result")
	}
}
