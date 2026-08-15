package config

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	BaseFolder string `env:"BASE_FOLDER"`
	BaseUrl    string `env:"BASE_URL"`
}

var cfg *Config

func GetConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse env: %v", err)
	}

	log.Printf("Config loaded: %+v\n", cfg)
	return cfg
}
