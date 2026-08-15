package client

import (
	"os"
	"path/filepath"
	"strings"
)

func defaultTokenPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gophkeeper_token" // consider parametrizing this one
	}
	return filepath.Join(home, ".gophkeeper", "token")
}

func saveToken(path, token string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil { // r
		return err
	}
	return os.WriteFile(path, []byte(token), 0o600) // rw
}

func loadToken(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
