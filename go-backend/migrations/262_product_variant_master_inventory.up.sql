ALTER TABLE product_variants
    ADD COLUMN IF NOT EXISTS master_variant_id BIGINT NULL;

-- Reconcile translations created before this migration. Option combinations
-- are stable across localized copies, so the root product variant is the
-- physical inventory owner for each matching translated row.
UPDATE product_variants translated
SET master_variant_id = master.id
FROM products translated_product
JOIN products root_product ON root_product.id = translated_product.parent_id
JOIN product_variants master
  ON master.product_id = root_product.id
WHERE translated.product_id = translated_product.id
  AND translated_product.parent_id IS NOT NULL
  AND master.option_values = translated.option_values
  AND translated.id <> master.id
  AND translated.master_variant_id IS NULL;

UPDATE product_variants
SET stock = 0
WHERE master_variant_id IS NOT NULL;

ALTER TABLE product_variants
    ADD CONSTRAINT fk_product_variants_master_variant
    FOREIGN KEY (master_variant_id) REFERENCES product_variants(id) ON DELETE RESTRICT;

ALTER TABLE product_variants
    ADD CONSTRAINT ck_product_variants_master_stock_zero
    CHECK (master_variant_id IS NULL OR stock = 0);

CREATE INDEX IF NOT EXISTS idx_product_variants_master_variant_id
    ON product_variants(master_variant_id);

-- Rows that cannot be matched remain physical inventory owners; future
-- translated copies populate master_variant_id directly.
