package localstore

import "fmt"

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
