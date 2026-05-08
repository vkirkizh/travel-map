package config

import (
	"os"
)

type Config struct {
	AppEnv      string
	HTTPAddr    string
	DatabaseURL string

	GeocoderProvider   string
	NominatimBaseURL   string
	NominatimUserAgent string
}

func Load() Config {
	return Config{
		AppEnv:      getEnv("APP_ENV", "local"),
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://travel_map:travel_map@localhost:5432/travel_map?sslmode=disable"),

		GeocoderProvider:   getEnv("GEOCODER_PROVIDER", "nominatim"),
		NominatimBaseURL:   getEnv("NOMINATIM_BASE_URL", "https://nominatim.openstreetmap.org"),
		NominatimUserAgent: getEnv("NOMINATIM_USER_AGENT", "TravelMap/0.1 (https://github.com/vkirkizh/travel-map)"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
