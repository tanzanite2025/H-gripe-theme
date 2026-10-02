-- Products must be detached from this category before rolling the migration
-- back; the product foreign key intentionally prevents deleting a category in use.
DELETE FROM product_categories
WHERE slug = 'spoke-repair-kits';
