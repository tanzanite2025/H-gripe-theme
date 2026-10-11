-- Seed editable tire/rim width selection FAQs for the existing route-owned page.
-- Preserve administrator content, publication choices, and custom page metadata.

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
    'guides-tireguides-choose',
    '/guides/tireguides/choose',
    'guides-tireguides-choose',
    '2026-10-11', 'current', NOW(), 'guides', supported_locales.locale,
    CASE WHEN supported_locales.locale = 'zh_cn'
         THEN '外胎与车圈宽度选择 FAQ'
         ELSE 'Tire and Rim Width Selection FAQs' END,
    CASE WHEN supported_locales.locale = 'zh_cn'
         THEN '了解尺寸选择、标准正常配合范围、最佳宽度与模拟图示'
         ELSE 'Questions about size selection, standard normal fitment ranges, best widths, and simulation' END,
    250, 'active', NOW(), NOW()
FROM supported_locales
ON CONFLICT (page_id, locale) DO UPDATE
SET route_path = EXCLUDED.route_path,
    route_key = EXCLUDED.route_key,
    manifest_version = EXCLUDED.manifest_version,
    route_status = EXCLUDED.route_status,
    last_seen_at = EXCLUDED.last_seen_at,
    domain = EXCLUDED.domain,
    title = CASE
        WHEN faq_pages.title = 'Tire and rim width reference calculator FAQs'
        THEN EXCLUDED.title ELSE faq_pages.title END,
    subtitle = CASE
        WHEN faq_pages.subtitle IN (
            'Convert inch tire widths and resolve DT Swiss rim internal width references',
            'Convert common tire-width markings and calculate engineering best-fit rim inner-width options'
        ) THEN EXCLUDED.subtitle ELSE faq_pages.subtitle END,
    updated_at = NOW();

