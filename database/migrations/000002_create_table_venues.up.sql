CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create location type enum
CREATE TYPE location_type AS ENUM ('STADIUM', 'VENUE', 'HALL', 'OTHER');

CREATE TYPE venue_status AS ENUM ('ACTIVE', 'INACTIVE', 'DISABLE');

CREATE TABLE IF NOT EXISTS venues (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    venue_type location_type not null,
    name varchar(255) not null,
    country varchar(255) not null,
    city varchar(255) not null,
    status venue_status not null,
    capacity int,
    created_at timestamptz not null default CURRENT_TIMESTAMP,
    updated_at timestamptz,
    deleted_at timestamptz
);

-- Dummy stadium
INSERT INTO
    venues (
        id,
        venue_type,
        name,
        country,
        city,
        status,
        capacity,
        created_at,
        updated_at,
        deleted_at
    )
VALUES
    (
        'fa6b76af-fcf2-4a04-b63d-d933c61905d8',
        'STADIUM',
        'Gelora Bungkarno - Test',
        'Indonesia',
        'Jakarta',
        'ACTIVE',
        70000,
        NOW(),
        NOW(), 
        NULL
    );