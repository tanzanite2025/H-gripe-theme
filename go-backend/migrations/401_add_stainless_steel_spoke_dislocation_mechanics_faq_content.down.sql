-- Remove only unchanged FAQ rows seeded by migration 401. Preserve the
-- route-owned FAQ page shell and any answers administrators have edited.

WITH seed_faqs(locale, question, answer) AS (
    VALUES
        ('en', 'What does the spoke elongation calculator estimate?', '<p>It estimates the axial length change of one spoke body from the selected AISI 304 material curve, effective load-bearing area, loaded length, and tension. The result separates total elongation, recoverable elastic elongation, and the residual elongation after ideal elastic unloading. It does not calculate the complete wheel''s tension loss.</p>'),
        ('en', 'Why does the selected spoke model change the result?', '<p>Each model supplies an effective load-bearing cross-sectional area. The calculator uses that area with force in σ=F/A, then reads strain from the same material curve. The Custom option lets you enter an area within the allowed input range; it does not measure a physical spoke or verify its manufacturer specification.</p>'),
        ('en', 'Is the 180 kgf reference a recommended wheel-building tension?', '<p>No. The table uses 180 kgf as a calculation allowance to compare spoke sections while leaving room for stress relief and service-load effects that the single-spoke model does not fully reproduce. It is not a target tension or universal safety limit. Follow the wheel and rim manufacturer''s specification and verify tension with suitable tools.</p>'),
        ('en', 'How do the material curve and cold-work hardening affect elongation?', '<p>The page uses a defined multi-pass, cold-drawn AISI 304 material state. Its engineering stress–strain curve already represents the rising post-yield response for that state. The calculator derives elastic strain from σ/E and reports the remaining curve strain as permanent elongation after ideal unloading. This is not a cyclic-fatigue model.</p>'),
        ('en', 'Does a result below the yield or rim reference prove that a wheel is safe?', '<p>No. The yield value is a reference from the selected material curve, and the spoke-hole force is a page-defined comparison baseline, not a rim certification or universal limit. The model omits whole-wheel geometry, local rim and hole-seat behavior, contact effects, and measured build tension. Use the exact component specifications and an appropriate wheel inspection.</p>'),
        ('en', 'What does the tensile-test failure elongation reference mean?', '<p>It converts the material reference''s total elongation-to-failure percentage over a tensile-test gauge length into a length reference for the entered spoke length. It is not a predicted fracture force, and it does not model necking or localized deformation after necking begins.</p>'),
        ('en', 'Why does the calculator stop when stress exceeds the curve endpoint?', '<p>The material curve does not support a result beyond its endpoint. The page returns no estimate there instead of extrapolating to fracture stress or force. Check the selected area and inputs, and treat an out-of-range load as a reason to verify the setup and component specifications.</p>'),
        ('zh_cn', '辐条伸长计算器估算的是什么？', '<p>计算器根据选定的 AISI 304 材料曲线、有效承力截面积、受力长度和张力，估算单根辐条杆身的轴向长度变化。结果分别列出总伸长、可恢复的弹性伸长，以及理想弹性卸载后的残余伸长。它不计算整组轮的总张力损失。</p>'),
        ('zh_cn', '为什么选择不同辐条型号会改变计算结果？', '<p>每个型号提供自己的有效承力截面积。计算器将该面积与力代入 σ=F/A，再从同一条材料曲线读取应变。“自定义”选项允许在规定输入范围内填写面积，但不会测量实物辐条，也不会核实制造商规格。</p>'),
        ('zh_cn', '180 kgf 参考值是推荐的编轮张力吗？', '<p>不是。表格用 180 kgf 作为计算余量基准，便于比较不同辐条截面，并为单根辐条模型无法完整还原的应力释放和使用载荷留出空间。它不是目标张力，也不是通用安全上限。请遵循轮组和轮圈制造商的规格，并使用合适工具测量张力。</p>'),
        ('zh_cn', '材料曲线和冷作硬化如何影响伸长结果？', '<p>本页使用定义好的多道次冷拔 AISI 304 材料状态。其工程应力—应变曲线已经包含该状态屈服后继续受拉时上升的材料响应。计算器用 σ/E 求弹性应变，并将曲线总应变的剩余部分作为理想卸载后的永久伸长。这不是循环疲劳模型。</p>'),
        ('zh_cn', '结果低于屈服或轮圈参考值，是否就能证明轮组安全？', '<p>不能。屈服值是所选材料曲线中的参考值，辐条孔载荷则是本页设定的比较基准；两者都不是轮圈认证结论或通用限值。模型没有覆盖整轮几何、轮圈及孔座的局部行为、接触效应和实测编轮张力。请结合具体零件规格和适当的轮组检查判断。</p>'),
        ('zh_cn', '拉伸试验的断裂伸长参考值是什么意思？', '<p>该参考值将材料资料中的拉伸试验总断裂伸长率按试验标距换算为输入辐条长度对应的长度参考。它不是预测的断裂力，也不模拟缩颈或缩颈后的局部变形。</p>'),
        ('zh_cn', '为什么应力超过曲线终点后计算器不再给出结果？', '<p>材料曲线不支持超出终点的结果，因此本页会停止估算，而不会外推断裂应力或断裂力。请检查所选截面积和输入；如果载荷超出范围，应重新核对计算设置和零件规格。</p>')
)
DELETE FROM faqs existing
USING seed_faqs
WHERE existing.page_id = 'guides-spokeguides-stainless-steel-microstructural-dislocation-mechanics'
  AND existing.locale = seed_faqs.locale
  AND existing.question = seed_faqs.question
  AND existing.answer = seed_faqs.answer
  AND existing.deleted_at IS NULL;
