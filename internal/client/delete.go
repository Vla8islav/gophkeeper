package client

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/localstore"
)

// runDelete removes a secret: delete on the server (write-through), then drop it
func runDelete(cfg *config.OptionsClient, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: delete <id>")
	}
	id, err := uuid.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid secret id: %w", err)
	}

	token, err := loadToken(cfg.TokenFile.Value)
	if err != nil {
		return fmt.Errorf("load token (are you logged in?): %w", err)
	}

	api, err := NewAPIClient(cfg.ServerAddress.Value, cfg.CACertPath.Value)
	if err != nil {
		return err
	}
	if err = api.DeleteSecret(token, id); err != nil {
		return err
	}

	store, err := localstore.Open(dbPath(cfg))
	if err != nil {
		return err
	}
	defer store.Close()
	if err := store.RemoveSecret(id.String()); err != nil {
		return err
	}

	fmt.Println("deleted", id)
	return nil
}
