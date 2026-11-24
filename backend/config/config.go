package config

import (
	"os"
)

type Config struct {
	Port         string
	Environment  string
	DatabasePath string
	JWTSecret    string
	OTelEndpoint string
	CORSOrigin   string
}

func LoadConfig() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		Environment:  getEnv("APP_ENV", "development"),
		DatabasePath: getEnv("DATABASE_PATH", "./data/todo.db"),
		JWTSecret:    getEnv("JWT_SECRET", "super-secret-production-grade-key-change-in-prod-32chars"),
		OTelEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317"),
		CORSOrigin:   getEnv("CORS_ORIGIN", "http://localhost:3000,http://127.0.0.1:3000"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
