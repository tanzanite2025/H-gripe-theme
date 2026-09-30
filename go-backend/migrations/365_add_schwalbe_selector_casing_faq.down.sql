-- Remove only the exact casing FAQ rows seeded by migration 365.
-- Keep rows that an administrator edited during the lifetime of this seed.

WITH seed_faqs(locale, question, answer) AS (
    VALUES
        ('en',
         'What do the Schwalbe casing construction values mean?',
         '<p>The selector reads exact casing-construction tokens from the catalog <code>version_label</code>. They describe the construction or intended use direction of the casing; they are not a universal quality ranking and do not replace the exact model specification.</p><ul><li><strong>Super Race</strong>: race-oriented construction with a light and supple direction.</li><li><strong>Super Ground</strong>: a balanced cross-country and all-round off-road construction direction.</li><li><strong>Super Trail</strong>: a trail-oriented construction with more emphasis on support and durability.</li><li><strong>Super Downhill</strong>: a reinforced gravity/downhill construction for demanding impacts.</li><li><strong>TRAIL</strong> and <strong>TRAIL PRO</strong>: trail-family construction labels. <code>PRO</code> is part of the catalog name and is not a cross-family grade.</li><li><strong>GRAVITY</strong> and <strong>GRAVITY PRO</strong>: gravity-family construction labels. <code>PRO</code> is part of the catalog name and is not a universal ranking.</li></ul><p>These values are separate from puncture protection, EPI, compound, color, bead, seal, radial construction, and load capacity. A record can contain more than one token, such as <code>GRAVITY PRO, Radial</code>; keep the complete original combination and check the exact product page, size, rim, pressure, and use case.</p>'),
        ('zh_cn',
         'Schwalbe 目录中的胎体结构值分别是什么意思？',
         '<p>筛选器读取目录 <code>version_label</code> 中的精确胎体结构 token。它们用于说明胎体的结构方向或使用取向，不是统一的质量排名，也不能替代具体型号的官方规格。</p><ul><li><strong>Super Race</strong>：竞赛取向，偏轻量和柔顺。</li><li><strong>Super Ground</strong>：偏 XC 和全能越野的平衡型结构方向。</li><li><strong>Super Trail</strong>：Trail 取向，更强调越野支撑和耐用方向。</li><li><strong>Super Downhill</strong>：Gravity/下坡取向的加强结构，用于应对更高冲击。</li><li><strong>TRAIL</strong> 和 <strong>TRAIL PRO</strong>：Trail 系列的胎体结构标签；<code>PRO</code> 是目录名称的一部分，不能当成跨系列通用等级。</li><li><strong>GRAVITY</strong> 和 <strong>GRAVITY PRO</strong>：Gravity 系列的胎体结构标签；<code>PRO</code> 是目录名称的一部分，不能理解成统一的高低排名。</li></ul><p>这些值要和防刺等级、EPI、胶料、颜色、胎圈、密封结构、径向结构及承重能力分开理解。一个记录可能有多个 token，例如 <code>GRAVITY PRO, Radial</code>；应保留完整原始组合，并结合具体型号产品页、尺寸、车圈、胎压和使用场景判断。</p>')
)
DELETE FROM faqs existing
USING seed_faqs
WHERE existing.page_id = 'guides-schwalbe-tire-selector'
  AND existing.locale = seed_faqs.locale
  AND existing.question = seed_faqs.question
  AND existing.answer = seed_faqs.answer;
