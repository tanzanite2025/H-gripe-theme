-- Both canonical columns belong to migration 001. Preserve reconciled content
-- when rolling the version back. Recreating obsolete duplicates would make an
-- old application's incompatible writers diverge again; do not restore them.
SELECT 1;
