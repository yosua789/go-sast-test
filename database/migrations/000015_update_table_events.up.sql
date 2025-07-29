ALTER TABLE events DROP COLUMN status CASCADE;
ALTER TABLE events DROP COLUMN is_active CASCADE;

DROP TYPE event_status;

ALTER TABLE events ADD COLUMN publish_status varchar(50) DEFAULT 'DRAFT';