--- Per-account token TTL, in seconds. NULL means "use the default TTL".
ALTER TABLE service_accounts ADD COLUMN ttl_seconds BIGINT;
