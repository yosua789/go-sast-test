CREATE TABLE IF NOT EXISTS event_venue_sector_seatmap_matrix(
    id serial primary key,
    event_id uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE ON UPDATE CASCADE,
    sector_id uuid NOT NULL REFERENCES venue_sectors(id) ON DELETE CASCADE ON UPDATE CASCADE,
    seat_row integer not null,
    seat_column integer not null,
    label varchar(50),
    status varchar(255),
    
    created_at timestamptz not null default CURRENT_TIMESTAMP,
    updated_at timestamptz
);

INSERT INTO event_venue_sector_seatmap_matrix (
    event_id,
    sector_id, 
    seat_row, 
    seat_column, 
    status, 
    created_at
) VALUES (
    '77797e23-a2b7-40bd-b8b0-ef628568f815',
    '495c79ef-65a0-43ee-ade8-fe5dba6883aa',
    5,
    5,
    'UNAVAILABLE',
    NOW()
), (
    'aabeaad8-8725-41f9-bd8f-3e6baff86573',
    '495c79ef-65a0-43ee-ade8-fe5dba6883aa',
    4,
    5,
    'UNAVAILABLE',
    NOW()
);