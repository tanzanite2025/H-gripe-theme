WITH current_policy AS (
    SELECT
        id,
        value::jsonb AS payload
    FROM settings
    WHERE key = 'refund_cancellation_policy'
      AND locale = 'en'
      AND jsonb_typeof(value::jsonb->'sections') = 'array'
),
updated_policy AS (
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
                                to_jsonb('For stocked, non-custom items, we accept return requests within 30 days of delivery when the item is unused and in original packaging. Made-to-order and custom-configured items follow the cancellation rules in Special Orders.'::text)
                            )
                        WHEN 'special-orders' THEN
                            jsonb_set(
                                section,
                                '{body}',
                                to_jsonb('Non-stock or custom-configured products (special orders) are made to order. Cancellation may be requested before production or material cutting starts; after production begins, cancellation may be unavailable. If a post-production cancellation is approved, a 15%-20% custom handling fee may be deducted for committed materials and labor.'::text)
                            )
                        WHEN 'restocking-fee' THEN
                            jsonb_set(
                                jsonb_set(
                                    section,
                                    '{title}',
                                    to_jsonb('Restocking & Refurbishment'::text)
                                ),
                                '{body}',
                                to_jsonb('If an eligible return requires refurbishment or is missing original packaging or accessories, a fee of up to 20% of the original purchase value, with a minimum of USD $100, may be deducted after review.'::text)
                            )
                        ELSE section
                    END
                    ORDER BY ordinal
                )
                FROM jsonb_array_elements(payload->'sections') WITH ORDINALITY AS entries(section, ordinal)
            ),
            'updated_at',
            to_jsonb('2026-09-04T00:00:00Z'::text)
        ) AS payload
    FROM current_policy
)
UPDATE settings AS target
SET value = updated_policy.payload::text,
    updated_at = NOW()
FROM updated_policy
WHERE target.id = updated_policy.id;
