ALTER TABLE spoke_hub_models
    DROP CONSTRAINT IF EXISTS fk_spoke_hub_models_fitment_hub_specification;

DROP INDEX IF EXISTS idx_spoke_hub_models_fitment_specification_id;

ALTER TABLE spoke_hub_models
    DROP COLUMN IF EXISTS fitment_hub_specification_id;
