-- FAQ answers use one sanitized text source plus one optional dedicated image.
-- The image is not stored inside answer HTML so Nuxt can keep a stable layout.

ALTER TABLE faqs ADD COLUMN IF NOT EXISTS answer_image_url VARCHAR(500) NOT NULL DEFAULT '';
ALTER TABLE faqs ADD COLUMN IF NOT EXISTS answer_image_alt VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE faqs ADD COLUMN IF NOT EXISTS answer_image_width INTEGER NOT NULL DEFAULT 0;
ALTER TABLE faqs ADD COLUMN IF NOT EXISTS answer_image_height INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_faqs_answer_image_url ON faqs (answer_image_url) WHERE answer_image_url <> '';
