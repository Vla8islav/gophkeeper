-- +goose Up

CREATE TABLE secrets
(
    id         UUID PRIMARY KEY,
    user_id    BIGINT        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type       VARCHAR(32) NOT NULL,
    payload    BYTEA       NOT NULL,
    meta       BYTEA,
    version    BIGINT      NOT NULL DEFAULT 1,
    deleted    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT secrets_type_check CHECK ( type IN ('login_password', 'text', 'binary', 'card') )
);

CREATE INDEX idx_secrets_user_id ON secrets (user_id);

-- +goose Down
DROP TABLE secrets;
