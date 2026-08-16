package main

import (
	"context"
	"log"
	"os"

	_ "github.com/Vla8islav/gophkeeper/docs" // generated OpenAPI spec (swag init)
	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/repository"
	"github.com/Vla8islav/gophkeeper/internal/run"
	"go.uber.org/zap"
)

// @title           GophKeeper API
// @version         1.0
// @description     End-to-end-encrypted secrets manager. Secrets are encrypted on the client (Argon2id + AES-256-GCM); the server stores only ciphertext it cannot read.
// @BasePath        /
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Paste "Bearer <token>" - the token returned by /api/user/login or /api/user/register.
func main() {
	lg, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize lg: %v", err)
	}
	defer lg.Sync() // flushes buffer, if any

	currentConfig, err := config.ReadFlagsServer(os.Args[1:], lg)
	if err != nil {
		lg.Fatal("failed to read config" + err.Error())
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
