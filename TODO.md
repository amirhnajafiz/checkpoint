# TODO

User gets into the system and does authentication via Google OAuth.
Then user can manage its own workspaces, roles, accounts, and bindings.
Services will use the services API to validate their tokens.

* [ ] Base HTTP service
    - /api/users [POST, GET]
    - /api/services [POST, PUT, GET]
    - /api/workspace [POST, PUT, GET, DEL]
    - /api/roles [POST, PUT, GET, DEL]
    - /api/accounts [POST, PUT, GET, DEL]
    - /api/role-bindings [POST, PUT, GET, DEL]
* [ ] Database models
    - User: email (Google OAuth)
    - Workspace: id, user_email
    - Role: id, name, description, workspace_id
    - Account: id, name, description, workspace_id
    - Binding: role_id, account_id
* [ ] Database migrations
* [ ] Simple user interface

