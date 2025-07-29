-- Settings
-- 1. IS_GARUDA_ID_VERIFICATION_ACTIVE
-- 2. MAX_ADULT_TICKET_PURCHASE_PER_TRANSACTION

-- # Removed
-- value_type = BOOLEAN | STRING | INTEGER

CREATE TABLE IF NOT EXISTS settings (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name varchar(255) not null UNIQUE,
    default_value varchar(255) not null,
    created_at timestamptz not null default CURRENT_TIMESTAMP,
    updated_at timestamptz,
    deleted_at timestamptz
);

-- Insert default settings
INSERT INTO settings (
    id,
    name,
    default_value,
    created_at
) VALUES (
    'e148f94c-4420-424b-b636-15cfc12561c5',
    'IS_GARUDA_ID_VERIFICATION_ACTIVE',
    'false',
    NOW()
), (
    '74aea68a-2b18-46d1-8257-1f8be10a6e73',
    'MAX_ADULT_TICKET_PURCHASE_PER_TRANSACTION',
    '1',
    NOW()
);