package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv      string
	AppPort     string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	SiteURL     string
	FrontendURL string

	MinioEndpoint       string
	MinioPublicEndpoint string
	MinioAccessKey      string
	MinioSecretKey      string
	MinioBucket         string
	MinioUseSSL         bool

	// Email (Resend)
	ResendAPIKey   string
	EmailFromEmail string
	EmailFromName  string
}

func Load() *Config {
	minioEndpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	return &Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		AppPort:             getEnv("APP_PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://politics:localdev@localhost:5432/politics_db"),
		RedisURL:            getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:           getEnv("JWT_SECRET", "your-secret-key"),
		SiteURL:             getEnv("SITE_URL", "https://pulpulitiko.com"),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:3000"),
		MinioEndpoint:       minioEndpoint,
		MinioPublicEndpoint: getEnv("MINIO_PUBLIC_ENDPOINT", minioEndpoint),
		MinioAccessKey:      getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:      getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:         getEnv("MINIO_BUCKET", "politics-media"),
		MinioUseSSL:         getEnvBool("MINIO_USE_SSL", false),
		ResendAPIKey:        getEnv("RESEND_API_KEY", ""),
		EmailFromEmail:      getEnv("EMAIL_FROM_EMAIL", "noreply@pulpulitiko.com"),
		EmailFromName:       getEnv("EMAIL_FROM_NAME", "Pulpulitiko"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return b
	}
	return defaultValue
}
