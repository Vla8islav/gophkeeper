package client

import (
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/localstore"
)

// runSync pulls the full server state and reconciles the local cache. No master
func runSync(cfg *config.OptionsClient, args []string) error {
	token, err := loadToken(cfg.TokenFile.Value)
	if err != nil {
		return fmt.Errorf("load token (are you logged in?): %w", err)
	}

	api, err := NewAPIClient(cfg.ServerAddress.Value, cfg.CACertPath.Value)
	if err != nil {
		return err
	}
	items, err := api.SyncSecrets(token)
	if err != nil {
		return err
	}

	store, err := localstore.Open(dbPath(cfg))
	if err != nil {
		return err
	}
	defer store.Close()

	pulled, removed, err := reconcile(store, items)
	if err != nil {
		return err
	}

	fmt.Printf("synced: %d updated, %d removed\n", pulled, removed)
	return nil
}

// reconcile matches the local cache to the server state
func reconcile(store *localstore.Store, items []domain.SyncSecretResponse) (pulled, removed int, err error) {
	for _, it := range items {
		if it.Deleted {
			if err := store.RemoveSecret(it.ID.String()); err != nil {
				return pulled, removed, err
			}
			removed++
			continue
		}
		if err := store.SaveSecret(localstore.Secret{
			ID:      it.ID.String(),
			Type:    string(it.Type),
			Payload: it.Payload,
			Meta:    it.Meta,
			Version: it.Version,
			Dirty:   false,
		}); err != nil {
			return pulled, removed, err
		}
		pulled++
	}
	return pulled, removed, nil
}
