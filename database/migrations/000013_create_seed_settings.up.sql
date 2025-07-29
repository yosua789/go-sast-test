INSERT INTO settings (
    id,
    name,
    default_value,
    created_at 
) VALUES (
    '3ec8a5fa-11fd-4fcd-ab4c-69d5f0ffc96f',
    'TAX_PERCENTAGE',
    '0',
    NOW()
), (
    '8eece242-f1b8-46b9-8e2c-c4c9b28e68c9',
    'ADMIN_FEE_PERCENTAGE',
    '0',
    NOW()
), (
    '8a902407-8c39-446a-930e-c884f29c50a5',
    'ADMIN_FEE_PRICE',
    '0',
    NOW()
);

INSERT INTO event_settings (
    setting_id,
    event_id,
    setting_value,
    created_at
) VALUES (
    '3ec8a5fa-11fd-4fcd-ab4c-69d5f0ffc96f',
    '77797e23-a2b7-40bd-b8b0-ef628568f815',
    '0',
    NOW()
), (
    '8eece242-f1b8-46b9-8e2c-c4c9b28e68c9',
    '77797e23-a2b7-40bd-b8b0-ef628568f815',
    '0',
    NOW()
), (
    '8a902407-8c39-446a-930e-c884f29c50a5',
    '77797e23-a2b7-40bd-b8b0-ef628568f815',
    '0',
    NOW()
);