-- Publish route-owned FAQ content through the existing editable backend tables.
-- Preserve administrator page identities, metadata, visibility, drafts and deletions.

WITH supported_locales(locale) AS (
    VALUES
        ('en'), ('fr'), ('de'), ('es'), ('ja'), ('ko'), ('it'), ('pt'),
        ('ru'), ('ar'), ('nl'), ('tr'), ('id'), ('th'), ('sv'), ('da'),
        ('fi'), ('hi'), ('ms'), ('zh_cn')
)
INSERT INTO faq_pages (
    page_id, route_path, route_key, manifest_version, route_status, last_seen_at,
    domain, locale, title, subtitle, sort_order, status, created_at, updated_at
)
SELECT
    'company-about-hole-patterns', '/company/about/hole-patterns', 'company-about-hole-patterns',
    '2026-10-15', 'current', NOW(), 'company', supported_locales.locale,
    CASE WHEN supported_locales.locale = 'zh_cn'
         THEN '孔位与打孔角度常见问题' ELSE 'Rim Hole Patterns & Drilling Angle FAQs' END,
    CASE WHEN supported_locales.locale = 'zh_cn'
         THEN '了解孔数、2:1 / G3 编法、左右偏孔及定制所需资料'
         ELSE 'Questions about hole counts, 2:1 / G3 lacing, offset holes and custom drilling requirements' END,
    250, 'active', NOW(), NOW()
FROM supported_locales
WHERE NOT EXISTS (
    SELECT 1 FROM faq_pages existing
    WHERE existing.route_path = '/company/about/hole-patterns'
      AND existing.locale = supported_locales.locale
      AND existing.deleted_at IS NULL
)
ON CONFLICT (page_id, locale) DO NOTHING;

UPDATE faq_pages
SET title = CASE WHEN title IN ('', 'Hole Patterns FAQs')
                THEN CASE WHEN locale = 'zh_cn'
                          THEN '孔位与打孔角度常见问题' ELSE 'Rim Hole Patterns & Drilling Angle FAQs' END
                ELSE title END,
    subtitle = CASE WHEN COALESCE(subtitle, '') IN ('', 'Rim drilling layouts and spoke hole customization')
                    THEN CASE WHEN locale = 'zh_cn'
                              THEN '了解孔数、2:1 / G3 编法、左右偏孔及定制所需资料'
                              ELSE 'Questions about hole counts, 2:1 / G3 lacing, offset holes and custom drilling requirements' END
                    ELSE subtitle END,
    updated_at = NOW()
WHERE route_path = '/company/about/hole-patterns'
  AND (title IN ('', 'Hole Patterns FAQs')
       OR COALESCE(subtitle, '') IN ('', 'Rim drilling layouts and spoke hole customization'))
  AND deleted_at IS NULL;

