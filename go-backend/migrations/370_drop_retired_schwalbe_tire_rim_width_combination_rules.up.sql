-- Remove the table for environments that ran the superseded migration 362
-- before the selector contract was simplified.
DROP TABLE IF EXISTS schwalbe_tire_rim_width_combination_rules;
