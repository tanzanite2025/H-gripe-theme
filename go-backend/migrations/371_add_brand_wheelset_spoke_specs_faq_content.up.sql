-- Seed the four repair questions that belong directly on the wheelset spoke
-- specification page. Answers stay database-owned so the FAQ editor can
-- update, translate, publish, and reorder them later.

WITH seed_faqs(locale, question, answer, sort_order) AS (
    VALUES
        (
            'en',
            'Why can the four spoke lengths on one disc-brake wheelset all be different?',
            '<p>Spoke length is calculated for each wheel and each side. The inputs include rim effective rim diameter (ERD), flange-hole circle diameter, flange-to-rim-center distance, spoke-hole position, and lacing pattern. When the two flange distances are different, the bracing angles and the final spoke lengths are different.</p><p>On a disc-brake front hub, the rotor, lockring, and bearing package often force a Rotor Dish flange offset, so the rotor side and non-rotor side have different bracing distances. On the rear hub, the freehub body, cassette, and drive-side bearing move the drive-side flange toward the centerline; the non-drive-side flange sits in another position. The drive-side spoke is often shorter, but the exact result depends on the hub and rim geometry.</p><p>For that reason, front-left, front-right, rear-left, and rear-right values can all differ. Confirm the drive/non-drive or rotor/non-rotor orientation, then recalculate from the exact ERD, hub model, hole count, and crossing pattern. A measured spoke length or a manufacturer service record is safer than copying a length from another side.</p>',
            110
        ),
        (
            'en',
            'What length tolerance is acceptable when replacing a straight-pull spoke?',
            '<p>For a straight-pull repair, treat ±1 mm as the target tolerance from the specified length. Straight-pull wheels often use a blind-hole or deep-hole nipple, so the available thread and bottoming clearance are limited. A spoke that is more than about 1 mm too long can bottom out; one that is too short can leave too little thread engagement to reach target tension.</p><p>A normal nipple often provides about 9–10 mm of usable thread. Workshop experience commonly keeps at least about 8 mm engaged, but this is a practical lower bound, not a universal guarantee: check the actual nipple depth, rim eyelet, washer, and spoke thread. Some factory records are not perfectly exact, so final acceptance should include full engagement, no bottoming, and measured spoke tension.</p><p>With a SAPIM POLYAX nipple and its rectangular self-locking thread, closer length matching gives more complete engagement and more reliable locking. Do not accept a replacement that is over 1 mm away simply because it can be screwed in; confirm it with the wheel manufacturer or an experienced wheel builder.</p>',
            120
        ),
        (
            'en',
            'Can a damaged DT Swiss Aerolite spoke be replaced directly with a Sapim CX-Ray?',
            '<p>Do not treat them as an automatic swap just because both are bladed spokes. The nominal sections are approximately 2.0 × 0.9 × 2.3 mm for DT Swiss Aerolite and 2.0 × 0.9 × 2.2 mm for Sapim CX-Ray. Their blade thickness and edge shape, forming, and tension elongation and fatigue behavior are not identical. The 0.1 mm profile difference can also change the airflow and drag; appearance alone cannot establish an aerodynamic match.</p><p>CX-Ray can be a service replacement when the hub head or slot, spoke length and thread, nipple, hole count, lacing pattern, and permitted tension all match, and the wheel is rebuilt, stress-relieved, and tension-checked. Keep the replacement matched within the same side or spoke group where possible.</p><p>Do not assume interchangeability for a dedicated straight-pull head, anti-twist feature, hidden nipple, or a wheel whose manufacturer specifies Aerolite only. Check the DT Swiss and Sapim data and the hub service instructions before ordering. Judge the repair by even tension and wheel runout, not just by whether the spoke fits through the hole.</p>',
            130
        ),
        (
            'en',
            'How can the hub serial number distinguish the Zipp 404 A1, B1, and Firecrest V2 generations?',
            '<p>Do not identify the generation from the words 404 Firecrest, a year, or a rim decal alone. A1 and B1 are SRAM/Zipp service-model suffixes. The serial number has no public, stable universal decoder, so record the complete hub serial number, rim label, and driver-body type, then verify them against the SRAM Service model page or a SRAM dealer. In the service tables, DS means drive side and NDS means non-drive side.</p><ul><li><strong>A1:</strong> The official <a href="https://www.sram.com/en/service/models/wh-404-ftld-a1" target="_blank">WH-404-FTLD-A1 service page</a> lists DS spoke lengths of 248/254 mm and NDS 252 mm, with entries that vary by driver-body configuration.</li><li><strong>B1:</strong> The official <a href="https://www.sram.com/en/service/models/wh-404-ftld-b1" target="_blank">WH-404-FTLD-B1 service page</a> lists DS 250/256 mm and NDS 252/254 mm.</li><li><strong>Firecrest V2:</strong> V2 is used as a generation or hub-family description, but there is no verified official <code>WH-404-FTLD-V2</code> service entry to support a length table. Do not invent a V2 length from the name.</li></ul><p>The current B1 record on this page is arranged by wheel and side as front 254/256 mm and rear 254/250 mm. SRAM service pages use DS/NDS and driver-body variants, so those fields are not interchangeable. If the serial, service model, driver body, or rim combination does not match, stop and confirm the official service data before ordering.</p>',
            140
        ),
        (
            'zh_cn',
            '为什么同一对碟刹轮组的前轮左右侧、后轮左右侧辐条长度会完全不同？',
            '<p>辐条长度不是只由轮圈框高决定，而是要对每个轮子、每个侧别分别计算。计算会用到轮圈有效直径（ERD）、花鼓法兰孔圈直径、法兰到轮圈中心面的距离、孔位和编法；只要左右法兰的支撑距离不同，力学夹角和最终长度就会不同。</p><p>前轮碟刹侧需要给碟片、锁环和轴承结构让位，花鼓法兰常出现 <code>Rotor Dish</code> 偏置，因此碟片侧和非碟片侧的支撑距离不一样。后轮则受到飞轮塔基、卡式飞轮和驱动侧轴承的限制，驱动侧法兰通常向轮组中线收拢，非驱动侧处在另一位置；驱动侧常见更短，但具体结果仍取决于花鼓和轮圈几何。</p><p>所以前左、前右、后左、后右四组长度都可能不同。维修时要先确认是碟片侧/非碟片侧还是驱动侧/非驱动侧，再按具体 ERD、花鼓型号、孔数和交叉数核算；实测辐条或厂商维修记录比照抄另一侧长度可靠。</p>',
            110
        ),
        (
            'zh_cn',
            '更换直拉辐条时，长度允许有多大的误差？',
            '<p>直拉辐条维修时，建议把相对规格长度的目标公差控制在 ±1 mm 内。直拉结构经常使用盲孔或深孔条帽，可用螺纹和到底余量更受限制；长出约 1 mm 可能在条帽内顶死，短了则可能有效啮合不足，无法拉到目标张力。</p><p>普通条帽的有效螺纹通常约 9–10 mm，维修经验一般要求至少保留约 8 mm 有效啮合，但这只是经验下限，不是所有结构都能套用的保证值。还要检查实际条帽深度、轮圈孔座、垫圈和辐条螺纹。部分官方记录的长度并不完全精准，因此最终应同时确认螺纹是否充分、是否顶底以及成品张力。</p><p>SAPIM POLYAX 这类带矩形螺纹自锁的条帽对长度更敏感；长度越匹配，完整啮合和自锁效果越可靠。不能因为辐条能拧进去就接受超过 1 mm 的偏差，应先让轮组厂商或有经验的轮组师确认。</p>',
            120
        ),
        (
            'zh_cn',
            'DT Swiss 轮组上的 Aerolite 辐条损坏，能否直接用 Sapim CX-Ray 替代？',
            '<p>不能因为两者都是扁辐条就默认直接替换。DT Swiss Aerolite 的标称截面约为 2.0 × 0.9 × 2.3 mm，Sapim CX-Ray 约为 2.0 × 0.9 × 2.2 mm；扁条厚度和边缘形状、成形工艺，以及受张后的延伸率和疲劳特性都不完全相同。0.1 mm 的截面差异也会影响迎风面和气动阻力，外观相似不等于气动表现相同。</p><p>如果花鼓辐条头或槽宽、辐条长度和螺纹、条帽、孔数、编法及允许张力全部匹配，并且重新编轮、消除应力和用张力计复核，CX-Ray 可以作为维修替换件。条件允许时，尽量在同一侧或同一受力组内保持型号一致。</p><p>如果是专用直拉头、带防转结构、轮圈内藏条帽，或厂商明确指定只能使用 Aerolite，就不能默认互换。下单前应核对 DT Swiss、Sapim 数据和花鼓维修手册；验收要看张力均匀度和轮圈跳摆，不只是看辐条能不能穿过孔位。</p>',
            130
        ),
        (
            'zh_cn',
            '如何通过花鼓序列号区分 Zipp 404 的 A1、B1 和 Firecrest V2 世代？',
            '<p>不能只看“404 Firecrest”字样、年份或轮圈贴花来判断世代。A1 和 B1 是 SRAM/Zipp 服务型号的后缀，序列号本身没有公开且稳定的通用解码表；维修时应记录完整花鼓序列号、轮圈标签和塔基类型，再到 SRAM Service 服务型号页或向 SRAM 经销商核对。服务表中的 DS 是驱动侧，NDS 是非驱动侧。</p><ul><li><strong>A1：</strong>官方 <a href="https://www.sram.com/en/service/models/wh-404-ftld-a1" target="_blank">WH-404-FTLD-A1 服务页</a>列出 DS 辐条长度 248/254 mm、NDS 252 mm，且数值会随塔基配置变化。</li><li><strong>B1：</strong>官方 <a href="https://www.sram.com/en/service/models/wh-404-ftld-b1" target="_blank">WH-404-FTLD-B1 服务页</a>列出 DS 250/256 mm、NDS 252/254 mm。</li><li><strong>Firecrest V2：</strong>“V2”更像产品或花鼓家族的代际称呼；目前没有可核验的 <code>WH-404-FTLD-V2</code> 官方服务条目，不能据名称推导一套 V2 出厂长度。</li></ul><p>本页当前的 B1 记录按轮子和侧别整理为前轮 254/256 mm、后轮 254/250 mm。SRAM 服务页使用 DS/NDS 并按塔基配置列值，字段维度不同，不能直接互相替换。只要序列号、服务型号、塔基或轮圈组合对不上，就应先确认官方维修资料再下单。</p>',
            140
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
    'resources-brand-wheelset-spoke-specs',
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
      AND existing.page_id = 'resources-brand-wheelset-spoke-specs'
      AND existing.locale = seed_faqs.locale
      AND existing.question = seed_faqs.question
);
