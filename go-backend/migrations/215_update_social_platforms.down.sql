-- Roll back only the settings introduced by this migration.
-- Removed social platforms are intentionally not recreated.

DELETE FROM settings
WHERE "group" = 'social'
  AND key IN ('x', 'reddit');
