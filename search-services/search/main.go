package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"yadro.com/course/closers"
	searchpb "yadro.com/course/proto/search"
	searchdb "yadro.com/course/search/adapters/db"
	searchgrpc "yadro.com/course/search/adapters/grpc"
	"yadro.com/course/search/adapters/initiator"
	"yadro.com/course/search/adapters/nats"
	words "yadro.com/course/search/adapters/word"
	"yadro.com/course/search/config"
	"yadro.com/course/search/core"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()
	cfg := config.MustLoad(configPath)

	log := mustSetupLogger(cfg.LogLevel)

	if err := run(cfg, log); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting server")
	log.Debug("debug messages are enabled")

	// database adapter
	storage, err := searchdb.New(log, cfg.DBAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %v", err)
	}

	// words adapter
	wordsClient, err := words.NewClient(cfg.WordsAddress, log)
	if err != nil {
		return fmt.Errorf("failed to create words client: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// indexer initiator
	initiator := initiator.NewInitiator(log, storage, cfg.IndexTTL)
	go initiator.Start(ctx)

	// service
	search, err := core.NewService(log, storage, wordsClient, initiator)
	if err != nil {
		return fmt.Errorf("failed to create search service: %v", err)
	}

	// nats listener
	listener, err := nats.NewListener(cfg.BrokerAddress, cfg.Topic, log, initiator)
	if err != nil {
		return fmt.Errorf("failed to create nats listener: %v", err)
	}
	go listener.Listen(ctx)

	// grpc server
	grpcListener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// closers for all adapters
	defer closers.CloseOrLog(log, wordsClient, storage, initiator, listener)

	s := grpc.NewServer()
	searchpb.RegisterSearchServer(s, searchgrpc.NewServer(search))
	reflection.Register(s)

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		s.GracefulStop()
	}()

	if err := s.Serve(grpcListener); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

func mustSetupLogger(logLevel string) *slog.Logger {
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
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}
