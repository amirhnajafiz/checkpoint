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
* [X] Database ORM (CRUD)
* [X] Base HTTP service (using Echo)
    - /api/users [POST]
      - POST users gets email and redirects to Google OAuth, creates a new user if not exists
    - /api/services [POST, PUT, GET]
      - FUTURE
    - /api/workspace [POST, PUT, GET, DEL]
    - /api/roles [POST, PUT, GET, DEL]
    - /api/accounts [POST, PUT, GET, DEL]
    - /api/role-bindings [POST, PUT, GET, DEL]
* [ ] Simple user interface (using Go tmpl)
* [ ] Zap JSON format logs
* [ ] OTel tracing
* [ ] Prometheus metrics
* [ ] Dockerfile + Compose deployment
