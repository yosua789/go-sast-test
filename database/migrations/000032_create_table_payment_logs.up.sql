CREATE TABLE IF NOT EXISTS payment_logs (
    id serial primary key,
    header text not null,
    body text not null,
    endpoint_path text not null, 
    response text not null,
    created_at timestamp with time zone default now(),
    error_response text,    
    error_code text,
);
