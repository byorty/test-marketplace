package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ProductService struct {
	URL string `yaml:"url"`
}
type Config struct {
    HTTP HTTPConfig `yaml:"http"`
    Postgres PostgresConfig `yaml:"postgres"`
    Log LogConfig `yaml:"log"`
	JWT JWT `yaml:"jwt"`
	ProductService ProductService `yaml:"product_service"`
}

type HTTPConfig struct {
    Host string `yaml:"host" env:"POSTGRES_HOST" env-required:"true"`
    Port int `yaml:"port" env:"POSTGRES_PORT" env-default:"8080"`
}

func (h HTTPConfig) Address() string {
    return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

type PostgresConfig struct {
    Host string `yaml:"host" env:"POSTGRES_HOST" env-required:"true"`
    Port int `yaml:"port" env:"HTTP_PORT" env-default:"5432"`
    User string `yaml:"user" env:"POSTGRES_USER" env-default:"postgres"`
    Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
    Database string `yaml:"database" env:"POSTGRES_DB"`
    SSLMode string `yaml:"sslmode" env:"POSTGRES_SSLMODE" env-default:"disable"`

    MaxOpenConns int `yaml:"max_open_conns" env-default:"20"`
    MaxIdleConns int `yaml:"max_idle_conns" env-default:"10"`
   
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env-default:"30m"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" env-default:"15m"`
}

type LogConfig struct {
    Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
} 

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("read config %q: %w", configPath, err)
	}

	return &cfg, nil
}

type JWT struct {
    Issuer        string `yaml:"issuer"`
    PublicKeyPath string `yaml:"public_key_path"`
}