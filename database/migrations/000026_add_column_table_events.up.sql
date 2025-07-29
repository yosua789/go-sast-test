ALTER TABLE events ADD COLUMN published_at timestamptz;
ALTER TABLE events ADD COLUMN paused_at timestamptz;