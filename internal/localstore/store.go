package localstore

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // registers the "sqlite" driver (pure Go, no cgo)
)

// ErrNotCached means the requested value isn't in the local store yet
// (e.g. the salt hasn't been fetched from the server on this device).
var ErrNotCached = errors.New("value not cached locally")

const metaKeySalt = "kdf_salt"

type Store struct {
	db *sql.DB
}

// Open opens - creating if needed - the local SQLite store and applies the schema
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	s := &Store{db: db}
	if err := s.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) init() error {
	_, err := s.db.Exec(`
                CREATE TABLE IF NOT EXISTS meta (
                        key   TEXT PRIMARY KEY,
                        value BLOB
                );
                CREATE TABLE IF NOT EXISTS secrets (
                        id         TEXT PRIMARY KEY,
                        type       TEXT    NOT NULL,
                        payload    BLOB    NOT NULL,   -- ciphertext on a server
                        meta       BLOB,               -- ciphertext (nullable)
                        version    INTEGER NOT NULL,
                        deleted    INTEGER NOT NULL DEFAULT 0,
                        updated_at TEXT,
                        dirty      INTEGER NOT NULL DEFAULT 0  -- local-only: has an un-pushed change
                );
        `)
	if err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// SaveSalt stores or replaces cached KDF salt.
func (s *Store) SaveSalt(salt []byte) error {
	_, err := s.db.Exec(
		`INSERT INTO meta (key, value) VALUES (?, ?)
                 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		metaKeySalt, salt,
	)
	if err != nil {
		return fmt.Errorf("save salt: %w", err)
	}
	return nil
}

// Salt returns the cached KDF salt, or ErrNotCached if it hasn't been fetched yet.
func (s *Store) Salt() ([]byte, error) {
	var salt []byte
	err := s.db.
		QueryRow(`SELECT value FROM meta WHERE key = ?`, metaKeySalt).Scan(&salt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotCached
	}
	if err != nil {
		return nil, fmt.Errorf("read salt: %w", err)
	}
	return salt, nil
}

func (s *Store) RemoveSecret(id string) error {
	_, err := s.db.Exec(`DELETE FROM secrets WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("remove secret: %w", err)
	}
	return nil
}
