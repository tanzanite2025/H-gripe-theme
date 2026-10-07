-- A missing rate must never become a valid zero-tax rule through a database
-- default. Explicit zero remains valid when rate_decimal = 0 is supplied.
ALTER TABLE tax_rates
    ALTER COLUMN rate_decimal DROP DEFAULT;
