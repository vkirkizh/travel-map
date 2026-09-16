package config

import "testing"

func TestLoadTrimsRegistrationInviteCode(t *testing.T) {
	t.Setenv("REGISTRATION_INVITE_CODE", "  my-code\t")

	cfg := Load()

	if cfg.RegistrationInviteCode != "my-code" {
		t.Errorf("RegistrationInviteCode = %q, want %q", cfg.RegistrationInviteCode, "my-code")
	}
}
