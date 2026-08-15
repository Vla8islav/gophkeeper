package client

import (
	"errors"
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/config"
)

func runRegister(cfg *config.OptionsClient, args []string) error {
	if len(args) != 2 {
		return errors.New("usage: register <login> <password>")
	}
	api, err := NewAPIClient(cfg.ServerAddress.Value, cfg.CACertPath.Value)
	if err != nil {
		return err
	}
	token, err := api.Register(args[0], args[1])
	if err != nil {
		return err
	}
	if err := saveToken(cfg.TokenFile.Value, token); err != nil {
		return fmt.Errorf("save token: %w", err)
	}
	fmt.Println("registered; token saved to", cfg.TokenFile.Value)
	return nil
}
