-- Remove only the exact rows seeded by migration 361. If Admin has edited an
-- answer or added another FAQ, leave the customized content and its page in
-- place instead of deleting live editorial data during a rollback.
DELETE FROM faqs
WHERE page_id = 'guides-schwalbe-tire-selector'
  AND (
    (locale = 'en'
      AND question = 'What does WIRED mean for a Schwalbe tire bead?'
      AND answer = '<p><strong>WIRED</strong> is the catalog label for a rigid wire bead. It describes how the tire bead is constructed and held in shape. This label alone does not establish the material details of every model, tubeless readiness, or rim compatibility; check the exact product page for those facts.</p>')
    OR (locale = 'en'
      AND question = 'What does Folding mean for a Schwalbe tire bead?'
      AND answer = '<p><strong>Folding</strong> is the catalog label for a flexible folding bead. It describes the bead construction and its ability to fold for storage or transport. The label alone does not establish the exact reinforcement material, tubeless readiness, or compatibility with every rim; check the exact product page.</p>')
    OR (locale = 'en'
      AND question = 'What is a tire bead?'
      AND answer = '<p>The <strong>bead</strong> is the reinforced edge of a tire that seats against the rim bead seat and helps hold the tire in place. In this catalog, <strong>WIRED</strong> and <strong>Folding</strong> describe the bead construction. They should be read separately from the <code>seal</code> field and from any model-specific TLE, TLR, or tubeless statement.</p>')
    OR (locale = 'zh_cn'
      AND question = 'Schwalbe 外胎上的 WIRED 胎圈是什么意思？'
      AND answer = '<p><strong>WIRED</strong> 是目录中表示刚性钢丝胎圈的标签，用于说明胎圈如何构成并保持形状。这个标签本身不能推出每个型号的具体材料、真空胎就绪状态或车圈兼容性；这些信息应以具体型号官方产品页为准。</p>')
    OR (locale = 'zh_cn'
      AND question = 'Schwalbe 外胎上的 Folding 胎圈是什么意思？'
      AND answer = '<p><strong>Folding</strong> 是目录中表示柔性折叠胎圈的标签，用于说明胎圈结构可以折叠收纳。这个标签本身不能推出具体增强材料、真空胎就绪状态或适配所有车圈；这些信息应以具体型号官方产品页为准。</p>')
    OR (locale = 'zh_cn'
      AND question = '外胎的胎圈具体是什么？'
      AND answer = '<p><strong>胎圈（bead）</strong> 是外胎边缘的加强部分，用于坐入车圈胎圈座并帮助外胎保持在车圈上。本目录中的 <strong>WIRED</strong> 和 <strong>Folding</strong> 都是在说明胎圈结构，应与 <code>seal</code> 字段以及型号页面中的 TLE、TLR 或真空胎说明分开理解。</p>')
  );

DELETE FROM faq_pages
WHERE page_id = 'guides-schwalbe-tire-selector'
  AND status = 'active'
  AND sort_order = 215
  AND (
    (locale = 'en'
      AND title = 'Schwalbe Tire Selector FAQs'
      AND subtitle = 'Explanations for Schwalbe catalog bead, construction, and field names')
    OR (locale = 'zh_cn'
      AND title = 'Schwalbe 外胎选型 FAQ'
      AND subtitle = '解释 Schwalbe 目录中的胎圈、结构和字段名称')
    OR locale NOT IN ('en', 'zh_cn')
  )
  AND NOT EXISTS (
    SELECT 1
    FROM faqs existing
    WHERE existing.page_id = faq_pages.page_id
      AND existing.locale = faq_pages.locale
      AND existing.deleted_at IS NULL
  );
