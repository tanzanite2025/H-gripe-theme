UPDATE settings
SET value = jsonb_set(
    value::jsonb,
    '{sections}',
    (
        SELECT jsonb_agg(
            CASE
                WHEN section->>'id' = 'special-orders' THEN
                    jsonb_set(
                        section,
                        '{body}',
                        to_jsonb('Non-stock or custom-configured products (special orders) are not eligible for return or refund, unless the issue is caused by our error.'::text)
                    )
                WHEN section->>'id' = 'high-value-signature' THEN NULL
                ELSE section
            END
            ORDER BY ordinal
        )
        FROM jsonb_array_elements(value::jsonb->'sections') WITH ORDINALITY AS entries(section, ordinal)
        WHERE section->>'id' <> 'high-value-signature'
    )
)::text,
updated_at = NOW()
WHERE key = 'refund_cancellation_policy'
  AND locale = 'en';
