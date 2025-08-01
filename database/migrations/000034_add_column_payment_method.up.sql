ALTER TABLE payment_methods ADD COLUMN is_percentage boolean NOT NULL DEFAULT false;
ALTER TABLE payment_methods ADD COLUMN additional_fee float8 NOT NULL DEFAULT 0;