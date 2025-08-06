CREATE INDEX IF NOT EXISTS idx_event_sector_created_at 
    ON event_seatmap_books (event_id, venue_sector_id, created_at DESC);

