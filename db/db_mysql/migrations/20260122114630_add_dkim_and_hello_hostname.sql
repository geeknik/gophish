-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
-- Add DKIM and HelloHostname support to SMTP profiles
ALTER TABLE smtp ADD COLUMN dkim_enabled BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE smtp ADD COLUMN dkim_domain VARCHAR(255);
ALTER TABLE smtp ADD COLUMN dkim_selector VARCHAR(255);
ALTER TABLE smtp ADD COLUMN dkim_private_key TEXT;
ALTER TABLE smtp ADD COLUMN hello_hostname VARCHAR(255);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE smtp DROP COLUMN dkim_enabled;
ALTER TABLE smtp DROP COLUMN dkim_domain;
ALTER TABLE smtp DROP COLUMN dkim_selector;
ALTER TABLE smtp DROP COLUMN dkim_private_key;
ALTER TABLE smtp DROP COLUMN hello_hostname;
