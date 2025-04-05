package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/0x24CaptainParrot/gophermart-service/internal/config"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/handlers"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
)

func main() {
	cfg := config.ParseCfg()

	db, err := repository.NewPostgresDB(cfg.DBUri)
	if err != nil {
		log.Fatalf("failed to initialize db: %s", err.Error())
	}
	defer db.Close()

	if err := repository.ApplyMigrations(db, "../../internal/pkg/repository/schema"); err != nil {
		log.Fatalf("failed to apply migrations: %s", err.Error())
	}

	handler := handlers.NewHandler()
	srv := &handlers.Server{}

	go func() {
		if err := srv.Run(cfg.RunAddr, handler.InitApiRoutes()); err != nil {
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
