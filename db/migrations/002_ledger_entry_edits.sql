ALTER TABLE ledger_entries
    ADD COLUMN edited_at DATETIME NULL AFTER created_at;