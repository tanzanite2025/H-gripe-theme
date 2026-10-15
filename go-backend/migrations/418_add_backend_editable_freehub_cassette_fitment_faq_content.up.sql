-- Add backend-editable FAQ content for the complete freehub/cassette tool route.
-- Preserve custom page identities, editorial metadata, visibility and existing answers.

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
    'guides-wheelset-buyers-choose-freehub',
    '/tools/freehub-cassette-fitment-engine', 'guides-wheelset-buyers-choose-freehub',
    '2026-10-14', 'current', NOW(), 'tools', supported_locales.locale,
    CASE WHEN supported_locales.locale = 'zh_cn'
         THEN '塔基与飞轮兼容性常见问题' ELSE 'Freehub and Cassette Compatibility FAQs' END,
    CASE WHEN supported_locales.locale = 'zh_cn'
         THEN '了解 HG、Micro Spline、XD、XDR、N3W 及垫圈与转接件要求'
         ELSE 'Questions about HG, Micro Spline, XD, XDR, N3W, spacers and adapters' END,
    250, 'active', NOW(), NOW()
FROM supported_locales
WHERE NOT EXISTS (
    SELECT 1 FROM faq_pages existing
    WHERE existing.route_path = '/tools/freehub-cassette-fitment-engine'
      AND existing.locale = supported_locales.locale
      AND existing.deleted_at IS NULL
)
ON CONFLICT (page_id, locale) DO NOTHING;

-- Localize generated route-catalog defaults without replacing administrator text.
UPDATE faq_pages
SET title = CASE WHEN title IN ('', 'Bicycle freehub body compatibility guide FAQs')
                THEN CASE WHEN locale = 'zh_cn'
                          THEN '塔基与飞轮兼容性常见问题' ELSE 'Freehub and Cassette Compatibility FAQs' END
                ELSE title END,
    subtitle = CASE WHEN COALESCE(subtitle, '') IN (
                    '', 'Match cassette standards, freehub bodies, spacers, and mechanical fitment rules'
                ) THEN CASE WHEN locale = 'zh_cn'
                            THEN '了解 HG、Micro Spline、XD、XDR、N3W 及垫圈与转接件要求'
                            ELSE 'Questions about HG, Micro Spline, XD, XDR, N3W, spacers and adapters' END
                ELSE subtitle END,
    updated_at = NOW()
WHERE route_path = '/tools/freehub-cassette-fitment-engine'
  AND (title IN ('', 'Bicycle freehub body compatibility guide FAQs')
       OR COALESCE(subtitle, '') IN (
           '', 'Match cassette standards, freehub bodies, spacers, and mechanical fitment rules'
       ))
  AND deleted_at IS NULL;

