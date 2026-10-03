-- Products must be detached from this template before rolling the migration
-- back; the foreign key intentionally prevents deleting a template in use.
DELETE FROM product_specification_templates
WHERE slug = 'spoke_repair_kit';
