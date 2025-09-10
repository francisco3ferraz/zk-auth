package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Security SecurityConfig
	SRP      SRPConfig
}

type DatabaseConfig struct {
	URL            string
	MaxConnections int
	MaxIdleConns   int
	MaxLifetime    time.Duration
	MaxIdleTime    time.Duration
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Environment  string
}

type SecurityConfig struct {
	JWTSecret       string
	JWTExpiry       time.Duration
	BCryptCost      int
	RateLimitReqs   int
	RateLimitWindow time.Duration
}

type SRPConfig struct {
	KeyLength     int
	HashAlgorithm string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}

	cfg.Database.URL = getEnv("DATABASE_URL", "postgres://localhost/zkauth?sslmode=disable")
	cfg.Database.MaxConnections = getEnvAsInt("DB_MAX_CONNECTIONS", 25)
	cfg.Database.MaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 5)
	cfg.Database.MaxLifetime = getEnvAsDuration("DB_MAX_LIFETIME", 5*time.Minute)
	cfg.Database.MaxIdleTime = getEnvAsDuration("DB_MAX_IDLE_TIME", 5*time.Minute)

	cfg.Server.Port = getEnv("SERVER_PORT", "8080")
	cfg.Server.ReadTimeout = getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second)
	cfg.Server.WriteTimeout = getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second)
	cfg.Server.IdleTimeout = getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second)
	cfg.Server.Environment = getEnv("ENVIRONMENT", "development")

	cfg.Security.JWTSecret = getEnv("JWT_SECRET", "")
	if cfg.Security.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	cfg.Security.JWTExpiry = getEnvAsDuration("JWT_EXPIRY", 24*time.Hour)
	cfg.Security.BCryptCost = getEnvAsInt("BCRYPT_COST", 12)
	cfg.Security.RateLimitReqs = getEnvAsInt("RATE_LIMIT_REQUESTS", 100)
	cfg.Security.RateLimitWindow = getEnvAsDuration("RATE_LIMIT_WINDOW", time.Minute)

	cfg.SRP.KeyLength = getEnvAsInt("SRP_KEY_LENGTH", 2048)
	cfg.SRP.HashAlgorithm = getEnv("SRP_HASH_ALGORITHM", "SHA256")

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsDuration(name string, defaultVal time.Duration) time.Duration {
	valueStr := getEnv(name, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultVal
}
