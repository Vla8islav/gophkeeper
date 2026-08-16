-- +goose Up
ALTER TABLE users ADD COLUMN kdf_salt BYTEA;

-- +goose Down
ALTER TABLE users DROP COLUMN kdf_salt;