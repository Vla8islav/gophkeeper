package client

import (
	"path/filepath"

	"github.com/Vla8islav/gophkeeper/internal/config"
)

func dbPath(cfg *config.OptionsClient) string {
	return filepath.Join(filepath.Dir(cfg.TokenFile.Value), "store.db")
}
