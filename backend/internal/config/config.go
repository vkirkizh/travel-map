package config

import (
	"os"
	"strings"
)

type Config struct {
	AppEnv      string
	HTTPAddr    string
	DatabaseURL string

	SecureCookies bool

	RegistrationInviteCode string
}

func Load() Config {
	appEnv := getEnv("APP_ENV", "local")

	return Config{
		AppEnv:      appEnv,
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),

		SecureCookies: appEnv == "production",

		RegistrationInviteCode: strings.TrimSpace(getEnv("REGISTRATION_INVITE_CODE", "")),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
