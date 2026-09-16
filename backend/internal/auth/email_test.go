package auth

import "testing"

func TestNormalizeEmail(t *testing.T) {
	got := NormalizeEmail("  User.Name@Example.COM \t")
	want := "user.name@example.com"

	if got != want {
		t.Errorf("NormalizeEmail() = %q, want %q", got, want)
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{name: "bare email", email: "user@example.com", want: true},
		{name: "display name", email: "Valery <user@example.com>", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsValidEmail(test.email); got != test.want {
				t.Errorf("IsValidEmail(%q) = %t, want %t", test.email, got, test.want)
			}
		})
	}
}
