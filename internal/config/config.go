package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
)

type Config struct {
	RunAddr         string `env:"RUN_ADDRESS"`
	DBUri           string `env:"DATABASE_URI"`
	AccrualAddr     string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	WorkerPoolConns int    `env:"WP_CONNS"`
	Workers         int    `env:"WORKERS"`
}

func ParseCfg() *Config {
	var cfg Config
	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "run address")
	flag.StringVar(&cfg.DBUri, "d", "", "database uri")
	flag.StringVar(&cfg.AccrualAddr, "r", "", "accrual system address")
	flag.IntVar(&cfg.WorkerPoolConns, "c", 12, "max connns for worker pool")
	flag.IntVar(&cfg.Workers, "w", 1, "total workers")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Printf("error occured parsing env variables: %v", err)
	}

	log.Println("Running with:")
	log.Printf("RunAddr: %s", cfg.RunAddr)
	log.Printf("DBUri: %s", cfg.DBUri)
	log.Printf("AccrualAddr: %s", cfg.AccrualAddr)

	return &cfg
}
