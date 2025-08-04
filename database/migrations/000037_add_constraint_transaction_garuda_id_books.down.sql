ALTER TABLE public.event_transaction_garuda_id_books
ADD CONSTRAINT event_garuda_unique UNIQUE (event_id, garuda_id);
