package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the template application.
// Designed for simplicity: ~15 required vars for a working backend.
// Extended config (AI, OAuth, etc.) can be added as needed per project.
type Config struct {
	// Application
	Port string
	Env  string

	// Multi-tenancy
	MultiTenantEnabled bool
	DefaultTenantID    string

	// Database
	DBHost     string
	DBPort     string
	DBUsername string
	DBPassword string
	DBName     string

	// Connection pool tuning
	DBPoolMaxOpen int
	DBPoolMaxIdle int

	// CORS
	CORSAllowedOrigins string // Comma-separated, e.g., "http://localhost:3000,http://localhost:5173"

	// JWT
	JWTSecret string

	// Email (optional, provide interface implementation or leave empty)
	SMTPHost      string
	SMTPPort      string
	SMTPUsername  string
	SMTPPassword  string
	SMTPFromEmail string

	// File upload (optional, provide interface implementation or leave empty)
	// Can be MinIO, S3, local filesystem, etc.
	FileStorageType string // "minio", "s3", "local", etc.
	MinioURL        string
	MinioBucket     string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioPublicURL  string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	// Load .env file if it exists (for local dev)
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: .env file not found or error reading it: %v", err)
	}

	cfg := &Config{
		Port:               getEnv("APP_PORT", "8080"),
		Env:                getEnv("APP_ENV", "development"),
		MultiTenantEnabled: getEnvBool("MULTI_TENANT_ENABLED", true),
		DefaultTenantID:    getEnv("DEFAULT_TENANT_ID", ""),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUsername:         getEnv("DB_USERNAME", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", ""),
		DBName:             getEnv("DB_NAME", "template_db"),
		DBPoolMaxOpen:      getEnvInt("DB_POOL_MAX_OPEN", 25),
		DBPoolMaxIdle:      getEnvInt("DB_POOL_MAX_IDLE", 5),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173"),
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		SMTPHost:           getEnv("SMTP_HOST", ""),
		SMTPPort:           getEnv("SMTP_PORT", ""),
		SMTPUsername:       getEnv("SMTP_USERNAME", ""),
		SMTPPassword:       getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:      getEnv("SMTP_FROM_EMAIL", ""),
		FileStorageType:    getEnv("FILE_STORAGE_TYPE", "local"),
		MinioURL:           getEnv("MINIO_URL", ""),
		MinioBucket:        getEnv("MINIO_BUCKET", ""),
		MinioAccessKey:     getEnv("MINIO_ACCESS_KEY", ""),
		MinioSecretKey:     getEnv("MINIO_SECRET_KEY", ""),
		MinioPublicURL:     getEnv("MINIO_PUBLIC_URL", ""),
	}

	return cfg, nil
}

// Validate checks that required configuration is present.
func (c *Config) Validate() error {
	required := []string{
		"DBHost", "DBPort", "DBName", "JWTSecret",
	}

	var missing []string
	for _, field := range required {
		switch field {
		case "DBHost":
			if c.DBHost == "" {
				missing = append(missing, "DB_HOST")
			}
		case "DBPort":
			if c.DBPort == "" {
				missing = append(missing, "DB_PORT")
			}
		case "DBName":
			if c.DBName == "" {
				missing = append(missing, "DB_NAME")
			}
		case "JWTSecret":
			if c.JWTSecret == "" || c.JWTSecret == "your-secret-key-change-in-production" {
				missing = append(missing, "JWT_SECRET (must be set to a secure value)")
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// getEnv gets an environment variable with a default.
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

// getEnvBool gets a boolean environment variable with a default.
func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}

// getEnvInt gets an integer environment variable with a default.
func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}
