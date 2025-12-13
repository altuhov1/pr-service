package main

import (
	"log/slog"
	"os"
	"subscription-budget/internal/app"
	"subscription-budget/internal/config"
)

func main() {
	cfg := config.MustLoad()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	app := app.NewApp(cfg)
	app.Run()
}
