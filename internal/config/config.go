package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string   `yaml:"env" env-required:"true"`
	Proxy    Proxy    `yaml:"proxy" env-required:"true"`
	Backends []string `yaml:"backends" env-required:"true"`
}

type Proxy struct {
	Host         string        `yaml:"host" env-required:"true"`
	Port         string        `yaml:"port" env-required:"true"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"5s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"5s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env-default:"5s"`
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
