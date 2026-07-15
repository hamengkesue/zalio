package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config memuat semua pengaturan yang dibaca dari environment (.env).
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	ServerPort string
	CORSOrigin string
	JWTSecret  string
}

// Load membaca .env (kalau ada) lalu environment, dengan nilai default aman.
func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5435"),
		DBUser:     getEnv("DB_USER", "zalio_erp"),
		DBPassword: getEnv("DB_PASSWORD", "zalio_erp_secret"),
		DBName:     getEnv("DB_NAME", "zalio_erp"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		ServerPort: getEnv("SERVER_PORT", "8082"),
		CORSOrigin: getEnv("CORS_ORIGIN", "*"),
		JWTSecret:  getEnv("JWT_SECRET", "dev-secret-ganti-di-produksi"),
	}
}

// DSN merangkai connection string PostgreSQL.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
