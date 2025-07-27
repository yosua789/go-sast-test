CREATE TABLE IF NOT EXISTS event_order_information_books (
    id serial primary key,
    event_id uuid not null REFERENCES events(id) ON DELETE CASCADE ON UPDATE CASCADE,
    event_transaction_id uuid REFERENCES event_transactions(id) ON DELETE CASCADE ON UPDATE CASCADE,

    email varchar(255) not null,
    full_name varchar(255) not null,

    created_at timestamptz not null
);

ALTER TABLE event_order_information_books ADD CONSTRAINT unique_event_email_book UNIQUE (event_id, email);

INSERT INTO event_order_information_books
(
    event_id,
    event_transaction_id,
    email,
    full_name,
    created_at
) VALUES (
    '77797e23-a2b7-40bd-b8b0-ef628568f815',
    '657d4ffe-51f7-4d5c-b519-b97591ec92d4',
    'test@gmail.com',
    'Riski Kukuh Wiranata',
    NOW()
), (
    '77797e23-a2b7-40bd-b8b0-ef628568f815',
    'ee89a3d4-45bc-455e-bdae-6069fc6fe910',
    'suparman@app.com',
    'Sample Suparman',
    NOW()
);