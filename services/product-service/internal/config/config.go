package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
    HTTP HTTPConfig `yaml:"http"`
    Postgres PostgresConfig `yaml:"postgres"`
    Log LogConfig `yaml:"log"`
	JWT JWT `yaml:"jwt"`
}

type HTTPConfig struct {
    Host string `yaml:"host" env:"HTTP_HOST" env-required:"true"`
    Port int `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
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

func MustLoad() *Config {
    configPath := os.Getenv("CONFIG_PATH")
    if configPath == "" {
        configPath = "config/config.yaml"
    }

    var cfg Config

    if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
        log.Fatalf("failed to read config %q: %v", configPath, err)
    }

    return &cfg
}

type JWT struct {
	Secret string `yaml:"secret"`
	Issuer string `yaml:"issuer"`
}