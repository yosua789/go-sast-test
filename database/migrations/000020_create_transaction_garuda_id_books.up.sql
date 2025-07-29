CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS transaction_garuda_id_books (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE ON UPDATE CASCADE,
    garuda_id varchar(255) not null,
    created_at timestamptz not null default CURRENT_TIMESTAMP
);
