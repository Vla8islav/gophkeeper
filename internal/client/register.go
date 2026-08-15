package client

import (
	"errors"
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/config"
)

func runRegister(cfg *config.OptionsClient, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: register <login>")
	}
	login := args[0]

	password, err := readPassword("password: ")
	if err != nil {
		return err
	}
	if password == "" {
		return errors.New("password cannot be empty")
	}
	confirm, err := readPassword("confirm password: ")
	if err != nil {
		return err
	}
	if password != confirm {
		return errors.New("passwords do not match")
	}

	api, err := NewAPIClient(cfg.ServerAddress.Value, cfg.CACertPath.Value)
	if err != nil {
		return err
	}
	token, err := api.Register(login, password)
	if err != nil {
		return err
	}
	if err := saveToken(cfg.TokenFile.Value, token); err != nil {
		return fmt.Errorf("save token: %w", err)
	}
	fmt.Println("registered; token saved to", cfg.TokenFile.Value)
	return nil
}
