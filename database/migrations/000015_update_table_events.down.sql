ALTER TABLE events DROP COLUMN publish_status CASCADE;

CREATE TYPE event_status AS ENUM ('UPCOMING', 'CANCELED', 'POSTPONED', 'FINISHED', 'ON_GOING');

ALTER TABLE events ADD COLUMN status event_status default 'UPCOMING';
ALTER TABLE events ADD COLUMN is_active boolean;