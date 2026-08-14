package run

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/handler"
	"github.com/Vla8islav/gophkeeper/internal/middlewares"
	"github.com/Vla8islav/gophkeeper/internal/service"
	"go.uber.org/zap"
)

func Run(ctx context.Context, db domain.GophkeeperRepository, cfg *config.OptionsServer, logger *zap.Logger) error {

	srvApp := service.NewMetricsService(db,
		cfg.AuthTokenSecret.Value)

	h := handler.NewHandler(srvApp, logger)
	r := handler.NewRouter(h, cfg)

	handlerWithMW := middlewares.ChainMiddlewares(
		r,
		middlewares.WithLogging(logger),
	)

	srv := &http.Server{
		Addr:         cfg.ServerAddress.Value,
		Handler:      handlerWithMW,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	err := srv.ListenAndServeTLS(cfg.PublicCertKey.Value, cfg.PrivateKey.Value)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
