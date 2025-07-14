CREATE TABLE IF NOT EXISTS event_seatmap_books (
    id SERIAL PRIMARY KEY,
    event_id uuid not null references events(id) REFERENCES events(id) ON DELETE CASCADE ON UPDATE CASCADE,
    venue_sector_id uuid not null references venue_sectors(id) ON DELETE CASCADE ON UPDATE CASCADE,
    
    seat_row int not null,
    seat_column int not null,
    
    created_at timestamptz
);

ALTER TABLE event_seatmap_books ADD CONSTRAINT unique_event_seat_book UNIQUE (event_id, venue_sector_id, seat_row, seat_column);

INSERT INTO event_seatmap_books (
    event_id,
    venue_sector_id,

    seat_row,
    seat_column,

    created_at
) VALUES (
    '77797e23-a2b7-40bd-b8b0-ef628568f815',
    '495c79ef-65a0-43ee-ade8-fe5dba6883aa',

    1,
    1,

    NOW()
), (
    '77797e23-a2b7-40bd-b8b0-ef628568f815',
    '495c79ef-65a0-43ee-ade8-fe5dba6883aa',

    2,
    2,

    NOW()
);
