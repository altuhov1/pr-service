package config

import (
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort                string `env:"PORT_APP" envDefault:"8080"`
	NameFileAllTasks          string `env:"ALL_TASKS_FILE" envDefault:"storage/AllTasks.json"`
	NameFileProcessTasksLinks string `env:"PROCESS_LINKS_FILE" envDefault:"storage/ProcessTasksLinks.json"`
	NameFileProcessTasksNums  string `env:"PROCESS_NUMS_FILE" envDefault:"storage/ProcessTasksNums.json"`
	PG_DBHost                 string `env:"DB_PG_HOST" envDefault:"postgres"`
	PG_DBUser                 string `env:"USER_DB_PG" envDefault:""`
	PG_DBPassword             string `env:"PASS_DB_PG" envDefault:""`
	PG_DBName                 string `env:"NAME_DB_PG" envDefault:"webdev"`
	PG_DBSSLMode              string `env:"DB_PG_SSLMODE" envDefault:"disable"`
	PG_PORT                   string `env:"DB_PG_PORT" envDefault:"5432"`
}

func MustLoad() *Config {

	if err := godotenv.Load(); err != nil {
		slog.Debug("Failed to load .env file", "error", err)
	} else {
		slog.Info("Loaded configuration from .env file")
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Failed to parse environment variables", "error", err)
		panic("configuration error: " + err.Error())
	}

	return &cfg
}
