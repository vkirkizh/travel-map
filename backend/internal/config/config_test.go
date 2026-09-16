package config

import "testing"

func TestLoadTrimsRegistrationInviteCode(t *testing.T) {
	t.Setenv("REGISTRATION_INVITE_CODE", "  my-code\t")

	cfg := Load()

	if cfg.RegistrationInviteCode != "my-code" {
		t.Errorf("RegistrationInviteCode = %q, want %q", cfg.RegistrationInviteCode, "my-code")
	}
}

func TestLoadConfiguresSecureCookiesFromAppEnv(t *testing.T) {
	tests := []struct {
		name   string
		appEnv string
		want   bool
	}{
		{name: "production", appEnv: "production", want: true},
		{name: "local", appEnv: "local", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("APP_ENV", test.appEnv)

			cfg := Load()

			if cfg.SecureCookies != test.want {
				t.Errorf("SecureCookies = %t, want %t", cfg.SecureCookies, test.want)
			}
		})
	}
}
