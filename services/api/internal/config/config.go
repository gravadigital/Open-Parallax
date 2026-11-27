package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	DB struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
		SSLMode  string
	}

	Server struct {
		Port        string
		GinMode     string
		FrontendURL string
	}

	Upload struct {
		Dir         string
		MaxFileSize int64
	}

	Storage struct {
		Provider  string // "local" or "s3"
		LocalPath string
		// S3-compatible object storage (Garage). Endpoints carry the scheme.
		S3Endpoint        string
		S3PresignEndpoint string
		S3Region          string
		S3Bucket          string
		S3AccessKeyID     string
		S3SecretAccessKey string
	}

	CORS struct {
		AllowOrigins string
		AllowMethods string
		AllowHeaders string
	}

	Email struct {
		Enabled      bool
		SMTPHost     string
		SMTPPort     string
		SMTPUser     string
		SMTPPassword string
		FromAddress  string
		FromName     string
		Secure       bool // true = SSL/TLS directo (puerto 465), false = STARTTLS (puerto 587)
	}

	Google struct {
		ClientID string
	}

	Auth struct {
		JWTSecret string // no default: Validate rejects an empty one
	}
}

// Load loads configuration from environment variables
func Load() *Config {
	_ = godotenv.Load()

	config := &Config{}

	config.DB.Host = getEnv("DB_HOST", "localhost")
	config.DB.Port = getEnv("DB_PORT", "5432")
	config.DB.User = getEnv("DB_USER", "openparallax")
	config.DB.Password = getEnv("DB_PASSWORD", "openparallax_password")
	config.DB.Name = getEnv("DB_NAME", "openparallax_db")
	config.DB.SSLMode = getEnv("DB_SSLMODE", "disable")

	config.Server.Port = getEnv("PORT", "8080")
	config.Server.GinMode = getEnv("GIN_MODE", "debug")
	config.Server.FrontendURL = getEnv("FRONTEND_URL", "http://localhost:3000")

	config.Upload.Dir = getEnv("UPLOADS_DIR", "./uploads")
	config.Upload.MaxFileSize = getEnvAsInt64("MAX_FILE_SIZE", 10485760)

	// Storage configuration
	config.Storage.Provider = getEnv("STORAGE_PROVIDER", "local") // "local" or "s3"
	config.Storage.LocalPath = getEnv("STORAGE_LOCAL_PATH", "./uploads")
	config.Storage.S3Endpoint = getEnv("S3_ENDPOINT", "http://localhost:3900")
	config.Storage.S3PresignEndpoint = getEnv("S3_PRESIGN_ENDPOINT", "")
	// Must match the server's s3_region; Garage's default is "garage".
	config.Storage.S3Region = getEnv("S3_REGION", "garage")
	config.Storage.S3Bucket = getEnv("S3_BUCKET", "openparallax")
	config.Storage.S3AccessKeyID = getEnv("S3_ACCESS_KEY_ID", "")
	config.Storage.S3SecretAccessKey = getEnv("S3_SECRET_ACCESS_KEY", "")

	config.CORS.AllowOrigins = getEnv("CORS_ALLOW_ORIGINS", "*")
	config.CORS.AllowMethods = getEnv("CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS")
	config.CORS.AllowHeaders = getEnv("CORS_ALLOW_HEADERS", "Origin,Content-Length,Content-Type,Authorization")

	config.Email.Enabled = getEnvAsBool("EMAIL_ENABLED", false)
	config.Email.SMTPHost = getEnv("SMTP_HOST", "")
	config.Email.SMTPPort = getEnv("SMTP_PORT", "587")
	config.Email.SMTPUser = getEnv("SMTP_USER", "")
	config.Email.SMTPPassword = getEnv("SMTP_PASSWORD", "")
	config.Email.FromAddress = getEnv("EMAIL_FROM", "")
	config.Email.FromName = getEnv("EMAIL_FROM_NAME", "OpenParallax")
	config.Email.Secure = getEnvAsBool("SMTP_SECURE", false)
	config.Google.ClientID = getEnv("GOOGLE_CLIENT_ID", "")
	config.Auth.JWTSecret = getEnv("JWT_SECRET", "")

	return config
}

// Validate checks the settings the API cannot run without. It is separate from Load because
// other binaries (cmd/migrate) load the same config but do not need them.
func (c *Config) Validate() error {
	if c.Auth.JWTSecret == "" {
		return errors.New("JWT_SECRET is required: set it to a long random value (the Makefile and deploy/docker-compose.yml set one for local development)")
	}
	return nil
}

// GetDatabaseURL returns the database connection URL
func (c *Config) GetDatabaseURL() string {
	return "postgres://" + c.DB.User + ":" + c.DB.Password + "@" + c.DB.Host + ":" + c.DB.Port + "/" + c.DB.Name + "?sslmode=" + c.DB.SSLMode
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt64 gets an environment variable as int64 or returns a default value
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as bool or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
