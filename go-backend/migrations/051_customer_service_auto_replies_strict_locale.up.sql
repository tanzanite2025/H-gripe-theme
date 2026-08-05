-- 051: automatic replies are language-specific.
--
-- Legacy wildcard/empty/unknown locale rows are preserved for audit but
-- disabled so they cannot silently answer in the wrong language.

ALTER TABLE ticket_auto_replies
    ALTER COLUMN locale SET DEFAULT 'en';

UPDATE ticket_auto_replies
SET locale = 'zh_cn'
WHERE LOWER(REPLACE(BTRIM(COALESCE(locale, '')), '-', '_'))
    IN ('zh', 'zh_cn', 'zh_hans', 'zh_sg');

UPDATE ticket_auto_replies
SET locale = SPLIT_PART(
    LOWER(REPLACE(BTRIM(locale), '-', '_')),
    '_',
    1
)
WHERE SPLIT_PART(
    LOWER(REPLACE(BTRIM(locale), '-', '_')),
    '_',
    1
) IN (
    'en', 'fr', 'de', 'es', 'ja', 'ko', 'it', 'pt', 'ru', 'ar',
    'fi', 'da', 'th', 'sv', 'id', 'ms', 'be', 'tr', 'bn', 'fa',
    'nl', 'hi', 'ur', 'mr', 'pcm', 'fil', 'te', 'ha', 'ps', 'sw',
    'tl', 'ta', 'jv'
);

UPDATE ticket_auto_replies
SET is_active = FALSE
WHERE BTRIM(COALESCE(locale, '')) = ''
   OR BTRIM(locale) = '*'
   OR LOWER(locale) NOT IN (
        'en', 'fr', 'de', 'es', 'ja', 'ko', 'it', 'pt', 'ru', 'ar',
        'fi', 'da', 'th', 'sv', 'id', 'ms', 'be', 'tr', 'bn', 'fa',
        'nl', 'hi', 'ur', 'mr', 'pcm', 'fil', 'te', 'ha', 'ps', 'sw',
        'tl', 'ta', 'jv', 'zh_cn'
   );

CREATE INDEX IF NOT EXISTS idx_ticket_auto_replies_strict_locale
    ON ticket_auto_replies(type, locale, is_active, priority DESC, created_at DESC);
