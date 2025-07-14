package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/nutochk/13-07-25/internal/config"
	"github.com/nutochk/13-07-25/internal/repository"
	"github.com/nutochk/13-07-25/internal/service"
	"github.com/nutochk/13-07-25/internal/transport"
	pkg_logger "github.com/nutochk/13-07-25/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	logger, err := pkg_logger.NewZapLogger()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.NewConfig("config/config.yaml")
	if err != nil {
		logger.Fatal("failed to read config", zap.Error(err))
	}

	repo := repository.NewTaskRepository()

	apiService := service.NewTaskService(repo, logger, *cfg)
	apiServer := transport.NewHttpServer(apiService)

	go func() {
		logger.Info("Server is listening on port:" + strconv.Itoa(cfg.Port))
		if err = apiServer.Run(cfg.Port); err != nil {
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down server...")
	apiServer.Shutdown(ctx)
	logger.Info("Server shut down")
}
