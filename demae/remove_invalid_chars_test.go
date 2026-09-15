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

func TestRemoveInvalidCharactersLeadingEmoji(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"two leading emoji", "🍕🍕 Pizza", " Pizza"},
		{"one leading emoji", "🍕 Pizza", " Pizza"},
		{"only emoji", "🍕🍕🍕", ""},
		{"emoji then text", "🍕Pizzeria", "Pizzeria"},
		{"emoji in the middle", "Pizza 🍕 Napoli", "Pizza Napoli"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveInvalidCharacters(tt.in); got != tt.want {
				t.Errorf("Expected %q but got %q", tt.want, got)
			}
		})
	}
}

func TestRemoveInvalidCharacters_Subscript(t *testing.T) {
	got := RemoveInvalidCharacters("Getränke mit H₂CO₃")
	want := "Getränke mit H2CO3"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
