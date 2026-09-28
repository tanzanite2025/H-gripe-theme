-- Give the Schwalbe selector its own database-owned FAQ page.
-- The products layout already renders PageFaqSlot; this route record makes
-- the automatic route lookup and the admin FAQ structure point at the same
-- page instead of falling back to the generic tire-guides page.

WITH supported_locales(locale) AS (
    VALUES
        ('en'), ('fr'), ('de'), ('es'), ('ja'), ('ko'), ('it'), ('pt'),
        ('ru'), ('ar'), ('nl'), ('tr'), ('id'), ('th'), ('sv'), ('da'),
        ('fi'), ('hi'), ('ms'), ('zh_cn')
)
INSERT INTO faq_pages (
    page_id,
    route_path,
    domain,
    locale,
    title,
    subtitle,
    sort_order,
    status,
    created_at,
    updated_at
)
SELECT
    'guides-schwalbe-tire-selector',
    '/guides/tireguides/schwalbe-tire-selector',
    'guides',
    supported_locales.locale,
    CASE
        WHEN supported_locales.locale = 'zh_cn' THEN 'Schwalbe 外胎选型 FAQ'
        ELSE 'Schwalbe Tire Selector FAQs'
    END,
    CASE
        WHEN supported_locales.locale = 'zh_cn' THEN '解释 Schwalbe 目录中的胎圈、结构和字段名称'
        ELSE 'Explanations for Schwalbe catalog bead, construction, and field names'
    END,
    215,
    'active',
    NOW(),
    NOW()
FROM supported_locales
ON CONFLICT (page_id, locale) DO UPDATE
SET route_path = EXCLUDED.route_path,
    domain = EXCLUDED.domain,
    title = EXCLUDED.title,
    subtitle = EXCLUDED.subtitle,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    updated_at = NOW(),
    deleted_at = NULL;

WITH seed_faqs(
    locale,
    question,
    answer,
    sort_order
) AS (
    VALUES
        (
            'en',
            'What does WIRED mean for a Schwalbe tire bead?',
            '<p><strong>WIRED</strong> is the catalog label for a rigid wire bead. It describes how the tire bead is constructed and held in shape. This label alone does not establish the material details of every model, tubeless readiness, or rim compatibility; check the exact product page for those facts.</p>',
            110
        ),
        (
            'en',
            'What does Folding mean for a Schwalbe tire bead?',
            '<p><strong>Folding</strong> is the catalog label for a flexible folding bead. It describes the bead construction and its ability to fold for storage or transport. The label alone does not establish the exact reinforcement material, tubeless readiness, or compatibility with every rim; check the exact product page.</p>',
            120
        ),
        (
            'en',
            'What is a tire bead?',
            '<p>The <strong>bead</strong> is the reinforced edge of a tire that seats against the rim bead seat and helps hold the tire in place. In this catalog, <strong>WIRED</strong> and <strong>Folding</strong> describe the bead construction. They should be read separately from the <code>seal</code> field and from any model-specific TLE, TLR, or tubeless statement.</p>',
            130
        ),
        (
            'zh_cn',
            'Schwalbe 外胎上的 WIRED 胎圈是什么意思？',
            '<p><strong>WIRED</strong> 是目录中表示刚性钢丝胎圈的标签，用于说明胎圈如何构成并保持形状。这个标签本身不能推出每个型号的具体材料、真空胎就绪状态或车圈兼容性；这些信息应以具体型号官方产品页为准。</p>',
            110
        ),
        (
            'zh_cn',
            'Schwalbe 外胎上的 Folding 胎圈是什么意思？',
            '<p><strong>Folding</strong> 是目录中表示柔性折叠胎圈的标签，用于说明胎圈结构可以折叠收纳。这个标签本身不能推出具体增强材料、真空胎就绪状态或适配所有车圈；这些信息应以具体型号官方产品页为准。</p>',
            120
        ),
        (
            'zh_cn',
            '外胎的胎圈具体是什么？',
            '<p><strong>胎圈（bead）</strong> 是外胎边缘的加强部分，用于坐入车圈胎圈座并帮助外胎保持在车圈上。本目录中的 <strong>WIRED</strong> 和 <strong>Folding</strong> 都是在说明胎圈结构，应与 <code>seal</code> 字段以及型号页面中的 TLE、TLR 或真空胎说明分开理解。</p>',
            130
        )
)
INSERT INTO faqs (
    page_id,
    locale,
    question,
    answer,
    answer_image_url,
    answer_image_alt,
    answer_image_width,
    answer_image_height,
    status,
    "order",
    created_at,
    updated_at
)
SELECT
    'guides-schwalbe-tire-selector',
    seed_faqs.locale,
    seed_faqs.question,
    seed_faqs.answer,
    '',
    '',
    0,
    0,
    'published',
    seed_faqs.sort_order,
    NOW(),
    NOW()
FROM seed_faqs
WHERE NOT EXISTS (
    SELECT 1
    FROM faqs existing
    WHERE existing.deleted_at IS NULL
      AND existing.page_id = 'guides-schwalbe-tire-selector'
      AND existing.locale = seed_faqs.locale
      AND existing.question = seed_faqs.question
);
