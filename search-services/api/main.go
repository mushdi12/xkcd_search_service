package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"yadro.com/course/api/adapters/aaa"
	"yadro.com/course/api/adapters/rest"
	"yadro.com/course/api/adapters/rest/middleware"
	"yadro.com/course/api/adapters/search"
	"yadro.com/course/api/adapters/update"
	"yadro.com/course/api/adapters/words"
	"yadro.com/course/api/config"
	"yadro.com/course/api/core"
	"yadro.com/course/closers"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting server")
	log.Debug("debug messages are enabled")

	wordsClient, err := words.NewClient(cfg.WordsAddress, log)
	if err != nil {
		log.Error("cannot init words adapter", "error", err)
		os.Exit(1)
	}

	updateClient, err := update.NewClient(cfg.UpdateAddress, log)
	if err != nil {
		log.Error("cannot init update adapter", "error", err)
		os.Exit(1)
	}

	searchClient, err := search.NewClient(cfg.SearchAddress, log)
	if err != nil {
		log.Error("cannot init search adapter", "error", err)
		os.Exit(1)
	}

	aaaService, err := aaa.New(cfg.TokenTTL, log)
	if err != nil {
		log.Error("cannot init aaa adapter", "error", err)
		os.Exit(1)
	}

	server := mustSetupHttpServer(log, cfg, updateClient, searchClient, wordsClient, aaaService)

	defer closers.CloseOrLog(log, wordsClient, updateClient, searchClient)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	log.Info("Running HTTP server", "address", cfg.HTTPConfig.Address)
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server closed unexpectedly", "error", err)
			return
		}
	}
}

func mustSetupHttpServer(log *slog.Logger, cfg config.Config, updateClient *update.Client, searchClient *search.Client, wordsClient *words.Client, aaaService aaa.AAA) *http.Server {
	mux := http.NewServeMux()

	pingers := map[string]core.Pinger{
		"update": updateClient,
		"search": searchClient,
		"words":  wordsClient,
	}
	// pinger
	mux.Handle("GET /api/ping",
		rest.NewPingHandler(log, pingers))
	// login handler
	mux.Handle("POST /api/login",
		rest.NewLoginHandler(log, aaaService))
	// search client
	mux.Handle("GET /api/search",
		middleware.Concurrency(rest.NewSearchHandler(log, searchClient), cfg.SearchConcurrency))
	mux.Handle("GET /api/isearch",
		middleware.Rate(rest.NewIndexSearchHandler(log, searchClient), cfg.SearchRate),
	)
	// update client
	mux.Handle("POST /api/db/update",
		middleware.Auth(rest.NewUpdateHandler(log, updateClient), aaaService),
	)
	mux.Handle("GET /api/db/stats",
		rest.NewUpdateStatsHandler(log, updateClient))
	mux.Handle("GET /api/db/status",
		rest.NewUpdateStatusHandler(log, updateClient))
	mux.Handle("DELETE /api/db",
		middleware.Auth(rest.NewDropHandler(log, updateClient), aaaService),
	)

	return &http.Server{
		Addr:        cfg.HTTPConfig.Address,
		ReadTimeout: cfg.HTTPConfig.Timeout,
		Handler:     mux,
	}
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level, AddSource: true})
	return slog.New(handler)
}
