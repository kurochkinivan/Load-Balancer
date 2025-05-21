package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string           `yaml:"env" env-required:"true"`
	Proxy      ProxyConfig      `yaml:"proxy" env-required:"true"`
	PostgreSQL PostgreSQLConfig `yaml:"postgresql" env-required:"true"`
	Cache      Cache            `yaml:"cache" env-required:"true"`
	Backends   []string         `yaml:"backends" env-required:"true"`
}

type ProxyConfig struct {
	Host         string        `yaml:"host" env-required:"true"`
	Port         string        `yaml:"port" env-required:"true"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"5s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"5s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env-default:"5s"`
	HealthCheck  HealthCheck   `yaml:"health_check"`
}

type HealthCheck struct {
	Interval     time.Duration `yaml:"interval" env-default:"30s"`
	WorkersCount int           `yaml:"workers_count" env-default:"10"`
}

type PostgreSQLConfig struct {
	Host       string               `yaml:"host" env-default:"localhost"`
	Port       string               `yaml:"port" env-default:"5435"`
	Username   string               `yaml:"username" env-default:"postgres"`
	Password   string               `yaml:"password" env-default:"postgres"`
	DB         string               `yaml:"db" env-default:"clients"`
	Connection PostgreSQLConnection `yaml:"connection"`
}

type PostgreSQLConnection struct {
	Attempts int           `yaml:"attempts" env-default:"5"`
	Delay    time.Duration `yaml:"delay" env-default:"5s"`
}

type Cache struct {
	MaxElements int `yaml:"max_elements" env-default:"50"`
}

func MustLoadConfig() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}

func LoadConfig() (*Config, error) {
	path, err := fetchConfigPath()
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	err = cleanenv.ReadConfig(path, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return cfg, nil
}

func fetchConfigPath() (string, error) {
	var path string
	flag.StringVar(&path, "path", "", "path to config.yaml")
	flag.Parse()

	if path == "" {
		return "", fmt.Errorf("you have to specify config path using --path flag")
	}

	return path, nil
}