-- Seed only empty English/Chinese locales, including any custom route-owned page ID.
-- Drafts and soft-deleted answers also count as editorial content and prevent reseeding.
WITH seed_faqs(locale, question, answer, sort_order) AS (
    VALUES
        ('en', 'How do I identify the correct freehub body for my cassette?', '<p>Check the cassette brand, exact model, number of sprockets and tooth range, especially the smallest sprocket. Match those details to the compatibility engine, then read the freehub standard and spacer or adapter requirement. The rear derailleur brand or speed count alone does not identify the cassette interface. If the marking is unclear, photograph the cassette, lockring and back of the largest sprocket before choosing a wheelset freehub.</p>', 10),
        ('zh_cn', '怎样确认我的飞轮需要哪种塔基？', '<p>先确认飞轮品牌、具体型号、速数和齿比范围，尤其是最小齿片。将这些信息与兼容性引擎中的飞轮规格匹配，再查看塔基标准及垫圈或转接件要求。仅凭后拨品牌或速数无法判断飞轮接口；如果刻字不清楚，可拍摄飞轮、锁环和最大齿片背面的照片，再确认轮组塔基配置。</p>', 10),
        ('en', 'Do Shimano 12-speed road and mountain bike cassettes use the same freehub?', '<p>No. Shimano 12-speed road cassettes such as the 105 R7100, Ultegra R8100 and Dura-Ace R9200 11-30T or 11-34T options fit a standard HG-11 road freehub or a compatible HG L2 freehub without a spacer. Shimano 12-speed mountain bike 10-45T and 10-51T cassettes use Micro Spline. Micro Spline and road HG are different interfaces; confirm the actual cassette model rather than choosing by the Shimano name or 12-speed label alone.</p>', 20),
        ('zh_cn', 'Shimano 12 速公路与山地飞轮使用同一种塔基吗？', '<p>不一样。105 R7100、Ultegra R8100、Dura-Ace R9200 等 Shimano 12 速公路 11-30T 或 11-34T 飞轮可安装在标准 HG-11 公路塔基或兼容的 HG L2 塔基上，无需垫圈。Shimano 12 速山地 10-45T、10-51T 飞轮使用 Micro Spline。Micro Spline 与公路 HG 接口不同，不能只凭 Shimano 品牌或 12 速标签选择塔基。</p>', 20),
        ('en', 'What is the difference between SRAM XD and XDR, and can an XD cassette fit XDR?', '<p>XDR is 1.85 mm longer than XD. An XD cassette can fit an XDR freehub when a 1.85 mm spacer is installed at the inner base behind the cassette. A native XDR road cassette fits XDR without that spacer and does not fit the shorter XD body. Check the cassette interface first: the SRAM brand alone does not mean that every cassette uses XD or XDR.</p>', 30),
        ('zh_cn', 'SRAM XD 与 XDR 有什么区别，XD 飞轮能装在 XDR 上吗？', '<p>XDR 比 XD 长 1.85 mm。XD 飞轮可安装在 XDR 塔基上，但必须在飞轮后方、塔基最内侧底座处加装 1.85 mm 垫圈。原生 XDR 公路飞轮安装在 XDR 上无需这片垫圈，也不能安装在较短的 XD 塔基上。应先确认飞轮接口，不能认为所有 SRAM 飞轮都使用 XD 或 XDR。</p>', 30),
        ('en', 'Which freehub does a SRAM NX or SX Eagle 12-speed 11-50T cassette need?', '<p>The SRAM NX or SX Eagle 12-speed 11-50T cassette uses the traditional HG interface. It fits a mountain-width HG freehub directly; on the longer HG-11 road body it needs a 1.85 mm spacer behind the cassette. The SRAM Eagle 10-50T or 10-52T XD cassette follows a different rule. Confirm the cassette model and smallest sprocket before selecting the freehub.</p>', 40),
        ('zh_cn', 'SRAM NX 或 SX Eagle 12 速 11-50T 飞轮需要什么塔基？', '<p>SRAM NX 或 SX Eagle 12 速 11-50T 飞轮采用传统 HG 接口，可直接安装在山地宽度的 HG 塔基上；安装在较长的 HG-11 公路塔基上时，需要在飞轮后方加装 1.85 mm 垫圈。SRAM Eagle 10-50T 或 10-52T 的 XD 飞轮适用另一套规则，应先核对具体型号与最小齿片。</p>', 40),
        ('en', 'When is a 1.85 mm cassette spacer required on an HG-11 road freehub?', '<p>A mountain-width HG cassette needs a 1.85 mm spacer when fitted to the longer HG-11 road freehub. Examples in the engine include Shimano 11-speed mountain bike cassettes and SRAM NX or SX Eagle 11-50T cassettes. Standard full-width 11-speed road HG cassettes and the listed Shimano 12-speed road cassettes fit HG-11 without this conversion spacer. Some cassette models have additional requirements, so check the exact model and install the specified spacer behind the cassette.</p>', 50),
        ('zh_cn', '哪些飞轮安装在 HG-11 公路塔基上需要 1.85 mm 垫圈？', '<p>山地宽度的 HG 飞轮安装在较长的 HG-11 公路塔基上，需要加装 1.85 mm 垫圈。引擎中的示例包括 Shimano 11 速山地飞轮和 SRAM NX 或 SX Eagle 11-50T 飞轮。标准全宽 11 速公路 HG 飞轮及已列出的 Shimano 12 速公路飞轮安装在 HG-11 上，无需这片转换垫圈。部分型号还有额外要求，应按具体飞轮型号核对，并将规定的垫圈安装在飞轮后方。</p>', 50),
        ('en', 'Do all Shimano 10-speed road cassettes need the same spacer stack?', '<p>No. On an HG-11 road freehub, Shimano CS-5700 and CS-6700 cassettes need the 1.85 mm freehub conversion spacer plus their original 1.0 mm cassette spacer, for a total of 2.85 mm. The flat-backed Tiagra CS-4600 and CS-HG500-10 options listed in the engine need only the 1.85 mm conversion spacer on HG-11. Spacer thickness depends on both the freehub and cassette model; do not apply one 10-speed rule to every cassette.</p>', 60),
        ('zh_cn', 'Shimano 10 速公路飞轮的垫圈厚度都一样吗？', '<p>不一样。在 HG-11 公路塔基上，Shimano CS-5700、CS-6700 飞轮需要 1.85 mm 塔基转换垫圈，再加飞轮原配的 1.0 mm 垫圈，总厚度为 2.85 mm。引擎中列出的平底 Tiagra CS-4600、CS-HG500-10 飞轮安装在 HG-11 上，只需 1.85 mm 转换垫圈。垫圈要求由塔基和具体飞轮型号共同决定，不能把所有 10 速飞轮套用同一规则。</p>', 60),
        ('en', 'Can a Campagnolo N3W freehub accept older Campagnolo road cassettes?', '<p>N3W is 4.4 mm shorter than the classic Campagnolo body. The traditional Campagnolo 11-speed and 12-speed road cassette options listed in the engine require the AC21-N3W adapter sleeve and its extended lockring on N3W; a loose spacer alone does not replace that kit. Native Ekar 13-speed 9T or 10T-start N3W cassettes fit directly without the extension. Confirm the adapter and lockring requirements for the exact cassette.</p>', 70),
        ('zh_cn', 'Campagnolo N3W 塔基能安装传统公路飞轮吗？', '<p>N3W 比经典 Campagnolo 塔基短 4.4 mm。引擎中列出的传统 Campagnolo 11 速、12 速公路飞轮安装在 N3W 上，需要 AC21-N3W 延长花键套筒及配套加长锁环，不能用普通垫圈代替这套转接件。原生 Ekar 13 速、9T 或 10T 起跳的 N3W 飞轮可直接安装，无需延长套筒。具体转接件与锁环要求应按飞轮型号核对。</p>', 70),
        ('en', 'Can a third-party cassette use the same freehub as my Shimano or SRAM drivetrain?', '<p>It can when the actual cassette interface matches, but the drivetrain brand does not prove that match. Third-party cassettes may use HG, Micro Spline or XD even at the same speed count. Select the exact brand and cassette specification in the engine, then check the supported freehub options, spacers and mechanical notes. A freehub fit result confirms the mounting interface; also check the cassette manufacturer specifications for chain, shifting and derailleur compatibility.</p>', 80),
        ('zh_cn', '第三方飞轮可以使用 Shimano 或 SRAM 套件对应的塔基吗？', '<p>实际飞轮接口匹配时可以，但传动套件品牌不能证明接口一定匹配。相同速数的第三方飞轮也可能分别采用 HG、Micro Spline 或 XD。应在引擎中选择具体品牌与飞轮规格，再核对支持的塔基、垫圈及机械配合说明。塔基配合结果用于确认安装接口，链条、变速与后拨兼容性仍应按飞轮厂商规格核对。</p>', 80),
        ('en', 'Can I convert my existing hub from HG to Micro Spline, XD or XDR by changing the freehub body?', '<p>Only if the hub manufacturer offers a compatible conversion body for that exact hub model and generation. The cassette interface and the connection between the freehub body and hub are separate compatibility checks. Confirm the hub engagement mechanism, axle, bearings, seals and end caps required by the conversion kit before ordering. A freehub body with the right cassette standard is not automatically interchangeable across hub brands or models.</p>', 90),
        ('zh_cn', '更换塔基就能把现有 HG 花鼓转换为 Micro Spline、XD 或 XDR 吗？', '<p>只有花鼓厂商提供适用于该具体型号及代次的转换塔基时才可以。飞轮与塔基的接口配合，以及塔基与花鼓的连接配合，是两项不同的检查。订购前需确认转换套件要求的啮合机构、轴心、轴承、密封与端盖；飞轮标准相同不代表塔基可在不同品牌或型号的花鼓之间通用。</p>', 90)
)
INSERT INTO faqs (
    page_id, locale, question, answer,
    answer_image_url, answer_image_alt, answer_image_width, answer_image_height,
    status, "order", created_at, updated_at
)
SELECT
    page.page_id, seed_faqs.locale, seed_faqs.question, seed_faqs.answer,
    '', '', 0, 0, 'published', seed_faqs.sort_order, NOW(), NOW()
FROM seed_faqs
JOIN faq_pages page ON page.route_path = '/tools/freehub-cassette-fitment-engine'
    AND page.locale = seed_faqs.locale AND page.deleted_at IS NULL
WHERE NOT EXISTS (
    SELECT 1 FROM faqs existing
    WHERE existing.page_id = page.page_id
      AND existing.locale = seed_faqs.locale
);
