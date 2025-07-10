-- Ticket catagories for event
-- CODE digunakan untuk membedakan ticket categories yang valuenya manusiawi / dibaca manusia
-- Default price currency is Rupiah
CREATE TABLE IF NOT EXISTS event_ticket_categories (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id uuid not null REFERENCES events(id) ON DELETE CASCADE ON UPDATE CASCADE,
    name varchar(255) not null,
    description text not null,
    price integer not null,
    
    total_stock integer not null default 0,
    total_public_stock integer not null default 0,
    public_stock integer not null default 0,

    total_compliment_stock integer not null default 0,
    compliment_stock integer not null default 0,

    code varchar(255) not null, 
    entrance varchar(255),

    created_at timestamptz not null default CURRENT_TIMESTAMP,
    updated_at timestamptz,
    deleted_at timestamptz
);

-- Insert test event
INSERT INTO event_ticket_categories (
    id,
    event_id,
    name,
    description,
    price,

    total_stock,
    total_public_stock,
    public_stock,
    total_compliment_stock,
    compliment_stock,
    
    code,
    entrance,

    created_at
) VALUES (
    'fefb5c70-25ee-4326-94e4-3aa21a007299',
    '77797e23-a2b7-40bd-b8b0-ef628568f815',
    'Test Ticket categories - VVIP',
    'yaa ticket test VVIP',
    750000,

    100,
    80,
    80,
    20,
    20,

    'VVIP_NORTH',
    'Gate F',
    NOW()
), (
    'f6caa128-ced2-42b1-8d05-d10490023fbb',
    '77797e23-a2b7-40bd-b8b0-ef628568f815',
    'Test Ticket categories - VIP',
    'yaa ticket test VIP',
    600000,

    100,
    80,
    80,
    20,
    20,

    'VIP_NORTH',
    'Gate A',
    NOW()
), (
    '8668df6c-1a0a-4175-9f90-21c4427bb798',
    '77797e23-a2b7-40bd-b8b0-ef628568f815',
    'Test Ticket categories - Reguler',
    'yaa ticket test REGULER',
    350000,

    150,
    100,
    100,
    40,
    40,

    'REGULER_SOUTH',
    'Gate G',
    NOW()
);