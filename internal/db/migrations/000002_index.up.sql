--- 000002_index.up.sql

CREATE INDEX idx_workspaces_user_email ON workspaces (user_email);
CREATE INDEX idx_roles_workspace_id ON roles (workspace_id);
CREATE INDEX idx_accounts_workspace_id ON accounts (workspace_id);
CREATE INDEX idx_account_roles_account_id ON account_roles (account_id);
