--- 000001_schema.up.sql

--- Users table
CREATE TABLE users (
    email      TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--- Workspace table
CREATE TABLE workspaces (
    id         SERIAL PRIMARY KEY,
    user_email TEXT NOT NULL REFERENCES users (email) ON DELETE CASCADE
);

--- Roles table
CREATE TABLE roles (
    id           SERIAL PRIMARY KEY,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    workspace_id BIGINT NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE
);

--- Accounts table
CREATE TABLE accounts (
    id           SERIAL PRIMARY KEY,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    workspace_id BIGINT NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE
);

--- Account & Roles table
CREATE TABLE account_roles (
    role_id    BIGINT NOT NULL REFERENCES roles (id) ON DELETE CASCADE,
    account_id BIGINT NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, account_id)
);
