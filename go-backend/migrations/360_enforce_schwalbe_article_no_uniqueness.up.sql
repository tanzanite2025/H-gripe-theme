-- Enforce the one-Article-No-to-one-sales-Product contract at the database
-- boundary. The application preflight remains useful for a clear validation
-- message, but this index closes the concurrent-write race.
DO $$
DECLARE
    article_definition_id BIGINT;
    duplicate_article_no TEXT;
BEGIN
    SELECT definition.id
    INTO article_definition_id
    FROM product_spec_definitions AS definition
    JOIN product_specification_templates AS template
      ON template.id = definition.product_specification_template_id
    WHERE template.slug = 'schwalbe_tire'
      AND definition.slug = 'article_no';

    IF article_definition_id IS NULL THEN
        RAISE EXCEPTION 'Schwalbe article_no specification definition is missing';
    END IF;

    SELECT lower(btrim(value))
    INTO duplicate_article_no
    FROM product_spec_values
    WHERE spec_definition_id = article_definition_id
      AND btrim(value) <> ''
    GROUP BY lower(btrim(value))
    HAVING COUNT(*) > 1
    ORDER BY lower(btrim(value))
    LIMIT 1;

    IF duplicate_article_no IS NOT NULL THEN
        RAISE EXCEPTION 'duplicate Schwalbe Article No. prevents uniqueness index: %', duplicate_article_no;
    END IF;

    EXECUTE format(
        'CREATE UNIQUE INDEX IF NOT EXISTS uq_schwalbe_product_article_no ON product_spec_values (lower(btrim(value))) WHERE spec_definition_id = %s AND btrim(value) <> ''''',
        article_definition_id
    );
END $$;
