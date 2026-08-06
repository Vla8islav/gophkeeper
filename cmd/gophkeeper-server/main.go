package main

import (
	"context"
	"log"
	"os"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/repository"
	"github.com/Vla8islav/gophkeeper/internal/run"
	"go.uber.org/zap"
)

func main() {
	lg, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize lg: %v", err)
	}
	defer lg.Sync() // flushes buffer, if any

	currentConfig := config.ReadFlagsServer(os.Args[1:])
	if currentConfig == nil {
		lg.Fatal("failed to read config")
		return
	}
	lg.Info("starting server ", zap.String("Server addr", currentConfig.ServerAddress.Value))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := repository.WrapPostgres(currentConfig)
	if err != nil {
		lg.Fatal("init db: " + err.Error())
	}

	err = run.Run(ctx, db, currentConfig, lg)
	if err != nil {
		lg.Fatal("failed to start server", zap.Error(err))
		return
	}

}
