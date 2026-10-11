-- Remove only untouched long-tail seed answers. Retain administrator edits,
-- drafts, images, reordered or touched entries and the existing FAQ page.
WITH seed_faqs(locale, question, answer, sort_order) AS (
    VALUES
        ('en', 'Can I install a 28 mm tire on a 25 mm hookless rim?', '<p>The Schwalbe 05/2024 ETRTO width chart shown on this page lists 28 mm tires for 16–23 mm inner rims, so 28 mm on 25 mm is outside that chart. Its 25 mm column starts at 29 mm tire width. This reference does not approve every tire or rim model: check the exact manufacturers’ hookless compatibility requirements before fitting.</p>', 80),
        ('zh_cn', '28C 外胎能装在 25 mm 内宽的无钩车圈上吗？', '<p>本页展示的 Schwalbe 05/2024 ETRTO 对照表将 28 mm 外胎对应到 16–23 mm 内宽，因此 28／25 mm 不在该表范围内；25 mm 内宽列从 29 mm 胎宽开始。这个参考表不能替每个型号作安装批准，装配前仍须核对具体外胎与车圈制造商的无钩兼容要求。</p>', 80),
        ('en', 'What inner rim width is best for 28C road tires in this calculator?', '<p>The current calculator selects 21 mm inner width for a 28C tire. With the displayed estimate of 3 mm per rim wall, that corresponds to about 27 mm outer width; the entered tire width stays 28 mm. The ISO normal width range is shown separately as 15–23 mm. The selected value is this page’s recommendation, rather than an ISO-defined universal optimum or approval for every product.</p>', 90),
        ('zh_cn', '28C 公路外胎在这个计算器中推荐多大的车圈内宽？', '<p>当前计算器为 28C 外胎选出 21 mm 内宽。按页面显示的每侧约 3 mm 圈壁估算，对应约 27 mm 外宽，输入的胎宽仍为 28 mm；ISO 正常宽度范围另列为 15–23 mm。单一选值是本页推荐，不是 ISO 规定的通用最佳尺寸，也不代表所有具体型号都获准配合。</p>', 90),
        ('en', 'How much does inner rim width affect actual tire width?', '<p>A change in rim inner width can change the measured width of the same tire, but there is no universal 0.4 mm increase for every 1 mm of inner width. Tire construction and the manufacturer’s measuring rim affect the result. This calculator keeps your selected tire width unchanged; it does not turn 28C or 35C into a larger width automatically. Use product measurements or measure the fitted tire when actual clearance or aerodynamic dimensions matter.</p>', 100),
        ('zh_cn', '车圈内宽会让同一条外胎的实测宽度改变多少？', '<p>更换车圈内宽可能改变同一条外胎的实测宽度，但不存在对所有外胎都成立的“内宽每增加 1 mm，胎宽固定增加 0.4 mm”。胎体结构和制造商标定尺寸所用的测量车圈都会影响结果。本计算器保留你选定的胎宽，不会自动把 28C 或 35C 改成更大的数值；判断实际间隙或气动尺寸时，应查产品实测资料或测量装配后的外胎。</p>', 100)
)
DELETE FROM faqs existing
USING seed_faqs
WHERE existing.page_id = 'guides-tireguides-tire-rim-width-compatibility'
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