-- Seed only empty locales. Existing edited or renamed questions are never
-- overwritten or supplemented again if this migration is reapplied.
WITH seed_faqs(locale, question, answer, sort_order) AS (
    VALUES
        ('en', 'How do I choose a size when I know only the tire width or the rim inner width?', '<p>For a known tire, select its width and the rim system to see the best rim inner width and the standard normal fitment range. For a known rim, switch to choosing a tire by rim inner width, then select the inner width and rim system to see the best tire width and the standard normal fitment range. The application is identified automatically from the size.</p>', 110),
        ('zh_cn', '只知道外胎尺寸，或者只知道车圈内宽，应该怎么选？', '<p>已知外胎时，选择外胎宽度和车圈系统，查看对应的最佳车圈内宽与标准正常配合范围。已知车圈内宽时，切换到按内宽选外胎，选择内宽和车圈系统，查看对应的最佳胎宽与标准正常配合范围。用途会随尺寸自动判断，无需再手动选择。</p>', 110),
        ('en', 'What do 25C, 35C, and inch sizes mean in the width selector?', '<p>This page uses 25C as a 25 mm tire width and 35C as a 35 mm tire width, without increasing the selected tire width. The selector shows corresponding millimeter and inch values, and calculations use millimeters. This selector covers width only: the tire and rim bead-seat diameters must also match. Equal width does not make different diameters interchangeable.</p>', 120),
        ('zh_cn', '25C、35C 和英寸规格在选择器里是什么意思？', '<p>本页将 25C 按 25 mm 胎宽、35C 按 35 mm 胎宽计算，不再额外放大成另一种胎宽。选择器同时列出毫米和英寸对应值，选中后按毫米统一计算。这里只选择宽度；外胎与车圈的胎圈座直径还必须一致，不能因为宽度相同就把不同直径的规格互换。</p>', 120),
        ('en', 'How do rim inner and outer widths differ, and why is each wall estimated at about 3 mm?', '<p>Inner width is the distance between the inside faces of the rim; outer width is the distance between its outside faces. This page estimates each wall at about 3 mm, so estimated outer width = inner width + 6 mm. A 19 mm inner width therefore gives an estimated 25 mm outer width. This estimate explains narrow-tire aerodynamic matching; use the actual rim specification or measurement for its real wall thickness and outer width.</p>', 130),
        ('zh_cn', '车圈内宽和外宽有什么区别，为什么结果会提到每侧约 3 mm？', '<p>内宽是车圈内侧两边之间的宽度，外宽是车圈外侧两边之间的宽度。本页按每侧约 3 mm 估算圈壁，所以估算外宽 = 内宽 + 6 mm；例如 19 mm 内宽对应约 25 mm 外宽。这是用于解释窄胎气动匹配的估算，实际圈壁厚度和外宽应以具体车圈资料或测量为准，不能把估算值当作实测值。</p>', 130),
        ('en', 'Should I use the standard normal fitment range or the best width?', '<p>First check the normal fitment range and standards reference card to see whether the pairing falls within the listed ISO width range. Then check the best width for the single size recommended by this page. A standard range describes possible nominal size combinations, rather than equal support or shape at every size. Compare the best value with the normal range as well, and check the exact product requirements; a width table alone cannot guarantee that a tire will stay seated.</p>', 140),
        ('zh_cn', '标准正常配合范围和最佳宽度，应该看哪一个？', '<p>先看“正常配合范围与标准参考”卡片，确认尺寸组合是否在列出的 ISO 宽度范围内；再看“最佳宽度”，了解本页针对该尺寸给出的单一推荐。标准范围回答哪些名义尺寸可以配合，不表示范围内每个尺寸的支撑或轮廓效果都一样。最佳值也要与正常范围对照，具体型号仍需满足制造商要求，不能仅凭宽度表承诺不会脱胎。</p>', 140),
        ('en', 'How does the 105% aerodynamic matching work for tires up to 30C?', '<p>This page applies its 105% matching only to tires up to and including 30C. For a known rim, target tire width = estimated rim outer width × 1.05. For a known tire, target rim outer width = tire width ÷ 1.05; subtract both walls to obtain the target inner width. The result selects the nearest integer size and does not alter the input tire width. This aerodynamic selection method is displayed separately from the ISO normal fitment range.</p>', 150),
        ('zh_cn', '30C 及以下的 105% 气动匹配怎样计算？', '<p>本页的 105% 匹配仅用于 30C 及以下，包含 30C。已知车圈时，目标胎宽 = 估算车圈外宽 × 1.05；已知外胎时，目标车圈外宽 = 胎宽 ÷ 1.05，再扣除两侧圈壁得到目标内宽。结果选取最接近目标的整数规格，输入胎宽本身保持不变。这个气动选值方法与 ISO 正常配合范围分开显示。</p>', 150),
        ('en', 'How are tires wider than 30C recommended, and why is a best value sometimes unavailable?', '<p>Tires wider than 30C do not use the 1.05 factor. Their size automatically selects the gravel, mountain, Plus, or fat/snow matching path, which uses the available width fitment information for that application to select one recommendation. When that information cannot support a single recommendation, the normal fitment range remains visible and the best value is unavailable. An arbitrary number in the range is not presented as best, and an unavailable recommendation does not by itself mean installation is impossible.</p>', 160),
        ('zh_cn', '超过 30C 的宽胎怎么推荐，为什么有时没有最佳值？', '<p>超过 30C 后不再套用 1.05。页面根据尺寸自动进入砾石、山地、Plus 或雪地宽胎对应的匹配路径，结合该用途的宽度配合资料给出单一推荐。当前资料不足以支持单一推荐时，会保留正常配合范围并说明暂无最佳值，不把范围中任意一个数字当作最佳，也不把暂无推荐等同于无法安装。</p>', 160),
        ('en', 'How do I choose hooked or hookless, and does adding a tube change the rim system?', '<p>Select the actual rim structure: HOOKED has bead hooks, while HOOKLESS does not. The question-mark button beside the rim system opens images explaining the difference. Adding an inner tube does not change the rim structure or make an unapproved tire suitable for a hookless rim. In addition to checking widths, verify that the exact tire model is approved for that rim system.</p>', 170),
        ('zh_cn', '有钩和无钩应该怎样选择？装内胎能把无钩变成有钩吗？', '<p>按实际车圈结构选择：HOOKED 是有钩，HOOKLESS 是无钩。点击车圈系统旁的问号可以查看两种结构的图片说明。装入内胎不会改变车圈结构，也不会让未获准用于无钩车圈的外胎自动变得适用。宽度计算之外，还应核对该外胎型号是否获准用于对应车圈系统。</p>', 170),
        ('en', 'Why are the DT Swiss and Schwalbe tables separate from the calculation results?', '<p>These tabs display the width combinations published by each source for comparison with their scope and context. A broad range describes possible combinations, rather than an optimum at every size. The calculation tab obtains its normal range and single best value from the existing calculation results; it does not recalculate from either displayed reference table. Switching reference tabs does not change the recommendation.</p>', 180),
        ('zh_cn', 'DT Swiss 和 Schwalbe 参考表为什么与计算结果分开？', '<p>这两个页签展示各自发布的宽度组合资料，用来对照其来源和适用范围。宽范围描述可能的尺寸组合，不代表每个组合都是最佳。计算页的正常范围和单一最佳值来自现有计算结果，不通过 DT Swiss 或 Schwalbe 展示表重新计算，也不会因切换参考页签而改变推荐。</p>', 180),
        ('en', 'What is the pairing diagram in the simulation tab for?', '<p>The simulation shows the cross-section and fitment status for the rim inner width, tire width, and rim system you select. Try different pairings to inspect tire shape, lateral support, and mismatches. The diagram keeps your selected sizes instead of replacing them with a recommendation. Use the calculation tab for the single best width and standard normal range. The simulation is illustrative and does not replace checks for the exact products.</p>', 190),
        ('zh_cn', '模拟页的搭配图有什么用途？', '<p>模拟页展示你自己选择的车圈内宽、外胎宽度和车圈系统所形成的截面与配合状态。可以尝试不同搭配，观察胎形、侧向支撑和不匹配情况；图示始终保留你选定的尺寸，不会偷偷替换成推荐尺寸。要查单一最佳宽度与标准正常范围，请查看计算页；模拟图是示意，不能代替具体型号的安装核对。</p>', 190)
)
INSERT INTO faqs (
    page_id, locale, question, answer,
    answer_image_url, answer_image_alt, answer_image_width, answer_image_height,
    status, "order", created_at, updated_at
)
SELECT
    'guides-tireguides-choose', seed_faqs.locale, seed_faqs.question, seed_faqs.answer,
    '', '', 0, 0, 'published', seed_faqs.sort_order, NOW(), NOW()
FROM seed_faqs
WHERE NOT EXISTS (
    SELECT 1 FROM faqs existing
    WHERE existing.deleted_at IS NULL
      AND existing.page_id = 'guides-tireguides-choose'
      AND existing.locale = seed_faqs.locale
);
