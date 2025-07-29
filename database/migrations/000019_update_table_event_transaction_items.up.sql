ALTER TABLE event_transaction_items ADD COLUMN garuda_id varchar(20);

ALTER TABLE event_transaction_items ADD COLUMN full_name varchar(255);
ALTER TABLE event_transaction_items ADD COLUMN email varchar(255);
ALTER TABLE event_transaction_items ADD COLUMN phone_number varchar(255);

UPDATE event_transaction_items SET garuda_id = 'IDA-ASDDS-012390' WHERE id = 1;
UPDATE event_transaction_items SET garuda_id = 'IDA-ASOKKK-312232' WHERE id = 2;