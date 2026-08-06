package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() *Config {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := &Config{
		DatabaseURL: getEnv("Database_URL", ""),
		Port:        getEnv("PORT", ""),
	}

	if cfg.DatabaseURL == "" || cfg.Port == "" {
		log.Fatal("No database URL found")
	}

	return cfg
}

func getEnv(key string, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
