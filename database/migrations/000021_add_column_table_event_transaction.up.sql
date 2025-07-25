

ALTER TABLE event_transactions ADD COLUMN virtual_account_number varchar(70);
ALTER TABLE event_transactions ADD COLUMN channel_transaction_id varchar(150); -- for lookup purposes



