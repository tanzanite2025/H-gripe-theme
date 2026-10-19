-- Product-wide shipping inheritance is intentionally retired and cannot be restored.
DO $$ BEGIN
  RAISE EXCEPTION 'SKU-only shipping ownership is irreversible; restore a reviewed database backup to revert';
END $$;
