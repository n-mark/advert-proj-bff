package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"bff-finalproj/internal/config"
	"bff-finalproj/internal/handlers"
	"bff-finalproj/internal/server"
	"bff-finalproj/internal/service"
)

func main() {
	cfg := config.MustLoad()
	bff := service.NewBFF(cfg)
	h := handlers.NewHandler(bff)
	router := server.New(h)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx, server.AddrFromEnv(), router); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
