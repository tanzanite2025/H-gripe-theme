DO $$
BEGIN
  RAISE EXCEPTION 'Product specification template retirement is irreversible; restore from a pre-migration backup only after explicitly redesigning the product catalog';
END $$;
