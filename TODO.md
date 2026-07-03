# TODO

User gets into the system and does authentication via Google OAuth.
Then user can manage its own workspaces, roles, accounts, and bindings.
Services will use the services API to validate their tokens.

* [X] Database models
    - User: email
    - Workspace: id, user_email
    - Role: id, name, description, workspace_id
    - Account: id, name, description, workspace_id
    - Account Roles: role_id, account_id
* [X] Database migrations
* [ ] Database ORM (CRUD)
* [ ] Base HTTP service
    - /api/users [POST, GET]
    - /api/services [POST, PUT, GET]
    - /api/workspace [POST, PUT, GET, DEL]
    - /api/roles [POST, PUT, GET, DEL]
    - /api/accounts [POST, PUT, GET, DEL]
    - /api/role-bindings [POST, PUT, GET, DEL]
* [ ] Simple user interface
