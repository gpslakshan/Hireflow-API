package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the application.
// Loaded once at startup and passed around — no global state.
type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret      string
	JWTExpiryHours string

	AdminEmail    string
	AdminPassword string

	AWSAccessKeyID          string
	AWSSecretAccessKey      string
	AWSRegion               string
	AWSS3Bucket             string
	CVUploadURLExpiryMins   string
	CVDownloadURLExpiryMins string
}

// Load reads the .env file and returns a populated Config struct.
func Load() *Config {
	// In production you won't have a .env file — env vars are injected directly.
	// godotenv.Load() doesn't fail if the file is missing, which is the correct behaviour.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "hireflow_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpiryHours: getEnv("JWT_EXPIRY_HOURS", "72"),

		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@hireflow.com"),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),

		AWSAccessKeyID:          getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey:      getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSRegion:               getEnv("AWS_REGION", "ap-south-1"),
		AWSS3Bucket:             getEnv("AWS_S3_BUCKET", ""),
		CVUploadURLExpiryMins:   getEnv("CV_UPLOAD_URL_EXPIRY_MINUTES", "15"),
		CVDownloadURLExpiryMins: getEnv("CV_DOWNLOAD_URL_EXPIRY_MINUTES", "15"),
	}
}

// getEnv reads an environment variable, returning a fallback if not set.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
