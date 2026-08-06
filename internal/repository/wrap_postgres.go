package repository

import (
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/domain"
)

func WrapPostgres(currentConfig *config.OptionsServer) (domain.GophkeeperRepository, error) {
	var db domain.GophkeeperRepository
	var err error

	// Case 1
	if currentConfig.DatabaseURI.BeenSet {
		db, err = NewPostgresStorage(currentConfig, currentConfig.MigrationsFolder.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize metrics repository: %w", err)
		}
		return db, nil
	}

	return nil, fmt.Errorf("something strange happened: " +
		"restore and connection string parameters are incorrect")
}
