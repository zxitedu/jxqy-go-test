package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	MySQL mysqlConfig `yaml:"mysql"`
}

type mysqlConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DB       string `yaml:"db"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func loadConfig(path string) (config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return config{}, err
	}

	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return config{}, err
	}
	return cfg, nil
}

func (cfg config) validate(serviceName string) error {
	mysql := cfg.MySQL

	if mysql.Host == "" {
		return fmt.Errorf("mysql.host is required")
	}
	if mysql.Port <= 0 || mysql.Port > 65535 {
		return fmt.Errorf("mysql.port must be between 1 and 65535")
	}
	if mysql.DB == "" {
		return fmt.Errorf("mysql.db is required")
	}
	if mysql.DB != serviceName {
		return fmt.Errorf("mysql.db must match executable name: got %q, want %q", mysql.DB, serviceName)
	}
	if mysql.User == "" {
		return fmt.Errorf("mysql.user is required")
	}
	if mysql.Password == "" {
		return fmt.Errorf("mysql.password is required")
	}

	return nil
}
