CREATE TABLE IF NOT EXISTS event_additional_fees (
    id SERIAL PRIMARY KEY,
    event_id uuid not null references events(id) REFERENCES events(id) ON DELETE CASCADE ON UPDATE CASCADE,
    name varchar(100) not null,
    is_percentage boolean default false,
    is_tax boolean default false, -- if true, this fee is tax and if false then it's and admin fee
    value float8 not null default 0,

    created_at timestamptz default NOW(),
    updated_at timestamptz default NOW()
);
