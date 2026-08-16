-- Remove the independently sellable hub and spoke templates.
-- Foreign keys intentionally prevent rollback after products reference them.
DELETE FROM product_specification_templates
WHERE slug IN ('hub', 'spoke');
