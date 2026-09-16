package auth

import "testing"

func TestNormalizeUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     string
	}{
		{name: "uppercase", username: "Valery", want: "valery"},
		{name: "surrounding whitespace", username: " Valery_Kirkizh ", want: "valery_kirkizh"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NormalizeUsername(test.username); got != test.want {
				t.Errorf("NormalizeUsername(%q) = %q, want %q", test.username, got, test.want)
			}
		})
	}
}

func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{name: "minimum length", username: "vale", want: true},
		{name: "maximum length", username: "abcdefghijklmnopqrst", want: true},
		{name: "letters", username: "valery", want: true},
		{name: "underscore", username: "valery_kirkizh", want: true},
		{name: "digits", username: "valery1990", want: true},
		{name: "letter followed by digits", username: "v123", want: true},
		{name: "too short", username: "val", want: false},
		{name: "too long", username: "abcdefghijklmnopqrstu", want: false},
		{name: "starts with digit", username: "123valery", want: false},
		{name: "starts with underscore", username: "_valery", want: false},
		{name: "ends with underscore", username: "valery_", want: false},
		{name: "consecutive underscores", username: "valery__kirkizh", want: false},
		{name: "hyphen", username: "valery-kirkizh", want: false},
		{name: "period", username: "valery.kirkizh", want: false},
		{name: "space", username: "valery kirkizh", want: false},
		{name: "non-ASCII letters", username: "валерий", want: false},
		{name: "uppercase without normalization", username: "Valery", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsValidUsername(test.username); got != test.want {
				t.Errorf("IsValidUsername(%q) = %t, want %t", test.username, got, test.want)
			}
		})
	}
}
