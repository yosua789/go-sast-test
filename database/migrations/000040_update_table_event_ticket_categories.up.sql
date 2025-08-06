BEGIN;

ALTER TABLE event_ticket_categories ALTER COLUMN event_id DROP NOT NULL;
ALTER TABLE event_ticket_categories ALTER COLUMN venue_sector_id DROP NOT NULL;

-- Drop the old constraint
ALTER TABLE event_ticket_categories
DROP CONSTRAINT event_ticket_categories_venue_sector_id_fkey;

ALTER TABLE event_ticket_categories
DROP CONSTRAINT event_ticket_categories_event_id_fkey;

-- Add the new constraint
ALTER TABLE event_ticket_categories
ADD CONSTRAINT event_ticket_categories_venue_sector_id_fkey
FOREIGN KEY (venue_sector_id)
REFERENCES venue_sectors(id)
ON DELETE SET NULL;

ALTER TABLE event_ticket_categories
ADD CONSTRAINT event_ticket_categories_event_id_fkey
FOREIGN KEY (event_id)
REFERENCES events(id)
ON DELETE SET NULL;

COMMIT;