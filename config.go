package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	Settings settingsConfig `yaml:"settings"`
}

type settingsConfig struct {
	Application applicationConfig `yaml:"application"`
	Database    databaseConfig    `yaml:"database"`
	Gen         genConfig         `yaml:"gen"`
}

type applicationConfig struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

type databaseConfig struct {
	Driver string `yaml:"driver"`
	Source string `yaml:"source"`
}

type genConfig struct {
	DBName string `yaml:"dbname"`
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
	app := cfg.Settings.Application
	db := cfg.Settings.Database
	gen := cfg.Settings.Gen
	expectedDB := serviceDatabaseName(serviceName)

	if app.Name == "" {
		return fmt.Errorf("settings.application.name is required")
	}
	if app.Name != serviceName {
		return fmt.Errorf("settings.application.name must match executable name: got %q, want %q", app.Name, serviceName)
	}
	if app.Port <= 0 || app.Port > 65535 {
		return fmt.Errorf("settings.application.port must be between 1 and 65535")
	}
	if db.Driver == "" {
		return fmt.Errorf("settings.database.driver is required")
	}
	if db.Driver != "mysql" {
		return fmt.Errorf("settings.database.driver must be mysql: got %q", db.Driver)
	}
	if db.Source == "" {
		return fmt.Errorf("settings.database.source is required")
	}

	sourceDB, err := mysqlDatabaseName(db.Source)
	if err != nil {
		return fmt.Errorf("settings.database.source is invalid: %w", err)
	}
	if sourceDB != expectedDB {
		return fmt.Errorf("settings.database.source db must match service database name: got %q, want %q", sourceDB, expectedDB)
	}
	if gen.DBName != "" && gen.DBName != expectedDB {
		return fmt.Errorf("settings.gen.dbname must match service database name: got %q, want %q", gen.DBName, expectedDB)
	}

	return nil
}

func serviceDatabaseName(serviceName string) string {
	return serviceName + "db"
}
