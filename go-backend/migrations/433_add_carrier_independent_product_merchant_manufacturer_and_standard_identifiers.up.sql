ALTER TABLE products
    ADD COLUMN merchant_product_identifier VARCHAR(70) NOT NULL DEFAULT '',
    ADD COLUMN manufacturer_product_identifier VARCHAR(70) NOT NULL DEFAULT '',
    ADD COLUMN standard_product_identifier VARCHAR(70) NOT NULL DEFAULT '';

COMMENT ON COLUMN products.merchant_product_identifier IS 'Merchant product identifier (MPID), shared product data for carrier declarations';
COMMENT ON COLUMN products.manufacturer_product_identifier IS 'Actual manufacturer product identifier (NSPID), maintained independently of merchant SKU';
COMMENT ON COLUMN products.standard_product_identifier IS 'International standard product identifier (SPID), optional when unavailable';
