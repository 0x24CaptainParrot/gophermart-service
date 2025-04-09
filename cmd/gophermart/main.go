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
)

func main() {
	cfg := config.ParseCfg()
	if err := logger.NewZapLogger("info"); err != nil {
		log.Fatalf("logger innit failed: %v", err)
	}
	defer logger.Log.Sync()

	db, err := repository.NewPostgresDB(cfg.DBUri)
	if err != nil {
		log.Fatalf("failed to initialize db: %s", err.Error())
	}
	defer db.Close()

	if err := repository.ApplyMigrations(db, utils.GetMigrationsPath()); err != nil {
		log.Fatalf("failed to apply migrations: %s", err.Error())
	}

	repos := repository.NewRepository(db)
	services := service.NewService(repos.Authorization)
	handler := handlers.NewHandler(services)
	srv := &handlers.Server{}

	go func() {
		logger.Sugar.Infof("starting server on %s", cfg.RunAddr)
		if err := srv.Run(cfg.RunAddr, handler.InitAPIRoutes()); err != nil {
			log.Fatal("error occured on server:", err)
		}
	}()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, syscall.SIGINT, syscall.SIGTERM)
	<-quitCh
	log.Println("gophermart service shutting down")

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("error occured on server while shutting down: %s", err.Error())
	}
}
