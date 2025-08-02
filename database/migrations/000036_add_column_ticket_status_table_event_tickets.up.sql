ALTER TABLE event_tickets ADD COLUMN ticket_status VARCHAR(50) NOT NULL DEFAULT 'IN PROGRESS';
ALTER TABLE event_tickets ADD COLUMN ticket_message TEXT;
ALTER TABLE event_tickets ADD COLUMN ticket_generated_at timestamptz;
ALTER TABLE event_tickets ADD COLUMN ticket_filename text;