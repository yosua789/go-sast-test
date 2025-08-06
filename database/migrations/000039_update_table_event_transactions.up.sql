ALTER TABLE event_transactions ADD COLUMN invoice_status VARCHAR(50) NOT NULL DEFAULT 'IN PROGRESS';
ALTER TABLE event_transactions ADD COLUMN invoice_message TEXT;
ALTER TABLE event_transactions ADD COLUMN invoice_generated_at timestamptz;
ALTER TABLE event_transactions ADD COLUMN invoice_filename text;