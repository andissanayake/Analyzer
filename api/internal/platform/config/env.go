package config

import "os"

type AppConfig struct {
	Port       string
	CORSOrigin string
}

func Load() AppConfig {
	return AppConfig{
		Port:       getEnv("PORT", "5000"),
		CORSOrigin: getEnv("CORS_ORIGIN", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
