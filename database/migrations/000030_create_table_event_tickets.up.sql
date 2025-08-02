CREATE TABLE IF NOT EXISTS event_tickets (
    id serial primary key,
    event_id uuid references events(id) on delete set null on update cascade,
    ticket_category_id uuid references event_ticket_categories(id) on delete set null on update cascade,
    event_transaction_id uuid references event_transactions(id) on delete set null on update cascade,

    ticket_owner_email varchar(255) not null,
    ticket_owner_full_name varchar(255) not null,
    ticket_owner_phone_number varchar(255),
    ticket_owner_garuda_id varchar(20),

    ticket_number varchar(255) not null,
    ticket_code varchar(255) not null,
    
    event_time timestamptz not null,
    event_venue varchar(255) not null,
    event_city varchar(255) not null,
    event_country varchar(255) not null,
    
    sector_name varchar(255) not null,
    area_code varchar(255),
    entrance varchar(255),

    seat_row integer,
    seat_column integer,
    seat_label varchar(255),

    is_compliment bool,

    additional_information text,
    
    created_at timestamptz not null,
    updated_at timestamptz
);

ALTER TABLE event_tickets ADD CONSTRAINT unique_event_tickets UNIQUE (event_id, ticket_category_id, ticket_number, ticket_code);