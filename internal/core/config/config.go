package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	PostgresDSN string
	MongoURI    string
	RedisAddr   string
	RedisPass   string
	JWTSecret   string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[config] .env not found, using OS env")
	}

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		PostgresDSN: getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/grab?sslmode=disable"),
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:   getEnv("REDIS_PASS", ""),
		JWTSecret:   getEnv("JWT_SECRET", "super-secret-change-me"),
	}

	if cfg.PostgresDSN == "" {
		log.Fatal("[config] POSTGRES_DSN is required")
	}
	if cfg.JWTSecret == "" || cfg.JWTSecret == "super-secret-change-me" {
		log.Println("[config] WARNING: JWT_SECRET is empty or using the default value; set a strong secret in production")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
