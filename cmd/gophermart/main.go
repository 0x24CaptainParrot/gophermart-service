package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/0x24CaptainParrot/gophermart-service/internal/config"
	"github.com/0x24CaptainParrot/gophermart-service/internal/logger"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/handlers"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/service"
	"github.com/0x24CaptainParrot/gophermart-service/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.ParseCfg()
	if err := logger.NewZapLogger("info"); err != nil {
		log.Fatalf("logger innit failed: %v", err)
	}
	defer logger.Log.Sync()

	db, err := repository.NewPostgresDB(cfg.DBUri)
	if err != nil {
		logger.Log.Sugar().Fatalf("failed to initialize db: %s", err.Error())
	}
	defer db.Close()

	if err := repository.ApplyMigrations(db, utils.GetMigrationsPath()); err != nil {
		logger.Log.Sugar().Fatalf("failed to apply migrations: %s", err.Error())
	}

	workerPoolConfig, err := pgxpool.ParseConfig(cfg.DBUri)
	if err != nil {
		logger.Log.Sugar().Errorf("failed to init config for worker pool: %v", err)
	}
	workerPoolConfig.MaxConns = 12

	workerPool, err := pgxpool.NewWithConfig(context.Background(), workerPoolConfig)
	if err != nil {
		logger.Log.Sugar().Errorf("failed to init worker pool: %v", err)
	}
	defer workerPool.Close()

	orderProcessing, err := service.NewOrderProcessingService(workerPool, cfg.AccrualAddr, "order_notifications")
	if err != nil {
		logger.Log.Sugar().Fatalf("failed to init order processing service: %v", err)
	}
	defer orderProcessing.StopProcessing()

	orderProcessing.StartProcessing(context.Background(), 1)

	repos := repository.NewRepository(db)
	services := service.NewService(repos.Authorization, repos.Order, repos.Balance, orderProcessing)
	handler := handlers.NewHandler(cfg, services)
	srv := &handlers.Server{}

	go func() {
		logger.Sugar.Infof("starting server on %s", cfg.RunAddr)
		if err := srv.Run(cfg.RunAddr, handler.InitAPIRoutes()); err != nil {
			logger.Log.Sugar().Fatal("error occured on server:", err)
		}
	}()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, syscall.SIGINT, syscall.SIGTERM)
	<-quitCh
	logger.Log.Sugar().Infoln("gophermart service shutting down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Log.Sugar().Fatalf("error occured on server while shutting down: %s", err.Error())
	}
}
