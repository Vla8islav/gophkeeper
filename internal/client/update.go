package client

import (
	"errors"
	"fmt"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/localstore"
	"github.com/google/uuid"
)

func runUpdate(cfg *config.OptionsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: update <id> [label]")
	}
	id, err := uuid.Parse(args[0])
	if err != nil {
		return fmt.Errorf("invalid secret id: %w", err)
	}
	labelProvided := len(args) >= 2
	var newLabel string
	if labelProvided {
		newLabel = args[1]
	}

	token, err := loadToken(cfg.TokenFile.Value)
	if err != nil {
		return fmt.Errorf("load token (are you logged in?): %w", err)
	}

	store, err := localstore.Open(dbPath(cfg))
	if err != nil {
		return err
	}
	defer store.Close()

	cached, err := store.GetSecret(id.String())
	if errors.Is(err, localstore.ErrNotCached) {
		return fmt.Errorf("secret %s not found locally (try syncing)", id)
	}
	if err != nil {
		return err
	}

	masterPw, err := readPassword("master password: ")
	if err != nil {
		return err
	}
	salt, err := fetchSalt(cfg, store, token)
	if err != nil {
		return err
	}
	key := DeriveKey(masterPw, salt)

	plaintext, err := gatherPlaintext(domain.SecretType(cached.Type))
	if err != nil {
		return err
	}
	payloadCipher, err := Encrypt(key, plaintext)
	if err != nil {
		return err
	}

	metaCipher := cached.Meta
	if labelProvided {
		metaCipher = nil
		if newLabel != "" {
			metaCipher, err = Encrypt(key, []byte(newLabel))
			if err != nil {
				return err
			}
		}
	}

	api, err := NewAPIClient(cfg.ServerAddress.Value, cfg.CACertPath.Value)
	if err != nil {
		return err
	}
	newVersion, err := api.UpdateSecret(token, id, domain.UpdateSecretRequest{
		Payload: payloadCipher,
		Meta:    metaCipher,
		Version: cached.Version, // server rejects if this is stale
	})
	if err != nil {
		return err
	}

	if err := store.SaveSecret(localstore.Secret{
		ID:      id.String(),
		Type:    cached.Type,
		Payload: payloadCipher,
		Meta:    metaCipher,
		Version: newVersion,
		Dirty:   false,
	}); err != nil {
		return err
	}

	fmt.Printf("updated %s : v%d\n", id, newVersion)
	return nil
}
