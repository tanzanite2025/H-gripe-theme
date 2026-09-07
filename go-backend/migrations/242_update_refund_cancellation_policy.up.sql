WITH legacy AS (
    SELECT
        id,
        value::jsonb AS payload
    FROM settings
    WHERE key = 'refund_cancellation_policy'
      AND locale = 'en'
      AND value::jsonb->'sections' @> '[{"id":"special-orders","body":"Non-stock or custom-configured products (special orders) are not eligible for return or refund, unless the issue is caused by our error."}]'::jsonb
),
updated AS (
    SELECT
        id,
        jsonb_set(
            payload,
            '{sections}',
            (
                SELECT jsonb_agg(
                    CASE
                        WHEN section->>'id' = 'special-orders' THEN
                            jsonb_set(
                                section,
                                '{body}',
                                to_jsonb('Non-stock or custom-configured products (special orders) are made to order. Cancellation may be requested before production or material cutting starts; after production begins, cancellation may be unavailable and a custom handling fee may apply.'::text)
                            )
                        ELSE section
                    END
                    ORDER BY ordinal
                ) || jsonb_build_array(
                    jsonb_build_object(
                        'id', 'high-value-signature',
                        'title', 'High-Value Delivery',
                        'body', 'Orders totaling $750 USD or more require a direct signature at delivery. The signature requirement is part of our delivery and dispute-protection process.'
                    )
                )
                FROM jsonb_array_elements(payload->'sections') WITH ORDINALITY AS entries(section, ordinal)
            )
        ) || jsonb_build_object('updated_at', '2026-09-04T00:00:00Z') AS payload
    FROM legacy
)
UPDATE settings AS target
SET value = updated.payload::text,
    updated_at = NOW()
FROM updated
WHERE target.id = updated.id;
