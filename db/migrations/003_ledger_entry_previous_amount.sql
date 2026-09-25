ALTER TABLE ledger_entries
    ADD COLUMN previous_amount DECIMAL(10,2) NULL AFTER edited_at;
