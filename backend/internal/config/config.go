package config

import (
	"os"
	"strings"
)

type Config struct {
	AppEnv      string
	HTTPAddr    string
	DatabaseURL string

	RegistrationInviteCode string
}

func Load() Config {
	return Config{
		AppEnv:      getEnv("APP_ENV", "local"),
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),

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
