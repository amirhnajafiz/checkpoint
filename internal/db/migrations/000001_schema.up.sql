--- Users table
CREATE TABLE users (
    email      TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--- Service accounts table
CREATE TABLE service_accounts (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    user_email  TEXT NOT NULL REFERENCES users (email) ON DELETE CASCADE, 
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--- Service account metadata (1:1 with a service account)
CREATE TABLE service_account_meta (
    account_id INTEGER PRIMARY KEY REFERENCES service_accounts (id) ON DELETE CASCADE,
    last_used  TIMESTAMPTZ,
    usage      BIGINT NOT NULL DEFAULT 0
);

--- Service account key/value store
CREATE TABLE service_account_kv (
    id          SERIAL PRIMARY KEY,
    account_id  INTEGER NOT NULL REFERENCES service_accounts (id) ON DELETE CASCADE,
    xkey        TEXT NOT NULL,
    xvalue      TEXT NOT NULL DEFAULT '',
    UNIQUE (account_id, xkey)
);
