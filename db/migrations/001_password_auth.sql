ALTER TABLE shops
    ADD COLUMN owner_email VARCHAR(255) NULL AFTER owner_phone,
    ADD COLUMN password_hash VARCHAR(255) NULL AFTER owner_email,
    ADD UNIQUE KEY uq_shops_owner_email (owner_email);