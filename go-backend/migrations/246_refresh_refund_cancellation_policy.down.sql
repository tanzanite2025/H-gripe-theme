WITH current_policy AS (
    SELECT
        id,
        value::jsonb AS payload
    FROM settings
    WHERE key = 'refund_cancellation_policy'
      AND locale = 'en'
      AND jsonb_typeof(value::jsonb->'sections') = 'array'
),
restored_policy AS (
    SELECT
        id,
        payload || jsonb_build_object(
            'sections',
            (
                SELECT jsonb_agg(
                    CASE section->>'id'
                        WHEN 'eligibility' THEN
                            jsonb_set(
                                section,
                                '{body}',
                                to_jsonb('We accept returns within 30 days of delivery for unused items in original packaging. Custom or personalized items are non-refundable unless defective.'::text)
                            )
                        WHEN 'special-orders' THEN
                            jsonb_set(
                                section,
                                '{body}',
                                to_jsonb('Non-stock or custom-configured products (special orders) are made to order. Cancellation may be requested before production or material cutting starts; after production begins, cancellation may be unavailable and a custom handling fee may apply.'::text)
                            )
                        WHEN 'restocking-fee' THEN
                            jsonb_set(
                                jsonb_set(
                                    section,
                                    '{title}',
                                    to_jsonb('Restocking & Refurbishment Fee'::text)
                                ),
                                '{body}',
                                to_jsonb('A restocking and refurbishment fee may be charged at 20% of the original purchase value, with a minimum of USD $100.'::text)
                            )
                        ELSE section
                    END
                    ORDER BY ordinal
                )
                FROM jsonb_array_elements(payload->'sections') WITH ORDINALITY AS entries(section, ordinal)
            )
        ) AS payload
    FROM current_policy
)
UPDATE settings AS target
SET value = restored_policy.payload::text,
    updated_at = NOW()
FROM restored_policy
WHERE target.id = restored_policy.id;
