ALTER TABLE quick_buy_flows
    RENAME COLUMN selection_help_text TO help_text;

ALTER TABLE quick_buy_flow_translations
    RENAME COLUMN selection_help_text TO help_text;

-- Preserve legacy step instructions before removing the step-level fields.
WITH selected_versions AS (
    SELECT DISTINCT ON (flow_id)
        id,
        flow_id
    FROM quick_buy_flow_versions
    ORDER BY
        flow_id,
        CASE WHEN status = 'published' THEN 0 ELSE 1 END,
        version_number DESC,
        id DESC
),
step_content AS (
    SELECT
        selected_versions.flow_id,
        string_agg(
            format('%s: %s', step.name, details.text),
            E'\n\n'
            ORDER BY step.sort_order, step.id
        ) AS text
    FROM selected_versions
    JOIN quick_buy_steps AS step
        ON step.flow_version_id = selected_versions.id
    CROSS JOIN LATERAL (
        VALUES (
            NULLIF(
                BTRIM(
                    CONCAT_WS(
                        E'\n',
                        NULLIF(BTRIM(step.description), ''),
                        NULLIF(BTRIM(step.help_text), '')
                    )
                ),
                ''
            )
        )
    ) AS details(text)
    WHERE details.text IS NOT NULL
    GROUP BY selected_versions.flow_id
)
UPDATE quick_buy_flows AS flow
SET help_text = CASE
    WHEN BTRIM(flow.help_text) = '' THEN step_content.text
    ELSE flow.help_text || E'\n\n' || step_content.text
END
FROM step_content
WHERE flow.id = step_content.flow_id
  AND step_content.text IS NOT NULL
  AND BTRIM(step_content.text) <> '';

WITH selected_versions AS (
    SELECT DISTINCT ON (flow_id)
        id,
        flow_id
    FROM quick_buy_flow_versions
    ORDER BY
        flow_id,
        CASE WHEN status = 'published' THEN 0 ELSE 1 END,
        version_number DESC,
        id DESC
),
translated_step_content AS (
    SELECT
        selected_versions.flow_id,
        translation.locale,
        string_agg(
            format('%s: %s', step.name, details.text),
            E'\n\n'
            ORDER BY step.sort_order, step.id
        ) AS text
    FROM selected_versions
    JOIN quick_buy_steps AS step
        ON step.flow_version_id = selected_versions.id
    JOIN quick_buy_step_translations AS translation
        ON translation.step_id = step.id
    CROSS JOIN LATERAL (
        VALUES (
            NULLIF(
                BTRIM(
                    CONCAT_WS(
                        E'\n',
                        NULLIF(BTRIM(translation.description), ''),
                        NULLIF(BTRIM(translation.help_text), '')
                    )
                ),
                ''
            )
        )
    ) AS details(text)
    WHERE details.text IS NOT NULL
    GROUP BY selected_versions.flow_id, translation.locale
)
UPDATE quick_buy_flow_translations AS flow_translation
SET help_text = CASE
    WHEN BTRIM(flow_translation.help_text) = '' THEN translated_step_content.text
    ELSE flow_translation.help_text || E'\n\n' || translated_step_content.text
END
FROM translated_step_content
WHERE flow_translation.flow_id = translated_step_content.flow_id
  AND flow_translation.locale = translated_step_content.locale
  AND translated_step_content.text IS NOT NULL
  AND BTRIM(translated_step_content.text) <> '';

WITH selected_versions AS (
    SELECT DISTINCT ON (flow_id)
        id,
        flow_id
    FROM quick_buy_flow_versions
    ORDER BY
        flow_id,
        CASE WHEN status = 'published' THEN 0 ELSE 1 END,
        version_number DESC,
        id DESC
),
translated_step_content AS (
    SELECT
        selected_versions.flow_id,
        translation.locale,
        string_agg(
            format('%s: %s', step.name, details.text),
            E'\n\n'
            ORDER BY step.sort_order, step.id
        ) AS text
    FROM selected_versions
    JOIN quick_buy_steps AS step
        ON step.flow_version_id = selected_versions.id
    JOIN quick_buy_step_translations AS translation
        ON translation.step_id = step.id
    CROSS JOIN LATERAL (
        VALUES (
            NULLIF(
                BTRIM(
                    CONCAT_WS(
                        E'\n',
                        NULLIF(BTRIM(translation.description), ''),
                        NULLIF(BTRIM(translation.help_text), '')
                    )
                ),
                ''
            )
        )
    ) AS details(text)
    WHERE details.text IS NOT NULL
    GROUP BY selected_versions.flow_id, translation.locale
)
INSERT INTO quick_buy_flow_translations (
    flow_id,
    locale,
    help_text
)
SELECT
    translated_step_content.flow_id,
    translated_step_content.locale,
    translated_step_content.text
FROM translated_step_content
WHERE translated_step_content.text IS NOT NULL
  AND BTRIM(translated_step_content.text) <> ''
  AND NOT EXISTS (
      SELECT 1
      FROM quick_buy_flow_translations AS existing
      WHERE existing.flow_id = translated_step_content.flow_id
        AND existing.locale = translated_step_content.locale
  );

DROP TABLE IF EXISTS quick_buy_step_translations;

ALTER TABLE quick_buy_steps
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS help_text;
