-- +goose Up
-- Add OAuth 2.0 fields to IMAP table for Microsoft 365 support
ALTER TABLE imap ADD COLUMN use_oauth2 BOOLEAN DEFAULT FALSE;
ALTER TABLE imap ADD COLUMN oauth_tenant_id VARCHAR(255) DEFAULT '';
ALTER TABLE imap ADD COLUMN oauth_client_id VARCHAR(255) DEFAULT '';
ALTER TABLE imap ADD COLUMN oauth_client_secret TEXT DEFAULT NULL;

-- +goose Down
ALTER TABLE imap DROP COLUMN use_oauth2;
ALTER TABLE imap DROP COLUMN oauth_tenant_id;
ALTER TABLE imap DROP COLUMN oauth_client_id;
ALTER TABLE imap DROP COLUMN oauth_client_secret;
