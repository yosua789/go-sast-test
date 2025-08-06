BEGIN;

-- Drop the nullable foreign keys
ALTER TABLE event_ticket_categories
DROP CONSTRAINT event_ticket_categories_venue_sector_id_fkey;

ALTER TABLE event_ticket_categories
DROP CONSTRAINT event_ticket_categories_event_id_fkey;

-- Set columns back to NOT NULL
ALTER TABLE event_ticket_categories
ALTER COLUMN event_id SET NOT NULL;

ALTER TABLE event_ticket_categories
ALTER COLUMN venue_sector_id SET NOT NULL;

-- Add the original constraints back
ALTER TABLE event_ticket_categories
ADD CONSTRAINT event_ticket_categories_venue_sector_id_fkey
FOREIGN KEY (venue_sector_id)
REFERENCES venue_sectors(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

ALTER TABLE event_ticket_categories
ADD CONSTRAINT event_ticket_categories_event_id_fkey
FOREIGN KEY (event_id)
REFERENCES events(id)
ON DELETE CASCADE
ON UPDATE CASCADE;

COMMIT;
