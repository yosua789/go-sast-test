CREATE TYPE event_status AS ENUM ('UPCOMING', 'CANCELED', 'POSTPONED', 'FINISHED', 'ON_GOING');

CREATE TABLE IF NOT EXISTS events (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    organizer_id uuid not null REFERENCES organizers(id) ON DELETE CASCADE ON UPDATE CASCADE,
    name varchar(500) not null UNIQUE,
    description text,
    banner text not null,
    event_time timestamptz not null,
    status event_status not null,
    venue_id uuid not null REFERENCES venues(id) ON DELETE CASCADE ON UPDATE CASCADE,

    additional_information text,

    is_active boolean,
    
    start_sale_at timestamptz,
    end_sale_at timestamptz,

    created_at timestamptz not null default CURRENT_TIMESTAMP,
    updated_at timestamptz,
    deleted_at timestamptz
);

INSERT INTO
    events (
        id,
        organizer_id,
        name,
        description,
        banner,
        event_time,
        status,
        venue_id,

        additional_information,

        is_active,

        start_sale_at,
        end_sale_at,

        created_at,
        updated_at,
        deleted_at
    )
VALUES
    (
        '77797e23-a2b7-40bd-b8b0-ef628568f815',
        '160ec557-880f-4167-a30e-1a5dcaee78af',
        'Test Events',
        'Ini hanyalah testing belaka',
        'default-banner.png',
        '2025-07-14 07:37:15.154 +0000',
        'UPCOMING',
        'fa6b76af-fcf2-4a04-b63d-d933c61905d8',

        '# Request Deployment Notification Service',

        true,
        
        '2025-07-02 07:37:15.154 +0000',
        '2025-07-03 07:37:15.154 +0000',

        NOW(),
        NOW(), 
        NULL
    );