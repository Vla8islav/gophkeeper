package client

import (
	"errors"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/localstore"
)

func fetchSalt(cfg *config.OptionsClient, store *localstore.Store, token string) ([]byte, error) {
	// Cache hit — no network, works offline.
	salt, err := store.Salt()
	if err == nil {
		return salt, nil
	}
	if !errors.Is(err, localstore.ErrNotCached) {
		return nil, err
	}

	// Cache miss — fetch once from the server, then persist.
	api, err := NewAPIClient(cfg.ServerAddress.Value, cfg.CACertPath.Value)
	if err != nil {
		return nil, err
	}
	salt, err = api.GetUserSalt(token)
	if err != nil {
		return nil, err
	}
	if err := store.SaveSalt(salt); err != nil {
		return nil, err
	}
	return salt, nil
}
