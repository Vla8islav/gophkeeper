package client

import (
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/config"
)

func runList(cfg *config.OptionsClient, args []string) error {
	token, err := loadToken(cfg.TokenFile.Value)
	if err != nil {
		return fmt.Errorf("load token (are you logged in?): %w", err)
	}
	api, err := NewAPIClient(cfg.ServerAddress.Value, cfg.CACertPath.Value)
	if err != nil {
		return err
	}
	secrets, err := api.ListSecrets(token)
	if err != nil {
		return err
	}
	if len(secrets) == 0 {
		fmt.Println("no secrets")
		return nil
	}
	for _, s := range secrets {
		fmt.Printf("%s  %-14s  v%d\n", s.ID, s.Type, s.Version)
	}
	return nil
}
