package http

// createServiceAccountRequest is the JSON body for POST /api/accounts.
type createServiceAccountRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	// Active is optional; a nil value defaults to true on creation.
	Active *bool `json:"active"`
	// KV are the key/value labels attached to the account (may be empty).
	KV map[string]string `json:"kv"`
}

// updateServiceAccountRequest is the JSON body for PUT /api/accounts/:id.
type updateServiceAccountRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	// Active is optional; a nil value keeps the account active.
	Active *bool `json:"active"`
	// KV replaces the account's labels; an empty map clears them.
	KV map[string]string `json:"kv"`
}
