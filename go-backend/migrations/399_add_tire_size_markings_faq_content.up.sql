-- Seed database-owned FAQ content for the tire-size-markings guide.
-- The products layout renders the editable FAQ slot immediately before the
-- feedback slot. The seed is additive so administrators can revise, translate,
-- publish, or reorder the answers without Vue source changes.

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
    'guides-tireguides-tire-size-markings',
    '/guides/tireguides/tire-size-markings',
    'guides-tireguides-tire-size-markings',
    '2026-10-06',
    'current',
    NOW(),
    'guides',
    supported_locales.locale,
    CASE
        WHEN supported_locales.locale = 'zh_cn' THEN '轮胎尺寸标注 FAQ'
        ELSE 'Tire Size Marking FAQs'
    END,
    CASE
        WHEN supported_locales.locale = 'zh_cn' THEN '解读 ETRTO/ISO、英式和法式标注，并使用映射计算器核对对应关系'
        ELSE 'Questions about ETRTO/ISO, inch, and French markings and how to use the mapping calculator'
    END,
    165,
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
            'What does an ETRTO/ISO marking such as 37-622 mean?',
            '<p>The first number is the nominal tire width in millimetres and the second number is the bead seat diameter (BSD) in millimetres. <strong>37-622</strong> therefore means about 37 mm nominal width and a 622 mm BSD. The BSD is the clearest reference when checking the rim diameter.</p>',
            110
        ),
        (
            'en',
            'Why should I use ETRTO before an inch or French marking?',
            '<p>Inch and French names usually describe an approximate outside diameter and width, and historical naming systems can reuse similar names for different BSDs. ETRTO gives the width and bead seat diameter in millimetres, so it is the safer starting point for comparing the tire with the rim. Use the other markings to read packaging or a sidewall, then confirm the complete ETRTO value.</p>',
            120
        ),
        (
            'en',
            'What do 28 inch, 700C, and 29 inch mean when the BSD is 622 mm?',
            '<p>These are approximate or historical size names rather than exact rim measurements. A narrow 28 inch road tire and a wide 29 inch mountain-bike tire can both use a 622 mm BSD, while their mounted outside diameters and required frame clearance differ because their widths differ. Check the second ETRTO number first, then check width and clearance.</p>',
            130
        ),
        (
            'en',
            'What does the C in a French marking such as 700 x 35C mean?',
            '<p>In a French marking, <strong>700</strong> is an approximate outside-diameter family, <strong>35</strong> is an approximate width, and <strong>C</strong> is a historical rim-diameter code. The letter is useful for reading the name, but it is not a substitute for the exact BSD. Confirm the ETRTO marking before choosing a rim or replacement tire.</p>',
            140
        ),
        (
            'en',
            'Why can one inch or French name appear with more than one ETRTO size?',
            '<p>The older names are approximate and can be shared by different widths, bead seat diameters, or manufacturing conventions. The calculator keeps those rows separate and adds the ETRTO value when a displayed name is repeated. Select the complete marking that matches the tire sidewall instead of choosing by the shared name alone.</p>',
            150
        ),
        (
            'en',
            'What does the mapping calculator confirm, and what does it not confirm?',
            '<p>It shows common rows where ETRTO, inch, and French markings are recorded together, so selecting one value reveals the corresponding labels in that row. It does not certify rim compatibility, frame or fork clearance, pressure limits, or whether a particular tire is safe for a bicycle. Confirm the rim specification, the complete sidewall marking, and the manufacturer instructions before installation.</p>',
            160
        ),
        (
            'en',
            'What should I do when the tire sidewall markings do not match?',
            '<p>Do not average the names or rely on the largest-looking number. Recheck the complete ETRTO marking printed on the same tire, compare its BSD and width with the rim and bicycle clearance, and consult the manufacturer when the markings are damaged, incomplete, or contradictory. If the exact size cannot be verified, do not install the tire based only on an inch or French label.</p>',
            170
        ),
        (
            'zh_cn',
            '37-622 这样的 ETRTO/ISO 标注是什么意思？',
            '<p>第一个数字是外胎的名义宽度，单位是毫米；第二个数字是胎圈座直径（BSD），单位也是毫米。<strong>37-622</strong> 表示约 37 mm 名义宽度和 622 mm BSD。检查车圈直径时，BSD 是最清晰的依据。</p>',
            110
        ),
        (
            'zh_cn',
            '为什么要先看 ETRTO，而不是只看英式或法式标注？',
            '<p>英式和法式名称通常表达近似的外径和宽度，而且历史命名可能把相似名称用于不同 BSD。ETRTO 直接给出毫米制的宽度和胎圈座直径，因此更适合先用来核对车圈。其他标注可以帮助阅读包装或胎侧，但最终应确认完整的 ETRTO 数值。</p>',
            120
        ),
        (
            'zh_cn',
            '当 BSD 都是 622 mm 时，28 英寸、700C 和 29 英寸有什么关系？',
            '<p>这些是近似或历史尺寸名称，不是精确的车圈测量值。窄的 28 英寸公路外胎和较宽的 29 英寸山地外胎都可能使用 622 mm BSD，但安装后的外径和车架间隙会因宽度不同而变化。应先看 ETRTO 的第二个数字，再核对宽度和实际间隙。</p>',
            130
        ),
        (
            'zh_cn',
            '700 x 35C 中的 C 代表什么？',
            '<p>在法式标注中，<strong>700</strong> 是近似外径系列，<strong>35</strong> 是近似宽度，<strong>C</strong> 是历史上的车圈直径代码。这个字母有助于读懂名称，但不能替代精确 BSD。选择车圈或替换外胎前，仍要核对 ETRTO。</p>',
            140
        ),
        (
            'zh_cn',
            '为什么同一个英式或法式名称可能对应多个 ETRTO 尺寸？',
            '<p>这些旧式名称本身是近似值，可能被不同宽度、不同胎圈座直径或不同制造习惯重复使用。计算器会把这些行分开；当显示名称重复时，会附带 ETRTO 数值帮助区分。应选择与胎侧完整标注一致的行，不能只按相同名称判断。</p>',
            150
        ),
        (
            'zh_cn',
            '映射计算器能确认什么，不能确认什么？',
            '<p>计算器展示的是同时记录了 ETRTO、英式和法式标注的常见映射行；选择其中一项后，会显示同一行的另外两种标注。它不能认证车圈兼容性、车架或前叉间隙、胎压上限，也不能证明某条外胎适合某辆自行车。安装前仍要核对车圈规格、胎侧完整标注和制造商说明。</p>',
            160
        ),
        (
            'zh_cn',
            '外胎侧面的几种标注不一致时应该怎么办？',
            '<p>不要把几个名称取平均，也不要只按看起来最大的数字判断。应重新确认同一条外胎上印刷的完整 ETRTO 标注，再将 BSD 和宽度与车圈规格及自行车间隙进行比较；如果标注破损、不完整或互相矛盾，应咨询制造商。无法确认精确尺寸时，不要只凭英式或法式名称安装。</p>',
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
    'guides-tireguides-tire-size-markings',
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
      AND existing.page_id = 'guides-tireguides-tire-size-markings'
      AND existing.locale = seed_faqs.locale
      AND existing.question = seed_faqs.question
);
