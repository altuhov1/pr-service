package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"test-task/internal/models"
	"time"

	"test-task/internal/config"
	"test-task/internal/handlers"
	"test-task/internal/storage"
)

type App struct {
	cfg      *config.Config
	server   *http.Server
	services *Services
	storages *Storages
}

type Services struct {
}

type Storages struct {
}

func NewApp(cfg *config.Config) *App {

	app := &App{
		cfg: cfg,
	}

	app.initStorages()
	app.initServices()
	app.initHTTP()

	return app
}

func (a *App) initStorages() {
	dbPGConfig := &models.PGXConfig{
		Host:     a.cfg.PG_DBHost,
		User:     a.cfg.PG_DBUser,
		Password: a.cfg.PG_DBPassword,
		DBName:   a.cfg.PG_DBName,
		SSLMode:  a.cfg.PG_DBSSLMode,
		Port:     a.cfg.PG_PORT,
	}
	poolPG, err := storage.NewPoolPg(dbPGConfig)
	_ = poolPG
	if err != nil {
		slog.Error("Failed to initialize PG (pool)", "error", err)
		os.Exit(1)
	}

	a.storages = &Storages{}
}

func (a *App) initServices() {
	a.services = &Services{}
}

func (a *App) initHTTP() {
	handler, err := handlers.NewHandler()
	if err != nil {
		slog.Error("Failed to create handler", "error", err)
		os.Exit(1)
	}

	router := a.setupRoutes(handler)

	a.server = &http.Server{
		Addr:         ":" + a.cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
}

func (a *App) setupRoutes(handler *handlers.Handler) http.Handler {
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// API routes
	apiRoutes := map[string]http.HandlerFunc{
		// "/api/allFavoriteCoin":    handler.GetFavorites,

	}
	_ = handler
	for path, handlerFunc := range apiRoutes {
		mux.HandleFunc(path, handlerFunc)
	}

	return mux
}

func (a *App) Run() {
	go a.startServer()
	a.waitForShutdown()
}

func (a *App) startServer() {
	slog.Info("Server starting", "port", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func (a *App) waitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down server gracefully...")
	a.shutdown()
}

func (a *App) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}
	slog.Info("Server stopped")
	time.Sleep(3 * time.Second)
}
