package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type OrderServiceConfig struct {
	URL string `yaml:"url"`
}

type Config struct {
	HTTP         HTTPConfig         `yaml:"http"`
	Postgres     PostgresConfig     `yaml:"postgres"`
	Log          LogConfig          `yaml:"log"`
	JWT          JWT                `yaml:"jwt"`
	OrderService OrderServiceConfig `yaml:"order_service"`
	QR           QRConfig           `yaml:"qr"`
}

type HTTPConfig struct {
	Host string `yaml:"host" env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port int    `yaml:"port" env:"HTTP_PORT" env-default:"8082"`
}

func (h HTTPConfig) Address() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

type PostgresConfig struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST" env-required:"true"`
	Port     int    `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"POSTGRES_USER" env-default:"postgres"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
	Database string `yaml:"database" env:"POSTGRES_DB"`
	SSLMode  string `yaml:"sslmode" env:"POSTGRES_SSLMODE" env-default:"disable"`

	MaxOpenConns    int           `yaml:"max_open_conns" env-default:"20"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env-default:"10"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env-default:"30m"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" env-default:"15m"`
}

type LogConfig struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
}

type JWT struct {
	Issuer        string `yaml:"issuer" env:"JWT_ISSUER" env-required:"true"`
	PublicKeyPath string `yaml:"public_key_path" env:"JWT_PUBLIC_KEY_PATH" env-required:"true"`
}

type QRConfig struct {
	SecretKey string        `yaml:"secret_key" env:"QR_SECRET_KEY"`
	TokenTTL  time.Duration `yaml:"token_ttl" env:"QR_TOKEN_TTL" env-default:"24h"`
}

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", configPath, err)
	}

	data = []byte(os.ExpandEnv(string(data)))

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", configPath, err)
	}

	return &cfg, nil
}
