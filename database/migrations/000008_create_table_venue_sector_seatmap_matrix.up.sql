CREATE TABLE IF NOT EXISTS venue_sector_seatmap_matrix(
    id serial primary key,
    sector_id uuid NOT NULL REFERENCES venue_sectors(id) ON DELETE CASCADE ON UPDATE CASCADE,
    seat_row integer not null,
    seat_column integer not null,
    label varchar(50),
    status varchar(255),
    
    created_at timestamptz not null default CURRENT_TIMESTAMP,
    updated_at timestamptz
);


DO $$
DECLARE
    x INT;
    y INT;
BEGIN
    FOR x IN 1..15 LOOP
        FOR y IN 1..12 LOOP
            INSERT INTO venue_sector_seatmap_matrix (sector_id, seat_row, seat_column, label, status, created_at)
            VALUES (
                '495c79ef-65a0-43ee-ade8-fe5dba6883aa',
                x,
                y,
                CONCAT('R', x, 'C', y),  
                'AVAILABLE',
                NOW()
            );
        END LOOP;
    END LOOP;
END $$;