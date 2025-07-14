-- Transaction status
-- PENDING
-- FAILED
-- SUCCESS
-- UNKNOWN
CREATE TABLE IF NOT EXISTS event_transactions (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),

    invoice_number varchar(50) not null unique,

    transaction_status varchar(255) not null default 'UNKNOWN',
    transaction_status_information text,

    payment_method varchar(255) not null,
    payment_channel varchar(255) not null,
    payment_expired_at timestamptz not null,
    
    paid_at timestamptz,

    total_price int not null default 0,
    tax_percentage DECIMAL(5,2),
    total_tax int not null  default 0,
    admin_fee_percentage DECIMAL(5,2),
    total_admin_fee int not null default 0,
    grand_total int not null default 0,

    full_name varchar(255) not null,
    email varchar(255) not null,
    phone_number varchar(255) not null,
    
    is_compliment boolean default false,
    is_refunded boolean default false,

    created_at timestamptz not null default CURRENT_TIMESTAMP,
    updated_at timestamptz,
    deleted_at timestamptz
);

INSERT INTO event_transactions (
    id,
    invoice_number,

    transaction_status,
    transaction_status_information, 

    payment_method,
    payment_channel,
    payment_expired_at,

    paid_at, 

    total_price, 
    tax_percentage,
    total_tax,
    admin_fee_percentage,
    total_admin_fee,
    grand_total,

    full_name,
    email,
    phone_number,

    is_compliment,
    is_refunded,

    created_at
) VALUES (
    '657d4ffe-51f7-4d5c-b519-b97591ec92d4',
    '123456789123456789',
    
    'SUCCESS',
    '',

    'MANDIRI-VA',
    'PAYLABS',
    '2025-07-14 22:18:09.776 +0700',

    null,

    750000,
    0,
    0,
    0,
    0,
    750000,

    'Riski Kukuh Wiranata',
    'test@gmail.com',
    '1203808128',

    false,
    false,

    NOW()
), (
    'ee89a3d4-45bc-455e-bdae-6069fc6fe910',
    '123456789123456790',
    
    'PENDING',
    '',

    'MANDIRI-VA',
    'PAYLABS',
    '2025-07-14 22:18:09.776 +0700',

    null,

    750000,
    0,
    0,
    0,
    0,
    750000,

    'Sample Suparman',
    'suparman@app.com',
    '08123238880',

    false,
    false,

    NOW()
);

UPDATE event_ticket_categories SET public_stock = 78 WHERE id = 'fefb5c70-25ee-4326-94e4-3aa21a007299';