-- Existing answers, including drafts and soft-deleted rows, prevent reseeding a locale.
WITH seed_faqs(locale, question, answer, sort_order) AS (
    VALUES
        ('en', 'Which spoke hole counts can be customized?', '<p>We support 16, 20, 24, 28, 32 and 36-hole configurations. Confirm the selected rim model, hub hole count, lacing pattern and nipple assembly before ordering; the drilling layout is agreed for the complete wheel specification.</p>', 10),
        ('zh_cn', '可以定制哪些辐条孔数？', '<p>支持 16、20、24、28、32 和 36 孔配置。下单前需结合轮圈型号、花鼓孔数、预期编法及辐条帽装配方案，确认整套轮组的孔位与打孔要求。</p>', 10),
        ('en', 'Are radial, two-cross and three-cross the only supported lacing patterns?', '<p>The three diagrams are common examples. We also support 2:1, G3 and other hole layouts and lacing arrangements. Use the spoke lacing topology simulator linked on this page to explore the layout, then provide the hub drawing, left/right spoke distribution and rim specification so we can confirm the drilling plan.</p>', 20),
        ('zh_cn', '只能做放射、两交叉和三交叉这三种编法吗？', '<p>页面中的三种示意图只是常见示例，我们也支持 2:1、G3 等孔位排列与编法方案。可通过页面已有的辐条编法拓扑模拟器查看布局，再提供花鼓图纸、左右辐条分配及轮圈规格，确认实际打孔方案。</p>', 20),
        ('en', 'Do you support left/right offset rim holes?', '<p>Yes. Rim holes can be offset to the left and right of the rim center line, with axial drilling angles matched to the corresponding hub flanges. Hole position and drilling direction are separate requirements; confirm both from the drawing. In the third diagram, L / R mark the offset holes, while WL / WR measure flange distances from the hub center plane.</p>', 30),
        ('zh_cn', '轮圈支持左右偏孔吗？', '<p>支持。孔位可在轮圈中心线两侧左右偏置，并按对应花鼓法兰位置匹配轴向打孔角度。偏孔描述孔位位置，孔角度描述孔轴方向，两项要求都需按图纸确认。第 3 张示意图中的 L / R 标注左右偏孔，WL / WR 表示两侧法兰到花鼓中心面的距离。</p>', 30),
        ('en', 'How are axial and longitudinal rim drilling angles determined?', '<p>The two axial angles are matched to the left/right flange offsets and rim geometry. The two longitudinal angles also account for flange PCD, hole count and lacing pattern. We use the hub drawing and wheel geometry to match the hole axes to the spoke directions rather than applying one drilling angle to every wheel.</p>', 40),
        ('zh_cn', '轴向与纵向打孔角度是怎样确定的？', '<p>两个轴向角度结合左右法兰偏距与轮圈几何匹配；两个纵向角度还需考虑法兰 PCD、孔数及辐条编法。我们依据花鼓图纸和轮组几何，使孔轴方向与对应辐条走向匹配，而不是对所有轮组套用同一个打孔角度。</p>', 40),
        ('en', 'Does angle matching guarantee a completely straight spoke and nipple joint?', '<p>Zero deflection is the design target, not a guarantee of perfect alignment in every finished assembly. Machining tolerances and the internal carbon rim surface can affect the result. Correct nipples, any specified washers and controlled spoke tension remain necessary parts of the wheel build.</p>', 50),
        ('zh_cn', '角度匹配后能保证辐条与辐条帽完全零折角吗？', '<p>零折角是设计目标，不代表成品在所有装配条件下都能绝对同轴。加工公差及碳圈内部表面会影响实际配合，还需使用合适的辐条帽、按设计选用垫片，并控制编圈张力。</p>', 50),
        ('en', 'What information should I provide for a custom drilling plan?', '<p>Provide the hub drawing and revision, including WL / WR, left/right flange PCD and spoke exit positions; the rim model, ERD, hole count and left/right offset requirements; and the intended lacing pattern, left/right spoke distribution, nipples and washer arrangement. We then confirm the hole positions, drilling angles and manufacturing requirements before machining.</p>', 60),
        ('zh_cn', '确认定制打孔方案前需要提供哪些资料？', '<p>请提供花鼓图纸及版本，包括 WL / WR、左右法兰 PCD 与出条位置；轮圈型号、ERD、孔数及左右偏孔要求；以及预期编法、左右辐条分配、辐条帽和垫片方案。加工前再共同确认孔位、打孔角度及制造要求。</p>', 60)
)
INSERT INTO faqs (
    page_id, locale, question, answer,
    answer_image_url, answer_image_alt, answer_image_width, answer_image_height,
    status, "order", created_at, updated_at
)
SELECT
    page.page_id, seed_faqs.locale, seed_faqs.question, seed_faqs.answer,
    '', '', 0, 0, 'published', seed_faqs.sort_order, NOW(), NOW()
FROM seed_faqs
JOIN faq_pages page ON page.route_path = '/company/about/hole-patterns'
    AND page.locale = seed_faqs.locale AND page.deleted_at IS NULL
WHERE NOT EXISTS (
    SELECT 1 FROM faqs existing
    WHERE existing.page_id = page.page_id
      AND existing.locale = seed_faqs.locale
);
