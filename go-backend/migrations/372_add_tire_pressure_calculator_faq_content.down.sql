-- Remove only the exact FAQ answers seeded by migration 372. The FAQ page
-- shell belongs to route reconciliation and is intentionally preserved.

WITH seed_faqs(locale, question, answer) AS (
    VALUES
        ('en', 'Is this calculator a tire-pressure recommendation?', '<p>No. This is a force and contact-area demonstration. It uses a simplified flat-road model and does not replace the exact pressure limits from the tire, rim, or wheel manufacturer.</p>'),
        ('en', 'How does the selected tire pressure enter the calculation?', '<p>The local tire record supplies a minimum and maximum pressure range. The demonstration uses the midpoint as the shared front and rear input, then estimates the load-bearing area with the relationship <code>A = Fz / P</code>. This midpoint is a calculation input, not a riding recommendation.</p>'),
        ('en', 'What do rider weight, bike weight, speed, and lean change?', '<p>Rider and bike weight set the total load, while the selected riding position distributes that load between the wheels. At a fixed lean angle, speed changes the equivalent turn radius and lateral demand. The result panel shows those relationships as model outputs.</p>'),
        ('en', 'Does the model calculate tread, compound, casing, or road texture?', '<p>No. The road is fixed as a flat demonstration surface, the rubber is a generic baseline, and the tire-body factor is normalized to 1. Nominal tire width can affect the displayed contact shape, but the model does not claim to reproduce a specific tread, compound, casing, or road surface.</p>'),
        ('en', 'What does the displayed contact area mean?', '<p>It is an estimated static load-bearing area derived from wheel load, pressure, and the selected nominal width. It is useful for comparing inputs inside this model; it is not a measured footprint and does not prove a grip or compatibility limit.</p>'),
        ('en', 'What does the wet-pressure demonstration do?', '<p>It applies a fixed 1 mm water-film demonstration baseline, compares the midpoint pressure with the selected record minimum, and reports a friction-retention proxy plus the changed area. It does not establish a universal wet-pressure reduction such as minus 5 or minus 7 PSI, and it is not a safety setting.</p>'),
        ('en', 'Why is one pressure shown for both front and rear?', '<p>The current calculator intentionally uses one shared pressure input so the effect of load, speed, lean, and area can be compared without adding another coupling layer. Real front and rear pressures must still follow the exact tire, rim, wheel, and bicycle setup instructions.</p>'),
        ('zh_cn', '这个计算器是不是胎压建议？', '<p>不是。这是受力和接地面积演示，使用简化的平路模型，不能替代外胎、车圈或轮组制造商针对具体产品给出的胎压限制。</p>'),
        ('zh_cn', '选中的胎压如何进入计算？', '<p>本地外胎记录提供最低和最高胎压范围，演示取范围中点作为前后轮共同输入，再用 <code>A = Fz / P</code> 估算承压面积。这个中点只是计算输入，不是骑行建议。</p>'),
        ('zh_cn', '骑手重量、车辆重量、速度和倾角会改变什么？', '<p>骑手和车辆重量决定总载荷，骑行姿势会把载荷分配到前后轮。倾角固定时，速度会改变等效转弯半径和侧向需求，结果区会把这些关系作为模型数据显示出来。</p>'),
        ('zh_cn', '模型会计算胎纹、胶料、胎体或路面纹理吗？', '<p>不会。路面固定为平路演示条件，橡胶使用通用基准，胎体因子归一化为 1。名义胎宽可以影响显示的接地形状，但模型不声称复现某个具体胎纹、胶料、胎体或路面。</p>'),
        ('zh_cn', '页面显示的接地面积代表什么？', '<p>它是根据车轮载荷、胎压和选中的名义胎宽推导出的静态承压面积估算值，适合在本模型内比较输入变化。它不是实测接地印迹，也不能证明抓地力或兼容性上限。</p>'),
        ('zh_cn', '湿地压力演示具体做了什么？', '<p>它使用固定 1 mm 水膜作为演示基准，把中点胎压与所选记录的最低胎压对比，并显示摩擦保留代理值和面积变化。它不能推出统一的减压 5 或 7 PSI，也不是安全设定。</p>'),
        ('zh_cn', '为什么前后轮显示同一个胎压？', '<p>当前计算器有意使用一个共同胎压输入，让用户在不增加额外耦合的情况下比较载荷、速度、倾角和面积的变化。真实前后轮胎压仍应遵循具体外胎、车圈、轮组和自行车的说明。</p>')
)
DELETE FROM faqs
WHERE page_id = 'guides-tireguides-tire-pressure-calculator'
  AND EXISTS (
      SELECT 1
      FROM seed_faqs
      WHERE seed_faqs.locale = faqs.locale
        AND seed_faqs.question = faqs.question
        AND seed_faqs.answer = faqs.answer
  );
