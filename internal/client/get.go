package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/localstore"
)

func runGet(cfg *config.OptionsClient, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: get <id>")
	}
	id := args[0]

	store, err := localstore.Open(dbPath(cfg))
	if err != nil {
		return err
	}
	defer store.Close()

	// 1. Local read
	sec, err := store.GetSecret(id)
	if errors.Is(err, localstore.ErrNotCached) {
		return fmt.Errorf("secret %s not found locally (try syncing)", id)
	}
	if err != nil {
		return err
	}

	// Token is read from a local file
	token, err := loadToken(cfg.TokenFile.Value)
	if err != nil {
		return fmt.Errorf("load token (are you logged in?): %w", err)
	}

	masterPw, err := readPassword("master password: ")
	if err != nil {
		return err
	}
	salt, err := fetchSalt(cfg, store, token) // cache hit, no network
	if err != nil {
		return err
	}
	key := DeriveKey(masterPw, salt)

	plaintext, err := Decrypt(key, sec.Payload)
	if err != nil {
		return err
	}

	var label string
	if len(sec.Meta) > 0 {
		metaPlain, err := Decrypt(key, sec.Meta)
		if err != nil {
			return err
		}
		label = string(metaPlain)
	}

	return displaySecret(domain.SecretType(sec.Type), plaintext, label)
}

func displaySecret(t domain.SecretType, plaintext []byte, label string) error {
	if label != "" {
		fmt.Println("label:", label)
	}
	fmt.Println("type: ", t)

	switch t {
	case domain.SecretTypeLoginPassword:
		var lp LoginPassword
		if err := json.Unmarshal(plaintext, &lp); err != nil {
			return err
		}
		fmt.Println("login:   ", lp.Login)
		fmt.Println("password:", lp.Password)

	case domain.SecretTypeText:
		// text is just the raw bytes.
		fmt.Println("text:", string(plaintext))

	case domain.SecretTypeCard:
		var c Card
		if err := json.Unmarshal(plaintext, &c); err != nil {
			return err
		}
		fmt.Println("number:", c.Number)
		fmt.Println("holder:", c.Holder)
		fmt.Println("expiry:", c.Expiry)
		fmt.Println("cvv:   ", c.CVV)

	case domain.SecretTypeBinary:
		_, err := os.Stdout.Write(plaintext)
		return err

	default:
		fmt.Println(string(plaintext))
	}
	return nil
}
