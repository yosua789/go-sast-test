CREATE TABLE IF NOT EXISTS payment_methods (
    id serial primary key,
    
    logo text,
    name varchar(255) not null, -- It will be shown to user

    is_active boolean default false,
    is_paused boolean default false,
    pause_message text default '',

    payment_type varchar(255) not null, -- provider bank, like BRI, BCA
    payment_group varchar(50) not null, -- for grouping usage, like `Virtual Account`, `Others`
    payment_code varchar(50) not null, -- payment code for integration with payment gateway
    payment_channel varchar(255) not null, -- payment gateway provider like 'Paylabs'

    created_at timestamptz not null,
    updated_at timestamptz,
    paused_at timestamptz
);

ALTER TABLE payment_methods ADD CONSTRAINT payment_method_code_unique UNIQUE (payment_channel, payment_code);
CREATE INDEX payment_methods_code ON payment_methods (payment_code);

INSERT INTO payment_methods (
    logo,
    name,
    
    is_active,
    is_paused,
    pause_message,
    
    payment_type,
    payment_group,
    payment_code,
    payment_channel,

    created_at
) VALUES (
    'mandiri-va.png',
    'Mandiri Virtual Account',

    true,
    false,
    '',

    'Mandiri',
    'Virtual Account',
    'MandiriVA',
    'Paylabs',

    NOW()
), (
    'qris.png',
    'QRIS',

    true,
    false,
    '',

    'QRIS',
    'Others',
    'QRIS',
    'Paylabs',

    NOW()
);