# TODO

User gets into the system and does authentication via Google OAuth.
Then user can manage its own workspaces, roles, accounts, and bindings.
Services will use the services API to validate their tokens.

* [X] Database models
    - User: email, created_at
    - Service Account: id, name, description, active, created_at
    - Service Account Meta: account_id, last_used, usage
    - Service Account KV: id, account_id, key, value 
* [X] Database migrations
* [X] Database ORM (CRUD)
* [X] Base HTTP service (using Echo)
    - /api/users [POST]
      - POST users gets email and redirects to Google OAuth, creates a new user if not exists
    - /api/services [POST, PUT, GET]
      - FUTURE
    - /api/accounts [POST, PUT, GET, DEL]
* [ ] Simple user interface (using Go tmpl)
* [ ] Zap JSON format logs
* [ ] OTel tracing
* [ ] Prometheus metrics
* [ ] Dockerfile + Compose deployment
