-- Roll back only untouched seed answers; keep backend page shells and admin edits.

WITH seed_faqs(locale, question, answer, sort_order) AS (
    VALUES
        ('en', 'What is the difference between a UD coating and a raw/paintless carbon finish?', '<p>A UD coating gives a simple, uniform surface appearance. A paintless finish skips additional painting and retains the carbon surface. These are surface choices; they do not establish the underlying fabric, layup or rim performance. Confirm the logo method, placement and appearance after choosing a finish.</p>', 10),
        ('en', 'Can a 3K carbon fabric rim have a UD coating?', '<p>Yes. 3K describes the fabric pattern, while a UD coating describes the surface treatment; they can be combined. Availability of 3K, 12K, twill, marble and other appearances, and the underlying material and layup, depend on the selected rim specification and custom order.</p>', 20),
        ('en', 'How do matte, satin and glossy carbon rims differ?', '<p>Matte reflects less light, satin has a soft sheen, and glossy reflects more light. Gloss level and carbon pattern are separate appearance choices, and a paintless surface is not the same as a matte coating. Confirm the combination, treatment and additional production time before ordering.</p>', 30),
        ('en', 'What is the heat-resistance specification of the rim-brake track?', '<p>Our optimized high-temperature resin system gives the brake track as a whole a heat-resistance specification of 260°C. This applies to rim-brake tracks; it does not establish the heat resistance of other rim areas or performance under every braking condition.</p>', 40),
        ('en', 'What are the tradeoffs of a laser-treated brake track?', '<p>Laser treatment can increase surface friction and also affects the outer carbon fiber layer. These options have been tested, but a tradeoff remains. We recommend the standard brake track with its optimized surface first, subject to the selected rim model. Logo engraving and brake-track laser treatment serve different purposes.</p>', 50),
        ('zh_cn', 'UD 涂装与免涂装表面有什么区别？', '<p>UD 涂装提供简洁、统一的表面外观；免涂装省去额外涂装，保留碳纤维本身的表面质感。它们是表面方案，不能据此判断底材、铺层或轮圈性能。选择后再确认 Logo 工艺、位置和效果。</p>', 10),
        ('zh_cn', '3K 碳布底材可以采用 UD 涂装吗？', '<p>可以。3K 描述碳布纹理，UD 涂装描述表面方案，两者可以组合。3K、12K、斜纹、大理石纹等外观的可行性、底材与铺层，以所选轮圈规格及定制方案为准。</p>', 20),
        ('zh_cn', '哑光、缎面和亮光碳圈如何选择？', '<p>哑光反光较少，缎面具有柔和光泽，亮光反光更明显。光泽与碳布纹理是不同的外观维度，免涂装也不等于哑光涂装；具体组合、表面处理和增加的制作时间需在下单前确认。</p>', 30),
        ('zh_cn', '圈刹碳圈刹车边的耐温参数是多少？', '<p>我们采用经过优化的高温树脂，刹车边整体耐温参数为 260°C。此参数仅针对圈刹刹车边，不能据此推断碳圈其他部位的耐温或所有制动条件下的表现。</p>', 40),
        ('zh_cn', '刹车边做激光镭射处理有什么取舍？', '<p>镭射处理可增加刹车边表面摩擦力，也会影响外层碳纤维。相关方案经过测试，但仍存在取舍；我们优先推荐经过表面优化的常规刹车边，并按具体轮圈型号确认。Logo 镭射与刹车边镭射用途不同。</p>', 50)
)
DELETE FROM faqs existing
USING seed_faqs, faq_pages page
WHERE page.route_path = '/company/about/appearance'
  AND page.locale = seed_faqs.locale
  AND existing.page_id = page.page_id
  AND existing.locale = seed_faqs.locale
  AND existing.question = seed_faqs.question
  AND existing.answer = seed_faqs.answer
  AND existing."order" = seed_faqs.sort_order
  AND existing.status = 'published'
  AND existing.answer_image_url = ''
  AND existing.answer_image_alt = ''
  AND existing.answer_image_width = 0
  AND existing.answer_image_height = 0
  AND existing.updated_at = existing.created_at
  AND existing.deleted_at IS NULL;
