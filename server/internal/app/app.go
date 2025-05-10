package app

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liriquew/secret_storage/server/internal/app/api"
	"github.com/liriquew/secret_storage/server/internal/lib/config"
	service "github.com/liriquew/secret_storage/server/internal/service"
)

type App struct {
	router *gin.Engine
	srv    *http.Server
}

func New(log *slog.Logger, cfg config.AppConfig) *App {
	service := service.New(log, cfg.Storage)

	r := api.New(service)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Service.Port),
		Handler: r.Handler(),
	}

	return &App{
		router: r,
		srv:    srv,
	}
}

func (a *App) Start() {
	go func() {
		if err := a.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
}

func (a *App) Stop(ctx context.Context) error {
	return a.srv.Shutdown(ctx)
}
