package config

import "os"

type AppConfig struct {
	Port       string
	CORSOrigin string
	PprofAddr  string
	PprofOn    bool
}

func Load() AppConfig {
	return AppConfig{
		Port:       getEnv("PORT", "5000"),
		CORSOrigin: getEnv("CORS_ORIGIN", "http://localhost:3000"),
		PprofAddr:  getEnv("PPROF_ADDR", "127.0.0.1:6060"),
		PprofOn:    getEnvBool("PPROF_ENABLED", true),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	switch value {
	case "1", "true", "TRUE", "True", "yes", "YES", "on", "ON":
		return true
	case "0", "false", "FALSE", "False", "no", "NO", "off", "OFF":
		return false
	default:
		return fallback
	}
}
