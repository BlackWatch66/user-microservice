package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config stores all application configurations
type Config struct {
	HTTPPort     string
	GRPCPort     string
	DatabaseURL  string
	RedisAddr    string
	RedisPassword string
	RedisDB      int
	JWTSecret    string
	JWTExpiry    time.Duration
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Set default values
	cfg := &Config{
		HTTPPort:     "8080",
		GRPCPort:     "50051",
		RedisPassword: "",
		RedisDB:      0,
		JWTExpiry:    15 * time.Minute,
	}

	// Load from environment variables
	cfg.HTTPPort = getEnv("HTTP_PORT", cfg.HTTPPort)
	cfg.GRPCPort = getEnv("GRPC_PORT", cfg.GRPCPort)
	cfg.DatabaseURL = getEnvOrPanic("DATABASE_URL")
	cfg.RedisAddr = getEnvOrPanic("REDIS_ADDR")
	cfg.RedisPassword = getEnv("REDIS_PASSWORD", cfg.RedisPassword)
	redisDBStr := getEnv("REDIS_DB", strconv.Itoa(cfg.RedisDB))
	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		log.Printf("Invalid REDIS_DB value: %s. Using default: %d\n", redisDBStr, cfg.RedisDB)
	} else {
		cfg.RedisDB = redisDB
	}

	cfg.JWTSecret = getEnvOrPanic("JWT_SECRET")
	jwtExpiryStr := getEnv("JWT_EXPIRY_MINUTES", "15")
	jwtExpiryMinutes, err := strconv.Atoi(jwtExpiryStr)
	if err != nil {
		log.Printf("Invalid JWT_EXPIRY_MINUTES value: %s. Using default: 15 minutes\n", jwtExpiryStr)
	} else {
		cfg.JWTExpiry = time.Duration(jwtExpiryMinutes) * time.Minute
	}

	return cfg
}

// getEnv retrieves environment variable, returns fallback value if not found
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvOrPanic retrieves environment variable, panics if not found
func getEnvOrPanic(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Fatalf("Environment variable %s is not set", key)
	return "" // This line will not be executed
} 