package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/liriquew/secret_storage/server/internal/app"
	"github.com/liriquew/secret_storage/server/internal/lib/config"
	"github.com/liriquew/secret_storage/server/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupPrettySlog("SECRET_STORAGE")
	log.Info("Config is loaded: ", slog.Any("Config", cfg))

	application := app.New(log, cfg)

	application.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Info("Server shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.Stop(ctx); err != nil {
		log.Error("Error while server shutdown: %w", err)
		return
	}

	log.Info("Server shutdown successful")
}
