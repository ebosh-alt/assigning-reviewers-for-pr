// Package main wires the HTTP server for the reviewer assignment service.
package main

import (
	"context"
	"os/signal"
	"syscall"

	"assigning-reviewers-for-pr/internal/transport/http/server/handlers-fiber"
	"assigning-reviewers-for-pr/internal/usecase"

	"assigning-reviewers-for-pr/config"
	"assigning-reviewers-for-pr/internal/oapi"
	"assigning-reviewers-for-pr/internal/repository"
	"assigning-reviewers-for-pr/internal/transport/http/middleware"
	"assigning-reviewers-for-pr/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		panic(err)
	}

	repo, err := repository.New(ctx, "postgres", log, cfg)
	if err != nil {
		log.Errorw("repository initialization error", "error", err)
		return
	}
	if err := repo.OnStart(ctx); err != nil {
		log.Errorw("repository start error", "error", err)
		return
	}
	defer func() {
		_ = repo.OnStop(context.Background())
	}()

	timeout := cfg.HTTP.RequestTimeout
	uc := usecase.New(log, ctx, repo, timeout)

	serv := fiber.New(fiber.Config{
		ReadTimeout:  cfg.HTTP.RequestTimeout,
		WriteTimeout: cfg.HTTP.RequestTimeout,
	})
	serv.Use(recover.New())
	serv.Use(requestid.New())
	serv.Use(middleware.RequestLogger(log))

	serv.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	h := handlers_fiber.NewHandler(log, uc)
	api.RegisterHandlers(serv, h)

	go func() {
		if err := serv.Listen(cfg.ServerAddr()); err != nil {
			log.Errorw("failed to start server", "error", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	done := make(chan struct{})
	go func() {
		_ = serv.Shutdown()
		close(done)
	}()

	select {
	case <-done:
	case <-shutdownCtx.Done():
		log.Warnw("server shutdown timeout", "timeout", cfg.Server.ShutdownTimeout)
	}
}
