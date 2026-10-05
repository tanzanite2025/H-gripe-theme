-- Remove only unchanged FAQ rows seeded by migration 373. Keep the route-owned
-- page shell and any answers that an administrator has edited.

WITH seed_faqs(locale, question, answer) AS (
    VALUES
        ('en', 'Are the spoke lengths for 16-hole and 20-hole 0X patterns the same?', '<p>Hole count alone cannot determine the spoke length. The calculation needs the measured ERD, flange-hole PCD, flange spacing or offset, spoke-hole geometry, and the selected topology. ERD and PCD are diameters in mm; they are not radii.</p><p>With the same measured hub and rim dimensions, 16H and 20H 0X patterns use different hole-angle distributions, so the geometric result can differ. This page shows the hole mapping only. Use the spoke-length calculator with the actual measurements before ordering spokes.</p>'),
        ('en', 'How is a G3 2:1 pattern different from a uniform 2:1 pattern?', '<p>Both patterns distribute spokes 2:1 between the two sides, but their hole topology is different. The 21H G3 model is a 14/7 arrangement grouped into seven three-hole units. The uniform 24H 2:1 model is a 16/8 arrangement for a uniform-hole rim and follows its A-A-B hole sequence.</p><p>They are separate topology models, so their rim-hole mapping and spoke pairing cannot be exchanged. The labels describe the distribution and mapping shown here; they do not by themselves prove a stiffness, efficiency, tension, or build-safety result.</p>'),
        ('en', 'Are ERD and PCD entered as diameters or radii?', '<p>Enter both ERD and PCD as measured diameters in mm. ERD is the effective rim diameter at the nipple-seat reference, and PCD is the diameter of the flange-hole circle. Do not divide either input by two before entering it. A formula that explicitly uses a radius can perform that conversion internally, but the page and API measurement contract remains diameter in mm.</p>'),
        ('en', 'Do the displayed projection angles prove wheel stiffness or spoke tension?', '<p>No. The projection values describe the selected topology and its display geometry. They are not measured stiffness, efficiency, spoke tension, fatigue life, clearance, or assembly-safety conclusions. Those questions require the actual component dimensions, material properties, build process, and appropriate measurements or engineering validation.</p>'),
        ('zh_cn', '16 孔 0X 编法使用的辐条长度和 20 孔 0X 编法使用的辐条长度是一样的吗？', '<p>不能只看孔数判断辐条长度。计算还需要实测 ERD、花鼓法兰孔圈 PCD、法兰间距或偏置、辐条孔几何和选定拓扑。ERD 和 PCD 都按直径输入，单位是 mm，不是半径。</p><p>在花鼓和轮圈实测尺寸相同的假设下，16H 与 20H 的 0X 孔位角度分布不同，几何结果可能不同。本页只展示孔位映射，不计算辐条长度；下单前应把实际测量数据交给辐条长度计算器核算。</p>'),
        ('zh_cn', '2:1 G3 和普通均孔 2:1 有什么区别？', '<p>两者都是两侧按 2:1 分配辐条，但孔位拓扑不是同一种。21H G3 是 14/7 分配，按七组三孔单元组织；均孔 24H 2:1 是 16/8 分配，针对均匀孔距轮圈，并按 A-A-B 孔位序列处理。</p><p>因此两者的轮圈孔映射和辐条配对不能互换。这里的名称只说明页面展示的分配和拓扑，不直接代表刚度、效率、张力或整轮安全结论。</p>'),
        ('zh_cn', 'ERD 和 PCD 输入的是直径还是半径？', '<p>ERD 和 PCD 都应输入实测直径，单位是 mm。ERD 是条帽承座参考位置的有效轮圈直径，PCD 是法兰孔圈直径。输入前不要把数值除以二；如果某个公式内部明确使用半径，应由公式自行换算，但本页和 API 的测量口径仍是直径 mm。</p>'),
        ('zh_cn', '页面显示的投影角度能证明轮组刚度或辐条张力吗？', '<p>不能。投影值只描述当前拓扑及其显示几何关系，不是实测刚度、效率、辐条张力、疲劳寿命、间隙或装配安全结论。要回答这些问题，需要实际部件尺寸、材料参数、编轮工艺，以及相应的测量或工程验证。</p>')
)
DELETE FROM faqs existing
USING seed_faqs
WHERE existing.page_id = 'resources-wheelset-lacing-topology'
  AND existing.locale = seed_faqs.locale
  AND existing.question = seed_faqs.question
  AND existing.answer = seed_faqs.answer
  AND existing.deleted_at IS NULL;
