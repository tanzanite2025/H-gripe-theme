DO $$
DECLARE
  target_table TEXT;
BEGIN
  FOREACH target_table IN ARRAY ARRAY[
    'gift_cards',
    'transactions',
    'stripe_disputes',
    'payment_reviews',
    'loyalty_program_configs',
    'warranty_service_records',
    'shipping_carrier_services'
  ]
  LOOP
    IF EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = 'public'
        AND table_name = target_table
        AND column_name = 'currency'
    ) THEN
      EXECUTE format('ALTER TABLE %I ALTER COLUMN currency DROP DEFAULT', target_table);
    END IF;
  END LOOP;
END $$;
