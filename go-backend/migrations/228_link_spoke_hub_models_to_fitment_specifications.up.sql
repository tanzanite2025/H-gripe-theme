ALTER TABLE spoke_hub_models
    ADD COLUMN IF NOT EXISTS fitment_hub_specification_id BIGINT;

UPDATE spoke_hub_models AS hub_model
SET fitment_hub_specification_id = specification.id
FROM fitment_hub_specifications AS specification
WHERE hub_model.fitment_hub_specification_id IS NULL
  AND hub_model.deleted_at IS NULL
  AND specification.deleted_at IS NULL
  AND LOWER(BTRIM(hub_model.code)) = LOWER(BTRIM(specification.spec_code));

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_spoke_hub_models_fitment_hub_specification'
    ) THEN
        ALTER TABLE spoke_hub_models
            ADD CONSTRAINT fk_spoke_hub_models_fitment_hub_specification
            FOREIGN KEY (fitment_hub_specification_id)
            REFERENCES fitment_hub_specifications(id)
            ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_spoke_hub_models_fitment_specification_id
    ON spoke_hub_models(fitment_hub_specification_id);
