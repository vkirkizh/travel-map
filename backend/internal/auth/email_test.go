package auth

import "testing"

func TestNormalizeEmail(t *testing.T) {
	got := NormalizeEmail("  User.Name@Example.COM \t")
	want := "user.name@example.com"

	if got != want {
		t.Errorf("NormalizeEmail() = %q, want %q", got, want)
	}
}
