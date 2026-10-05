-- Seed database-owned FAQ content for the inner-tube selection guide.
-- The products layout renders PageFaqSlot after the guide, so the questions
-- remain editable from Admin FAQ instead of being embedded in Vue.

WITH supported_locales(locale) AS (
    VALUES
        ('en'), ('fr'), ('de'), ('es'), ('ja'), ('ko'), ('it'), ('pt'),
        ('ru'), ('ar'), ('nl'), ('tr'), ('id'), ('th'), ('sv'), ('da'),
        ('fi'), ('hi'), ('ms'), ('zh_cn')
)
INSERT INTO faq_pages (
    page_id,
    route_path,
    route_key,
    manifest_version,
    route_status,
    last_seen_at,
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
    'guides-tireguides-choose-inner-tube',
    '/guides/tireguides/choose-inner-tube',
    'guides-tireguides-choose-inner-tube',
    '2026-10-05',
    'current',
    NOW(),
    'guides',
    supported_locales.locale,
    CASE
        WHEN supported_locales.locale = 'zh_cn' THEN '内胎选择 FAQ'
        ELSE 'Inner Tube Selection FAQs'
    END,
    CASE
        WHEN supported_locales.locale = 'zh_cn' THEN '说明内胎尺寸、气嘴类型、材料和高框车圈法嘴长度的选择方法'
        ELSE 'Questions about inner-tube size, valve type, material, and valve length for deep rims'
    END,
    175,
    'active',
    NOW(),
    NOW()
FROM supported_locales
ON CONFLICT (page_id, locale) DO UPDATE
SET route_path = EXCLUDED.route_path,
    route_key = EXCLUDED.route_key,
    manifest_version = EXCLUDED.manifest_version,
    route_status = EXCLUDED.route_status,
    last_seen_at = EXCLUDED.last_seen_at,
    domain = EXCLUDED.domain,
    title = EXCLUDED.title,
    subtitle = EXCLUDED.subtitle,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    updated_at = NOW(),
    deleted_at = NULL;

WITH seed_faqs(locale, question, answer, sort_order) AS (
    VALUES
        (
            'en',
            'How do I choose the correct inner-tube size?',
            '<p>Start with the tire''s ETRTO/ISO marking and choose a tube whose printed range contains the complete tire size. The diameter must match exactly; the width range must cover the mounted tire. An inch label alone can be ambiguous, so use the sidewall marking and verify the rim/tire system before installation.</p>',
            110
        ),
        (
            'en',
            'Which valve type should I choose: AV, DV, or SV?',
            '<p>Match the valve to the rim hole and the pump you will use. AV (Schrader) and DV (Dunlop/Woods) need their matching hole and hardware; SV (Presta/French) is common on narrow and high-pressure rims. Never force a valve through a hole of another size, and confirm the rim maker''s drilling and washer requirements.</p>',
            120
        ),
        (
            'en',
            'How do I choose valve length for a deep rim?',
            '<p>Use nominal rim depth plus the passage and outer-lip allowance, then leave enough stem for reliable pump-head grip. The calculator models a 6.5 mm outer-lip offset, a configurable 10–30 mm grip depth, and a preferred 3 mm exposure margin. These are reference parameters; the exact rim hole and pump instructions take precedence.</p>',
            130
        ),
        (
            'en',
            'What should I do with a 100 mm rim?',
            '<p>A 100 mm rim is inside this page''s supported reference range. The automatic model selects an 80 mm native valve with a 40 mm extender for the tested defaults, but this is not a product recommendation. Confirm the exact rim section, valve-hole geometry, removable-core design, extender seal, and pump head before installation.</p>',
            140
        ),
        (
            'en',
            'Can I use an extender with any inner tube?',
            '<p>No. An RVC extender requires a removable valve core and the extender''s specified thread and seal. A fixed-core tube cannot normally move its core to the extender tip, so a generic sleeve is not equivalent. Verify the tube and extender instructions together.</p>',
            150
        ),
        (
            'en',
            'How should I choose butyl, latex, or TPU?',
            '<p>Butyl prioritizes air retention and easy maintenance; latex can reduce weight and change ride feel but needs more frequent pressure checks and careful installation; TPU is light and compact but follows model-specific repair and valve-root instructions. After the material choice, still verify size, valve length, rim compatibility, and pressure limits.</p>',
            160
        ),
        (
            'en',
            'What if the calculation changes when I include measurement error?',
            '<p>Use the worst status at the lower and upper rim-depth bounds. If the status changes across the interval, remeasure the rim and choose a longer or otherwise compatible assembly before installation. The calculator is a reference model, not certification; manufacturer instructions and the exact parts control.</p>',
            170
        ),
        (
            'zh_cn',
            '如何选择正确尺寸的内胎？',
            '<p>先看外胎侧面的 ETRTO/ISO 标记，再选择印刷尺寸范围完整覆盖该外胎的内胎。直径必须完全一致，宽度范围要覆盖实际安装的外胎。只看英寸标注可能产生歧义，安装前应以侧壁标记为准，并核对车圈与外胎系统。</p>',
            110
        ),
        (
            'zh_cn',
            'AV、DV 和 SV 气嘴应该怎么选？',
            '<p>气嘴要同时匹配车圈气门孔和你要使用的打气筒。AV（美式/汽车嘴）和 DV（Dunlop/Woods/英式嘴）需要对应的孔径与配件；SV（法嘴）常见于窄胎和高压轮组。不要把不匹配的气嘴硬塞进孔内，并确认车圈制造商对钻孔和垫圈的要求。</p>',
            120
        ),
        (
            'zh_cn',
            '高框车圈应该怎样选择法嘴长度？',
            '<p>应以名义框高加上穿透和外唇余量为基础，再为打气筒夹头留下可靠的夹持长度。本页计算器使用 6.5 mm 外唇偏置，可调 10–30 mm 夹持深度，并把额外 3 mm 作为优选外露余量。这些是参考模型参数，具体车圈气孔和打气筒说明优先。</p>',
            130
        ),
        (
            'zh_cn',
            '车圈框高达到 100 mm 时应该怎么办？',
            '<p>100 mm 框高在本页参考模型的支持范围内。在当前默认参数下，自动计算会选择 80 mm 原生法嘴加 40 mm 延长嘴，但这不是商品推荐。安装前仍需确认车圈截面、气门孔几何、可拆气门芯结构、延长嘴密封方式和打气筒夹头。</p>',
            140
        ),
        (
            'zh_cn',
            '所有内胎都可以接延长嘴吗？',
            '<p>不可以。RVC 延长嘴要求内胎具有可拆气门芯，并且螺纹和密封方式要与延长嘴匹配。一体气门芯通常不能把原气门芯移到延长嘴顶端，普通套筒也不等同于 RVC 延长嘴。应把内胎和延长嘴的说明放在一起核对。</p>',
            150
        ),
        (
            'zh_cn',
            '丁基、乳胶和 TPU 内胎应该怎么选？',
            '<p>丁基内胎优先考虑保气和维护方便；乳胶内胎可能带来更低重量和不同的滚动感，但需要更频繁检查胎压并仔细安装；TPU 内胎重量轻、备件体积小，但维修和气嘴根部要求要按具体型号说明执行。确定材料后，仍要复核尺寸、法嘴长度、车圈兼容性和胎压上限。</p>',
            160
        ),
        (
            'zh_cn',
            '加入测量误差后计算状态发生变化怎么办？',
            '<p>应按框高区间下界和上界中的最差状态判断。如果区间内状态发生变化，应重新测量车圈，并在安装前选择更长或其他兼容的组合。本计算器是选型参考模型，不是认证；制造商说明和实际零件要求优先。</p>',
            170
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
    'guides-tireguides-choose-inner-tube',
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
      AND existing.page_id = 'guides-tireguides-choose-inner-tube'
      AND existing.locale = seed_faqs.locale
      AND existing.question = seed_faqs.question
);
