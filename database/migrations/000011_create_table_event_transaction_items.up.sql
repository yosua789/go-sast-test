CREATE TABLE IF NOT EXISTS event_transaction_items (
    id serial primary key,
    transaction_id uuid not null REFERENCES event_transactions(id) ON DELETE SET NULL ON UPDATE CASCADE,
    event_ticket_category_id uuid not null REFERENCES event_ticket_categories(id) ON DELETE SET NULL ON UPDATE CASCADE,
    quantity int not null,
    seat_row int,
    seat_column int,
    additional_information jsonb,
    total_price int,
    created_at timestamptz not null default CURRENT_TIMESTAMP
);

INSERT INTO event_transaction_items (
    transaction_id,
    event_ticket_category_id,
    quantity,
    seat_row,
    seat_column,
    additional_information,
    total_price,
    created_at
) VALUES (
    'ee89a3d4-45bc-455e-bdae-6069fc6fe910',
    'fefb5c70-25ee-4326-94e4-3aa21a007299',
    1,
    1,
    1,
    '{"garuda_id": "IDA-ASDDS-012390"}',
    750000,
    NOW()
), (
    '657d4ffe-51f7-4d5c-b519-b97591ec92d4',
    'fefb5c70-25ee-4326-94e4-3aa21a007299',
    1,
    2,
    2,
    '{"garuda_id": "IDA-ASOKKK-312232"}',
    750000,
    NOW()
);
