package localstore

import (
	"database/sql"
	"errors"
	"fmt"
)

type Secret struct {
	ID        string
	Type      string
	Payload   []byte
	Meta      []byte
	Version   int64
	Deleted   bool
	UpdatedAt string
	Dirty     bool
}

func (s *Store) SaveSecret(sec Secret) error {
	_, err := s.db.Exec(
		`INSERT INTO secrets (id, type, payload, meta, version, deleted, updated_at, dirty)
                 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                 ON CONFLICT(id) DO UPDATE SET
                   type=excluded.type, payload=excluded.payload, meta=excluded.meta,
                   version=excluded.version, deleted=excluded.deleted,
                   updated_at=excluded.updated_at, dirty=excluded.dirty`,
		sec.ID, sec.Type, sec.Payload, sec.Meta, sec.Version,
		sec.Deleted, sec.UpdatedAt, sec.Dirty,
	)
	if err != nil {
		return fmt.Errorf("save secret: %w", err)
	}
	return nil
}

func (s *Store) GetSecret(id string) (Secret, error) {
	var sec Secret
	err := s.db.QueryRow(
		`SELECT id, type, payload, meta, version, deleted, updated_at, dirty
                 FROM secrets
                 WHERE id = ? AND deleted = 0`,
		id,
	).Scan(
		&sec.ID, &sec.Type, &sec.Payload, &sec.Meta,
		&sec.Version, &sec.Deleted, &sec.UpdatedAt, &sec.Dirty,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Secret{}, ErrNotCached
	}
	if err != nil {
		return Secret{}, fmt.Errorf("get secret: %w", err)
	}
	return sec, nil
}

func (s *Store) ListSecrets() ([]Secret, error) {
	rows, err := s.db.Query(
		`SELECT id, type, meta, version
                 FROM secrets
                 WHERE deleted = 0
                 ORDER BY updated_at DESC, id`)
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	defer rows.Close()

	var out []Secret
	for rows.Next() {
		var sec Secret
		if err := rows.Scan(&sec.ID, &sec.Type, &sec.Meta, &sec.Version); err != nil {
			return nil, fmt.Errorf("scan secret: %w", err)
		}
		out = append(out, sec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate secrets: %w", err)
	}
	return out, nil
}
