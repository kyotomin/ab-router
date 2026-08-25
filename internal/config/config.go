package config

import (
	"fmt"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	App struct {
		RouterPort int `env:"ROUTER_PORT" env-default:"8080"`
		AdminPort  int `env:"ADMIN_PORT" env-default:"8081"`
	} `yaml:"app"`
	DB struct {
		Host        string `env:"DB_HOST" env-default:"localhost"`
		Port        string `env:"DB_PORT" env-default:"5432"`
		User        string `env:"DB_USER" env-default:"postgres"`
		Password    string `env:"DB_PASSWORD" env-default:"postgres"`
		Database    string `env:"DB_NAME" env-default:"feature_router"`
		SSLMode     string `env:"DB_SSLMODE" env-default:"disable"`
		PingRetries int    `env:"PING_RETRIES" env-default:"5"`
	} `yaml:"db"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatalf(".env не найден: %v", err)
	}
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Fatalf("Не удалось загрузить конфиг: %v", err)
	}

	return &cfg
}

func (cfg *Config) GetDBConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Database,
		cfg.DB.SSLMode,
	)
}

func (cfg *Config) GetMigrationsConnString() string {
	return fmt.Sprintf(
		"pgx5://%s:%s@%s:%s/%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Database,
	)
}
