ALTER TABLE event_transactions DROP COLUMN event_id;
ALTER TABLE event_transactions DROP COLUMN event_ticket_category_id;
ALTER TABLE event_transaction_items ADD COLUMN event_ticket_category_id uuid not null REFERENCES event_ticket_categories(id) ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE event_transactions ADD COLUMN full_name varchar(255) not null;
ALTER TABLE event_transactions DROP COLUMN phone_number varchar(255) not null;

UPDATE event_transaction_items SET full_name = 'test', event_ticket_category_id = 'fefb5c70-25ee-4326-94e4-3aa21a007299' WHERE transaction_id IN ('ee89a3d4-45bc-455e-bdae-6069fc6fe910', '657d4ffe-51f7-4d5c-b519-b97591ec92d4', '8cac9818-9735-4eb8-963a-a18187a85e28', 'fb7fe09b-46cb-40af-9bf7-21b2e761dff6', '902f8897-452b-4d6f-9312-c09e3b0a93ac');