DROP INDEX IF EXISTS idx_order_evidence_attachments_item;
DROP INDEX IF EXISTS idx_order_evidence_items_order;
DROP INDEX IF EXISTS idx_order_evidence_items_package;
DROP INDEX IF EXISTS idx_order_evidence_packages_order;
DROP INDEX IF EXISTS uq_order_evidence_item_line_scope;
DROP INDEX IF EXISTS uq_order_evidence_item_order_scope;

DROP TABLE IF EXISTS order_evidence_attachments;
DROP TABLE IF EXISTS order_evidence_items;
DROP TABLE IF EXISTS order_evidence_packages;
