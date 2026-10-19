ALTER TABLE product_specification_templates DROP CONSTRAINT retired_carbon_rim_template_slugs_cannot_be_recreated;
-- Physical deletion is intentional. Never recreate the removed templates or fields.
