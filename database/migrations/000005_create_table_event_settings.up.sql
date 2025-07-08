-- Settings
-- 1. IS_GARUDA_ID_VERIFICATION_ACTIVE
-- 2. MAX_ADULT_TICKET_PURCHASE_PER_TRANSACTION

CREATE TABLE IF NOT EXISTS event_settings (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    setting_id uuid not null REFERENCES settings(id) ON DELETE CASCADE ON UPDATE CASCADE,
    event_id uuid not null REFERENCES events(id) ON DELETE CASCADE ON UPDATE CASCADE,
    setting_value varchar(255),
    created_at timestamptz not null default CURRENT_TIMESTAMP,
    updated_at timestamptz,
    deleted_at timestamptz
);

-- Insert test event
INSERT INTO event_settings (
    id,
    setting_id,
    event_id,
    setting_value,
    created_at
) VALUES (
    '971bbcae-f34b-4913-bc0e-cccfe53d5ba8',
    'e148f94c-4420-424b-b636-15cfc12561c5',
    '77797e23-a2b7-40bd-b8b0-ef628568f815',
    'true',
    NOW()
), (
    '302b6864-b63e-4c9b-8ac7-fc57007c3c1e',
    '74aea68a-2b18-46d1-8257-1f8be10a6e73',
    '77797e23-a2b7-40bd-b8b0-ef628568f815',
    '4',
    NOW()
);