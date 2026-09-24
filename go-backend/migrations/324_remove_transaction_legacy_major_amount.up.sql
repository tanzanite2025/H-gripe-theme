-- Payment transaction amounts are now persisted only as exact minor units.
-- The major-unit float column is removed before launch to prevent accidental
-- reintroduction of a second monetary representation.
ALTER TABLE transactions
    DROP COLUMN IF EXISTS amount;
