-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE campaigns ADD COLUMN encryption_key varchar(255);
ALTER TABLE email_requests ADD COLUMN encryption_key varchar(255);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE campaigns DROP COLUMN encryption_key;
ALTER TABLE email_requests DROP COLUMN encryption_key;
