ALTER TABLE event_transactions ADD COLUMN event_id UUID REFERENCES events(id) ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE event_transactions ADD COLUMN event_ticket_category_id UUID REFERENCES event_ticket_categories(id) ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE event_transactions DROP COLUMN full_name;
ALTER TABLE event_transactions DROP COLUMN phone_number;

UPDATE event_transactions SET event_id = '77797e23-a2b7-40bd-b8b0-ef628568f815', event_ticket_category_id = 'fefb5c70-25ee-4326-94e4-3aa21a007299' WHERE id = '657d4ffe-51f7-4d5c-b519-b97591ec92d4';
UPDATE event_transactions SET event_id = '77797e23-a2b7-40bd-b8b0-ef628568f815', event_ticket_category_id = 'fefb5c70-25ee-4326-94e4-3aa21a007299' WHERE id = 'ee89a3d4-45bc-455e-bdae-6069fc6fe910';

ALTER TABLE event_transaction_items DROP COLUMN event_ticket_category_id;