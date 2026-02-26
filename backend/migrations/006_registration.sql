-- +goose Up
CREATE TABLE invitation_codes (
    id         SERIAL PRIMARY KEY,
    code       VARCHAR(50) NOT NULL,
    count      INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

ALTER TABLE users
    ADD COLUMN invitation_code_id    INTEGER REFERENCES invitation_codes(id),
    ADD COLUMN password_last_changed_at TIMESTAMPTZ;

CREATE TABLE password_reset_tokens (
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id),
    token      VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_password_reset_tokens_user ON password_reset_tokens(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_password_reset_tokens_user;
DROP TABLE IF EXISTS password_reset_tokens;
ALTER TABLE users
    DROP COLUMN IF EXISTS invitation_code_id,
    DROP COLUMN IF EXISTS password_last_changed_at;
DROP TABLE IF EXISTS invitation_codes;
