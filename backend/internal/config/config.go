package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	ServerPort string
	ServerHost string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret     string
	JWTExpiration int // hours

	// Gemini AI
	GeminiAPIKey string
	GeminiModel  string

	// File Storage
	UploadDir string

	// CORS
	AllowedOrigins string

	// Scraper
	ScraperEnabled  bool
	ScraperInterval string // cron expression
}

// Load reads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		ServerHost:      getEnv("SERVER_HOST", "0.0.0.0"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/jobcrawler?sslmode=disable"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:       getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiration:   getEnvInt("JWT_EXPIRATION_HOURS", 72),
		GeminiAPIKey:    getEnv("GEMINI_API_KEY", ""),
		GeminiModel:     getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
		UploadDir:       getEnv("UPLOAD_DIR", "./uploads"),
		AllowedOrigins:  getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		ScraperEnabled:  getEnvBool("SCRAPER_ENABLED", false),
		ScraperInterval: getEnv("SCRAPER_INTERVAL", "0 */6 * * *"), // every 6 hours
	}

	if cfg.JWTSecret == "change-me-in-production" {
		fmt.Println("⚠️  WARNING: Using default JWT secret. Set JWT_SECRET in production!")
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return fallback
}
