-- +goose Up
-- Add OAuth 2.0 fields to IMAP table for Microsoft 365 support
ALTER TABLE imap ADD COLUMN use_oauth2 BOOLEAN DEFAULT 0;
ALTER TABLE imap ADD COLUMN oauth_tenant_id VARCHAR(255) DEFAULT '';
ALTER TABLE imap ADD COLUMN oauth_client_id VARCHAR(255) DEFAULT '';
ALTER TABLE imap ADD COLUMN oauth_client_secret TEXT DEFAULT '';

-- +goose Down
-- SQLite doesn't support DROP COLUMN, so we need to recreate the table
CREATE TABLE imap_backup AS SELECT user_id, enabled, host, port, username, password, tls, ignore_cert_errors, folder, restrict_domain, delete_reported_campaign_email, last_login, modified_date, imap_freq FROM imap;
DROP TABLE imap;
ALTER TABLE imap_backup RENAME TO imap;
