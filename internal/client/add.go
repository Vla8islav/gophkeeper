package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/localstore"
)

func runAdd(cfg *config.OptionsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: add <login_password|text|card|binary> [label]")
	}
	secretType := domain.SecretType(args[0])
	if !secretType.Valid() {
		return fmt.Errorf("unknown type %q (use login_password|text|card|binary)", args[0])
	}
	var label string
	if len(args) >= 2 {
		label = args[1]
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

	masterPw, err := readPassword("master password: ")
	if err != nil {
		return err
	}
	salt, err := fetchSalt(cfg, store, token)
	if err != nil {
		return err
	}
	key := DeriveKey(masterPw, salt)

	plaintext, err := gatherPlaintext(secretType)
	if err != nil {
		return err
	}
	payloadCipher, err := Encrypt(key, plaintext)
	if err != nil {
		return err
	}

	var metaCipher []byte
	if label != "" {
		metaCipher, err = Encrypt(key, []byte(label))
		if err != nil {
			return err
		}
	}

	id := uuid.New()

	api, err := NewAPIClient(cfg.ServerAddress.Value, cfg.CACertPath.Value)
	if err != nil {
		return err
	}
	if err := api.CreateSecret(token, domain.CreateSecretRequest{
		ID:      id,
		Type:    secretType,
		Payload: payloadCipher,
		Meta:    metaCipher,
	}); err != nil {
		return err
	}

	// Cache the ciphertext locally (version 1, not dirty - already pushed).
	if err := store.SaveSecret(localstore.Secret{
		ID:      id.String(),
		Type:    string(secretType),
		Payload: payloadCipher,
		Meta:    metaCipher,
		Version: 1,
		Dirty:   false,
	}); err != nil {
		return err
	}

	fmt.Println("added", id)
	return nil
}

func gatherPlaintext(t domain.SecretType) ([]byte, error) {
	switch t {
	case domain.SecretTypeLoginPassword:
		login, err := readLine("login: ")
		if err != nil {
			return nil, err
		}
		password, err := readPassword("secret password: ")
		if err != nil {
			return nil, err
		}
		return json.Marshal(LoginPassword{Login: login, Password: password})

	case domain.SecretTypeText:
		text, err := readLine("text: ")
		if err != nil {
			return nil, err
		}
		return []byte(text), nil

	case domain.SecretTypeCard:
		number, err := readLine("card number: ")
		if err != nil {
			return nil, err
		}
		holder, err := readLine("card holder: ")
		if err != nil {
			return nil, err
		}
		expiry, err := readLine("expiry (MM/YY): ")
		if err != nil {
			return nil, err
		}
		cvv, err := readPassword("cvv: ") // sensitive - no echo
		if err != nil {
			return nil, err
		}
		return json.Marshal(Card{Number: number, Holder: holder, Expiry: expiry, CVV: cvv})

	case domain.SecretTypeBinary:
		path, err := readLine("file path: ")
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", path, err)
		}
		return data, nil

	default:
		return nil, fmt.Errorf("unsupported type %q", t)
	}
}
