-- Venue sectors
CREATE TABLE IF NOT EXISTS venue_sectors (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    venue_id uuid not null REFERENCES venues(id) ON DELETE CASCADE ON UPDATE CASCADE,

    name varchar(255) not null,
    sector_row int not null,
    sector_column int not null,
    capacity int not null,

    is_active boolean default false,
    is_have_seatmap boolean default false,

    sector_color varchar(10),
    area_code varchar(255),

    created_at timestamptz not null default CURRENT_TIMESTAMP,
    updated_at timestamptz,
    deleted_at timestamptz
);

-- Insert test event
INSERT INTO venue_sectors (
    id,
    venue_id,
    
    name,
    sector_row,
    sector_column,
    capacity,

    is_active,
    is_have_seatmap,

    sector_color,
    area_code,

    created_at
) VALUES (
    '495c79ef-65a0-43ee-ade8-fe5dba6883aa',
    'fa6b76af-fcf2-4a04-b63d-d933c61905d8',

    'Zona 1',
    15,
    12,
    180,

    true,
    false,

    '#ff6b6b',
    'Area A',

    NOW()
), (
    'b64e0a9e-66fd-4d38-9b27-e743430fa4ab',
    'fa6b76af-fcf2-4a04-b63d-d933c61905d8',

    'Zona 2',
    15,
    12,
    180,

    true,
    false,

    '#48dbfb',
    'Area A',

    NOW()
), (
    '3f189cd4-a388-4f21-9957-1bc197733e00',
    'fa6b76af-fcf2-4a04-b63d-d933c61905d8',

    'Zona 3',
    15,
    12,
    180,

    true,
    false,

    '#feca57',
    'Area A',

    NOW()
), (
    'c376feb5-9666-43f4-83d7-16cade72b552',
    'fa6b76af-fcf2-4a04-b63d-d933c61905d8',

    'Zona 4',
    15,
    12,
    180,

    true,
    false,

    '#ff9ff3',
    'Area A',

    NOW()
), (
    '2b71ca23-f33a-443f-8a61-a4c8a8cd5737',
    'fa6b76af-fcf2-4a04-b63d-d933c61905d8',

    'Zona 5',
    15,
    12,
    180,

    true,
    false,

    '#00d2d3',
    'Area B',

    NOW()
), (
    '2c2986fe-727c-48a3-99ff-74b2691d7ca0',
    'fa6b76af-fcf2-4a04-b63d-d933c61905d8',

    'Zona 6',
    15,
    12,
    180,

    true,
    false,

    '#54a0ff',
    'Area B',

    NOW()
), (
    '9a3a3eeb-dd3e-47b9-ae2b-799c6a531ca0',
    'fa6b76af-fcf2-4a04-b63d-d933c61905d8',

    'Zona 7',
    15,
    12,
    180,

    true,
    false,

    '#5f27cd',
    'Area B',

    NOW()
), (
    '26aee19b-6bdf-4c63-84ec-086cf843f501',
    'fa6b76af-fcf2-4a04-b63d-d933c61905d8',

    'Zona 8',
    15,
    12,
    180,

    true,
    false,

    '#c8d6e5',
    'Area B',

    NOW()
), (
    '25c16dc0-c5bd-46e6-945d-d192f4480bec',
    'fa6b76af-fcf2-4a04-b63d-d933c61905d8',

    'Zona 9',
    15,
    12,
    180,

    true,
    false,

    '#576574',
    'Area C',

    NOW()
), (
    'f65eb8c7-d6d6-4760-aaf0-df473ccd07c7',
    'fa6b76af-fcf2-4a04-b63d-d933c61905d8',

    'Zona 10',
    15,
    12,
    180,

    true,
    false,

    '#1dd1a1',
    'Area C',

    NOW()
), (
    '106f6245-a116-4b56-97a3-6278cb0ce3e1',
    'fa6b76af-fcf2-4a04-b63d-d933c61905d8',

    'Zona 11',
    15,
    12,
    180,

    true,
    false,

    '#6ab04c',
    'Area C',

    NOW()
), (
    '8ce2302a-5351-4b19-81ce-0342bad97ef3',
    'fa6b76af-fcf2-4a04-b63d-d933c61905d8',

    'Zona 12',
    15,
    12,
    180,

    true,
    false,

    '#f0932b',
    'Area C',

    NOW()
);