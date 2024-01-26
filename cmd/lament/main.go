package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/kirinyoku/lament/internal/http-server/handlers/redirect"
	del "github.com/kirinyoku/lament/internal/http-server/handlers/url/delete"
	"github.com/kirinyoku/lament/internal/http-server/handlers/url/save"
	slogprty "github.com/kirinyoku/lament/internal/lib/logger/handlers/slog-pretty"
	"log/slog"
	"net/http"
	"os"

	"github.com/kirinyoku/lament/internal/config"
	"github.com/kirinyoku/lament/internal/lib/logger/sl"
	"github.com/kirinyoku/lament/internal/storage/sqlite"
)

const (
	envLOCAL = "local"
	envDEV   = "dev"
	envPROD  = "prod"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		os.Exit(1)
	}

	cfgPath := os.Getenv("CONFIG_PATH")
	cfg := config.MustLoad(cfgPath)

	log := setupLogger(cfg.Env)

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	var _ = storage

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.BasicAuth("lament", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/url", save.New(log, storage))
		r.Delete("/url", del.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLOCAL:
		log = setupPrettySlog()
	case envDEV:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envPROD:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogprty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
