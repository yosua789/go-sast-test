CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS organizers (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(255) not null,
    slug varchar(255) not null,
    logo text,
    created_at timestamptz not null default CURRENT_TIMESTAMP,
    updated_at timestamptz,
    deleted_at timestamptz
);

INSERT INTO
    organizers (
        id,
        name,
        slug,
        logo,
        created_at,
        updated_at,
        deleted_at
    )
VALUES
    (
        '160ec557-880f-4167-a30e-1a5dcaee78af',
        'Default Organizer',
        'default-organizer',
        'organizers/logo-default-organizer.png',
        NOW(),
        NOW(), 
        NULL
    );