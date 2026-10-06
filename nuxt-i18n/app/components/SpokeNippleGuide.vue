<template>
  <div class="spoke-nipple-guide" aria-label="外置与内置辐条帽几何剖面与测长基准图解">
  <!-- ================= 视图 01: 全景并排对比视图 (DUAL COMPARISON) ================= -->
  <section id="view-dual" class="main-blueprint-card">
    <div class="schematic-header-bar">
      <span class="schematic-header-title">07-A: 外置与内置式辐条帽车圈安装剖面全景对账 (DUAL-FLANGE SECTION COMPARISON)</span>
      <span class="schematic-header-badge">ISO 128 CAD SPEC / SECTION A-A & B-B</span>
    </div>

    <div class="blueprint-split-grid">
      <!-- 左侧：高精矢量图纸视口 -->
      <div class="svg-viewport-wrapper">
        <svg viewBox="0 0 760 480" preserveAspectRatio="xMidYMid meet">
          <defs>
            <marker id="nipple-arrow-emerald" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
              <path d="M 0 1.5 L 10 5 L 0 8.5 z" fill="#059669" />
            </marker>
            <marker id="nipple-arrow-amber" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
              <path d="M 0 1.5 L 10 5 L 0 8.5 z" fill="#d97706" />
            </marker>
            <marker id="nipple-arrow-sky" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
              <path d="M 0 1.5 L 10 5 L 0 8.5 z" fill="#0284c7" />
            </marker>
            <marker id="nipple-arrow-slate" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
              <path d="M 0 1.5 L 10 5 L 0 8.5 z" fill="#475569" />
            </marker>
          </defs>

          <!-- 中间主垂直分割虚线 -->
          <line x1="380" y1="20" x2="380" y2="460" stroke="rgba(15, 23, 42, 0.15)" stroke-width="1.5" stroke-dasharray="6 4" />

          <!-- ================= 左侧半场：外置式辐条帽 (EXTERNAL NIPPLE) ================= -->
          <g id="group-external-side">
            <!-- 栏目标题 -->
            <rect x="30" y="24" width="320" height="24" rx="6" fill="#f1f5f9" stroke="#cbd5e1" stroke-width="1" />
            <text x="190" y="40" fill="#0f172a" font-family="var(--tz-font-ui)" font-size="10.5" font-weight="900" text-anchor="middle" letter-spacing="0.06em">
              【外置式】SAPIM POLYAX 14mm 测长基准
            </text>

            <!-- 车圈截面 (外置侧) -->
            <!-- 胎唇卡槽与外圈壁 -->
            <path d="M 50 80 L 80 80 C 90 80 95 90 100 105 L 120 120 L 260 120 L 280 105 C 285 90 290 80 300 80 L 330 80 L 330 95 L 310 115 L 295 180 L 260 270 L 120 270 L 85 180 L 70 115 L 50 95 Z" fill="#e2e8f0" stroke="#475569" stroke-width="2" />

            <!-- 车圈内壁条帽承托座底面孔位 -->
            <rect x="120" y="270" width="55" height="18" fill="#cbd5e1" stroke="#475569" stroke-width="1.5" />
            <rect x="205" y="270" width="55" height="18" fill="#cbd5e1" stroke="#475569" stroke-width="1.5" />

            <!-- 外置条帽实体 (金铜色) -->
            <!-- 条帽头部 (承托在车圈底面内部) + 条帽身 (外露在车圈外部) -->
            <path d="M 160 245 L 220 245 C 225 245 228 250 228 258 L 228 270 L 202 270 L 202 360 L 178 360 L 178 270 L 152 270 L 152 258 C 152 250 155 245 160 245 Z" fill="#fef3c7" stroke="#d97706" stroke-width="2" />

            <!-- 条帽顶端一字槽 -->
            <rect x="182" y="245" width="16" height="10" fill="#ffffff" stroke="#b45309" stroke-width="1" />
            <!-- 内部螺纹腔通道 -->
            <rect x="183" y="255" width="14" height="105" fill="#ffffff" stroke="#475569" stroke-width="1" stroke-dasharray="3 2" />

            <!-- 辐条实体 (黑色圆柱线段) - 满牙旋入至一字刀槽底部 -->
            <line x1="190" y1="255" x2="190" y2="445" stroke="#0f172a" stroke-width="4.5" stroke-linecap="round" />
            <!-- 辐条螺纹区段标记 (绿色虚线) -->
            <line x1="190" y1="255" x2="190" y2="350" stroke="#059669" stroke-width="3" stroke-dasharray="2 2" />

            <!-- 外露方形扳手受力部 -->
            <rect x="175" y="335" width="30" height="25" fill="none" stroke="#d97706" stroke-width="1.5" stroke-dasharray="2 2" />

            <!-- 基准线 ①：车圈条帽承托座底面 (ERD Datum) -->
            <line x1="60" y1="270" x2="330" y2="270" stroke="#059669" stroke-width="2" stroke-dasharray="5 3" />

            <!-- 序号 ① 指向承托座底 -->
            <line x1="60" y1="225" x2="110" y2="225" stroke="#059669" stroke-width="1.5" />
            <line x1="110" y1="225" x2="148" y2="268" stroke="#059669" stroke-width="1.5" marker-end="url(#nipple-arrow-emerald)" />
            <g transform="translate(48, 225)">
              <circle cx="0" cy="0" r="11" fill="#059669" stroke="#ffffff" stroke-width="1.8" />
              <text cx="0" cy="4" fill="#ffffff" font-family="var(--tz-font-ui)" font-size="11" font-weight="900" text-anchor="middle">1</text>
            </g>

            <!-- 序号 ②：辐条满牙平齐线 (Slot Base) -->
            <line x1="228" y1="255" x2="310" y2="255" stroke="#0f172a" stroke-width="1.5" stroke-dasharray="4 2" />
            <circle cx="228" cy="255" r="3" fill="#0f172a" />
            <g transform="translate(325, 255)">
              <circle cx="0" cy="0" r="11" fill="#0f172a" stroke="#ffffff" stroke-width="1.8" />
              <text cx="0" cy="4" fill="#ffffff" font-family="var(--tz-font-ui)" font-size="11" font-weight="900" text-anchor="middle">2</text>
            </g>

            <!-- 序号 ⑤：外露调圈扳手夹持段 -->
            <line x1="205" y1="348" x2="270" y2="348" stroke="#d97706" stroke-width="1.5" />
            <g transform="translate(285, 348)">
              <circle cx="0" cy="0" r="11" fill="#d97706" stroke="#ffffff" stroke-width="1.8" />
              <text cx="0" cy="4" fill="#ffffff" font-family="var(--tz-font-ui)" font-size="11" font-weight="900" text-anchor="middle">5</text>
            </g>

            <!-- 底部尺寸说明标注 -->
            <text x="190" y="468" fill="#059669" font-family="var(--tz-font-ui)" font-size="9" font-weight="800" text-anchor="middle">
              ΔL = 0 mm (计算器直接作为 0 基准，无需填长度)
            </text>
          </g>

          <!-- ================= 右侧半场：内置/反装辐条帽 (INVERTED NIPPLE) ================= -->
          <g id="group-internal-side">
            <!-- 栏目标题 -->
            <rect x="410" y="24" width="320" height="24" rx="6" fill="#f0f9ff" stroke="#bae6fd" stroke-width="1" />
            <text x="570" y="40" fill="#0369a1" font-family="var(--tz-font-ui)" font-size="10.5" font-weight="900" text-anchor="middle" letter-spacing="0.06em">
              【内置式】反装深腔沉孔 (需输入实际帽长)
            </text>

            <!-- 车圈截面 (内置侧 - 外壁不开大孔，仅留 2.5mm 辐条极小光孔) -->
            <path d="M 430 80 L 460 80 C 470 80 475 90 480 105 L 500 120 L 640 120 L 660 105 C 665 90 670 80 680 80 L 710 80 L 710 95 L 690 115 L 675 180 L 640 270 L 500 270 L 465 180 L 450 115 L 430 95 Z" fill="#e2e8f0" stroke="#475569" stroke-width="2" />

            <!-- 车圈底壁极小光孔 (仅 2.5mm 紧凑穿孔) -->
            <rect x="500" y="270" width="66" height="18" fill="#cbd5e1" stroke="#475569" stroke-width="1.5" />
            <rect x="574" y="270" width="66" height="18" fill="#cbd5e1" stroke="#475569" stroke-width="1.5" />

            <!-- 胎床顶端工具伸入孔 (Rim Bed Access Hole 约 8mm) -->
            <line x1="550" y1="120" x2="590" y2="120" stroke="#ffffff" stroke-width="3" />

            <!-- 内置反装条帽实体 (天蓝色高对比) -->
            <!-- 注意：条帽方向完全颠倒！喇叭口法兰朝下坐在底座内侧，工具接口端朝上深入胎床 -->
            <path d="M 552 245 L 588 245 C 596 245 602 250 605 258 L 605 270 L 582 270 L 582 170 L 558 170 L 558 270 L 535 270 L 535 258 C 538 250 544 245 552 245 Z" fill="#e0f2fe" stroke="#0284c7" stroke-width="2" />

            <!-- 工具接口示意为六角；实际接口依型号可能为矩形或六角 -->
            <polygon points="562,170 578,170 582,185 578,200 562,200 558,185" fill="#bae6fd" stroke="#0369a1" stroke-width="1.2" />
            <!-- 内部深孔螺纹腔 -->
            <rect x="563" y="195" width="14" height="75" fill="#ffffff" stroke="#475569" stroke-width="1" stroke-dasharray="3 2" />

            <!-- 辐条实体 (黑色圆柱线段) - 必须一路向上深扎进车圈内部咬紧内置条帽 -->
            <line x1="570" y1="195" x2="570" y2="445" stroke="#0f172a" stroke-width="4.5" stroke-linecap="round" />
            <!-- 辐条螺纹深咬合段 (天蓝虚线) -->
            <line x1="570" y1="195" x2="570" y2="265" stroke="#0284c7" stroke-width="3" stroke-dasharray="2 2" />

            <!-- 专用调节工具路径 (虚线演示工具从胎床插入) -->
            <line x1="570" y1="90" x2="570" y2="165" stroke="#e11d48" stroke-width="2" stroke-dasharray="4 3" marker-end="url(#nipple-arrow-slate)" />
            <text x="585" y="110" fill="#e11d48" font-family="var(--tz-font-ui)" font-size="8.5" font-weight="800">
              专用工具通道（矩形/六角）
            </text>

            <!-- 关键测长落差线：辐条额外延伸量 ΔL -->
            <line x1="190" y1="255" x2="570" y2="255" stroke="#94a3b8" stroke-width="1.2" stroke-dasharray="4 4" />
            <line x1="570" y1="195" x2="680" y2="195" stroke="#0284c7" stroke-width="1.5" stroke-dasharray="4 2" />
            <line x1="660" y1="255" x2="660" y2="195" stroke="#e11d48" stroke-width="2" marker-start="url(#nipple-arrow-emerald)" marker-end="url(#nipple-arrow-emerald)" />
            <text x="672" y="228" fill="#e11d48" font-family="var(--tz-font-ui)" font-size="10" font-weight="900">
              +ΔL (约+8~10mm)
            </text>

            <!-- 序号 ③：内置条帽内腔反向承托面 -->
            <line x1="435" y1="255" x2="495" y2="255" stroke="#0284c7" stroke-width="1.5" />
            <line x1="495" y1="255" x2="538" y2="268" stroke="#0284c7" stroke-width="1.5" marker-end="url(#nipple-arrow-sky)" />
            <g transform="translate(425, 255)">
              <circle cx="0" cy="0" r="11" fill="#0284c7" stroke="#ffffff" stroke-width="1.8" />
              <text cx="0" cy="4" fill="#ffffff" font-family="var(--tz-font-ui)" font-size="11" font-weight="900" text-anchor="middle">3</text>
            </g>

            <!-- 序号 ④：深腔有效螺牙咬合区 -->
            <line x1="570" y1="195" x2="630" y2="175" stroke="#0284c7" stroke-width="1.5" />
            <circle cx="570" cy="195" r="3.5" fill="#0284c7" />
            <g transform="translate(645, 175)">
              <circle cx="0" cy="0" r="11" fill="#0284c7" stroke="#ffffff" stroke-width="1.8" />
              <text cx="0" cy="4" fill="#ffffff" font-family="var(--tz-font-ui)" font-size="11" font-weight="900" text-anchor="middle">4</text>
            </g>

            <!-- 序号 ⑥：车圈外侧气动极小光孔 (2.5mm) -->
            <line x1="574" y1="288" x2="630" y2="330" stroke="#475569" stroke-width="1.5" />
            <circle cx="574" cy="288" r="3" fill="#475569" />
            <g transform="translate(645, 330)">
              <circle cx="0" cy="0" r="11" fill="#475569" stroke="#ffffff" stroke-width="1.8" />
              <text cx="0" cy="4" fill="#ffffff" font-family="var(--tz-font-ui)" font-size="11" font-weight="900" text-anchor="middle">6</text>
            </g>

            <!-- 序号 ⑦：胎垫密封层与扳手工作空隙 -->
            <line x1="570" y1="170" x2="630" y2="135" stroke="#e11d48" stroke-width="1.5" />
            <g transform="translate(645, 135)">
              <circle cx="0" cy="0" r="11" fill="#e11d48" stroke="#ffffff" stroke-width="1.8" />
              <text cx="0" cy="4" fill="#ffffff" font-family="var(--tz-font-ui)" font-size="11" font-weight="900" text-anchor="middle">7</text>
            </g>

            <!-- 底部尺寸说明标注 -->
            <text x="570" y="468" fill="#0284c7" font-family="var(--tz-font-ui)" font-size="9" font-weight="800" text-anchor="middle">
              必须输入实际帽长（如10/12/14mm），严禁用外置14mm代替！
            </text>
          </g>
        </svg>
      </div>

      <!-- 右侧：图纸技术规范与序号索引 (完全对齐图1结构) -->
      <div class="schematic-legend-panel">
        <div class="legend-panel-header">
          <span class="legend-panel-title">图纸技术规范与序号索引 (ENGINEERING SPECIFICATION & INDEX)</span>
          <span class="legend-panel-badge">ISO 128 / NIPPLE DATUM</span>
        </div>

        <div class="legend-list">
          <!-- 序号 1 -->
          <div class="legend-item">
            <div class="legend-badge-num num-emerald">1</div>
            <div class="legend-content">
              <div class="legend-item-title">外置条帽承托座底面 (External Nipple Bed / Seat)</div>
              <div class="legend-item-desc">车圈内壁与外置条帽喇叭头贴合的承托受力面；计算器按标准 ERD 将此处设为测长初始零基准。</div>
            </div>
          </div>

          <!-- 序号 2 -->
          <div class="legend-item">
            <div class="legend-badge-num num-slate">2</div>
            <div class="legend-content">
              <div class="legend-item-title">外置满扣平齐面 (Slot Base Datum - 14mm 基准)</div>
              <div class="legend-item-desc">辐条末端旋入条帽的理想黄金咬合位置（位于一字刀槽底），标准 14mm 条帽在此处达到 100% 满牙受力。</div>
            </div>
          </div>

          <!-- 序号 3 -->
          <div class="legend-item">
            <div class="legend-badge-num num-sky">3</div>
            <div class="legend-content">
              <div class="legend-item-title">内置条帽内腔反向承托面 (Inverted Nipple Bed)</div>
              <div class="legend-item-desc">内置条帽的喇叭法兰倒置贴合在车圈空腔内壁，受力方向由内向外压紧；条帽主体完全沉入车圈深腔。</div>
            </div>
          </div>

          <!-- 序号 4 -->
          <div class="legend-item">
            <div class="legend-badge-num num-sky">4</div>
            <div class="legend-content">
              <div class="legend-item-title">深腔有效螺牙咬合区 (Deep Thread Engagement Zone)</div>
              <div class="legend-item-desc">由于条帽倒置，螺牙起始点深藏于车圈胎床内部；辐条必须穿越车圈内壁并大幅延伸深入，长度通常需增加 8~10mm。</div>
            </div>
          </div>

          <!-- 序号 5 -->
          <div class="legend-item">
            <div class="legend-badge-num num-amber">5</div>
            <div class="legend-content">
              <div class="legend-item-title">外置方形调圈扳手位 (External Spoke Wrench Flats)</div>
              <div class="legend-item-desc">露在车圈外部的 3.2mm~3.45mm 方形截面，无需拆卸轮胎即可从轮外直接拧动调圈。</div>
            </div>
          </div>

          <!-- 序号 6 -->
          <div class="legend-item">
            <div class="legend-badge-num num-slate">6</div>
            <div class="legend-content">
              <div class="legend-item-title">气动极小光孔 (Aerodynamic Smooth Bore 2.5mm)</div>
              <div class="legend-item-desc">内置式车圈外侧无任何条帽外露，仅留穿条微孔，极大降低风阻扰流并保留碳布层完整强度。</div>
            </div>
          </div>

          <!-- 序号 7 -->
          <div class="legend-item">
            <div class="legend-badge-num num-rose">7</div>
            <div class="legend-content">
              <div class="legend-item-title">胎床专用调节工具通道 (Tire Bed Tool Port)</div>
              <div class="legend-item-desc">接口形状依辐条帽型号而异，有些为矩形，有些为六角；需从胎床侧使用匹配的专用工具。辐条过长会顶住工具通道导致无法调节，过短则可能无法充分咬合。</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部图解状态条 (完全对齐图1) -->
    <div class="diagram-info-bar">
      <div class="diagram-info-text">
        <strong>工程基准 ⑦：外置与内置辐条帽几何基准与螺纹咬合深度差异 (Nipple Bed & Thread Engagement)</strong>
        <p>标明外置式以 SAPIM POLYAX 14mm 设为零基准（无需额外填长）；内置式必须依据实际条帽长度输入测长补偿值，严禁混淆！</p>
      </div>
      <span class="header-badge" style="background: rgba(5, 150, 105, 0.08); color: #047857; border-color: rgba(5, 150, 105, 0.25);">
        测长基准已对齐
      </span>
    </div>
  </section>

  <!-- ================= 计算器第7步实操黄金法则与换算看板 ================= -->
  <section class="calculator-action-banner">
    <div class="banner-rule-card">
      <div class="banner-rule-badge">14</div>
      <div>
        <div class="banner-rule-title">为什么外置式不需要额外输入长度？</div>
        <p class="banner-rule-text">
          本计算器的底层力学引擎是以 <strong>SAPIM POLYAX 14 mm 外置辐条帽</strong> 作为测量基准建立的。在标准外置装配下，辐条旋入一字刀槽底部即为 100% 满扣，公式已直接包含此项常数，因此外置式无需用户填写长度。
        </p>
      </div>
    </div>

    <div class="banner-rule-card">
      <div class="banner-rule-badge" style="background: var(--accent-rose);">⚠️</div>
      <div>
        <div class="banner-rule-title" style="color: #9f1239;">为什么内置式必须填写实际型号长度？</div>
        <p class="banner-rule-text" style="color: #881337;">
          内置辐条帽完全反装在车圈内部，不同品牌和型号（如 Sapim Inverted 10mm、ENVE 内置款 10mm 等）内部有效咬合深浅差异极大；工具接口也可能是矩形或六角。辐条需要额外深入车圈内部 <strong>+6mm 至 +10mm</strong>。<strong>绝对不能直接用外置的 14mm 代替</strong>，否则下料辐条可能过短，造成咬合不足。
        </p>
      </div>
    </div>
  </section>

  <!-- ================= 常见品牌内置辐条帽技术规格对账表 ================= -->
  <section class="spec-table-container">
    <div style="margin-bottom: 12px; display: flex; justify-content: space-between; align-items: center;">
      <h3 style="font-size: 11px; font-weight: 900; text-transform: uppercase; letter-spacing: 0.06em; color: var(--text-main);">
        常用高阶轮组内置/外置条帽工程参数对照表 (INDUSTRY MASTER REFERENCE)
      </h3>
      <span class="header-badge">SAPIM / ENVE / DT SWISS</span>
    </div>

    <table class="spec-table">
      <thead>
        <tr>
          <th>品牌与型号 (Model)</th>
          <th>结构类型 (Type)</th>
          <th>标称长度 (Length)</th>
          <th>调圈工具规格 (Tool Interface)</th>
          <th>车圈外侧穿孔直径 (Rim Outer Hole)</th>
          <th>计算器建议输入 (Input Value)</th>
          <th>技师工程建议 (Builder Notes)</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><strong>SAPIM Polyax Brass / Alloy</strong></td>
          <td><span class="header-badge badge-primary">外置式 (Standard)</span></td>
          <td><code>14.0 mm</code></td>
          <td>3.23 mm 方口外置扳手</td>
          <td>Ø 4.2 mm ~ 4.5 mm</td>
          <td><strong>选择外置式即可</strong></td>
          <td>行业黄金基准，自动补偿满扣位置</td>
        </tr>
        <tr>
          <td><strong>SAPIM Polyax Inverted</strong></td>
          <td><span class="header-badge" style="background: rgba(2, 132, 199, 0.08); color: #0284c7; border-color: rgba(2, 132, 199, 0.2);">内置式 (Hidden)</span></td>
          <td><code>10.0 mm</code></td>
          <td>依具体型号选配（矩形或六角）</td>
          <td>Ø 2.3 mm ~ 2.5 mm 微孔</td>
          <td><strong>输入 10 mm</strong></td>
          <td>常见于轻量化爬坡碳圈，辐条增大约 +7.5mm</td>
        </tr>
        <tr>
          <td><strong>SAPIM Polyax Inverted</strong></td>
          <td><span class="header-badge" style="background: rgba(2, 132, 199, 0.08); color: #0284c7; border-color: rgba(2, 132, 199, 0.2);">内置式 (Hidden)</span></td>
          <td><code>12.0 mm</code></td>
          <td>依具体型号选配（矩形或六角）</td>
          <td>Ø 2.5 mm 微孔</td>
          <td><strong>输入 12 mm</strong></td>
          <td>深框气动轮组高刚性推荐，辐条增大约 +9.0mm</td>
        </tr>
        <tr>
          <td><strong>ENVE Internal Brass / Alloy</strong></td>
          <td><span class="header-badge" style="background: rgba(2, 132, 199, 0.08); color: #0284c7; border-color: rgba(2, 132, 199, 0.2);">内置式 (Hidden)</span></td>
          <td><code>10.0 mm</code></td>
          <td>依具体型号选配（矩形或六角）</td>
          <td>Ø 2.4 mm 模压成型孔 (无钻孔)</td>
          <td><strong>输入 10 mm</strong></td>
          <td>必须配合 ENVE 专用球面垫片使用，防拉脱</td>
        </tr>
        <tr>
          <td><strong>DT SWISS Squorx Pro Head</strong></td>
          <td><span class="header-badge badge-primary">半隐藏外置 (Hybrid)</span></td>
          <td><code>15.0 mm</code></td>
          <td>Torx E-10 / 3.25mm 双接口</td>
          <td>Ø 4.3 mm</td>
          <td><strong>按外置式计算</strong></td>
          <td>具备内侧星形头，但安装面仍为外置底座</td>
        </tr>
      </tbody>
    </table>
  </section>

  <!-- 展示型图解：不承载独立表单或计算逻辑。 -->
  </div>
</template>

<style scoped>
/* ==========================================================================
       ERP UDS 工业设计系统规范 (v1.0) - 辐条计算器第7步工程图纸蓝图标准
       ========================================================================== */
    .spoke-nipple-guide {
      /* 基础淡色工业工程图纸灰阶 (Light Precision Blueprint & Technical Slate) */
      --bg-base: #f8fafc;            /* Slate-50 极清爽明净的工业工程底 */
      --bg-surface: #ffffff;         /* 纯白结构底板 */
      --bg-card: #ffffff;            /* 卡片纯白 */
      --bg-card-hover: #f8fafc;
      --bg-inset: #f1f5f9;           /* Slate-100 浅灰内嵌沉板 */
      --border-line: rgba(15, 23, 42, 0.08);
      --border-dashed: 1px dashed rgba(15, 23, 42, 0.16);
      --border-focus: #0f172a;

      /* 严谨工程语义配色 (Strict ERP UDS Semantics) */
      --accent-primary: #059669;     /* Tanzanite 翡翠精工绿 Emerald-600 */
      --accent-amber: #d97706;       /* 尺寸标注 / 铜条帽 琥珀金 Amber-600 */
      --accent-sky: #0284c7;         /* 内置式反装蓝 Sky-600 */
      --accent-rose: #e11d48;        /* 临界干涉 / 警示工程红 Rose-600 */
      --accent-steel: #475569;       /* 机械冷钢中性标尺 Slate-600 */

      /* 高可读性工业排版深色色阶 */
      --text-main: #0f172a;          /* Slate-900 锐利深黑高清晰文字 */
      --text-secondary: #334155;     /* Slate-700 次级信息 */
      --text-muted: #64748b;         /* Slate-500 辅助标签 */
      --text-dim: #94a3b8;           /* Slate-400 弱化网格与代码 */
    }

    .spoke-nipple-guide,
  .spoke-nipple-guide * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
    }

    .spoke-nipple-guide {
      background-color: var(--bg-base);
      background-image:
        linear-gradient(rgba(15, 23, 42, 0.035) 1px, transparent 1px),
        linear-gradient(90deg, rgba(15, 23, 42, 0.035) 1px, transparent 1px);
      background-size: 28px 28px;
      color: var(--text-main);
      font-family: var(--tz-font-ui);
      min-height: 100vh;
      padding: 24px;
      display: flex;
      flex-direction: column;
      gap: 20px;
      max-width: 1400px;
      margin: 0 auto;
    }

    .header-badge {
      display: inline-flex;
      align-items: center;
      padding: 3px 10px;
      border-radius: 9999px;
      background: rgba(15, 23, 42, 0.05);
      color: #334155;
      border: 1px solid rgba(15, 23, 42, 0.12);
      font-size: 8.5px;
      font-family: var(--tz-font-ui);
      font-weight: 800;
      letter-spacing: 0.08em;
    }

    .header-badge.badge-primary {
      background: rgba(5, 150, 105, 0.08);
      color: #047857;
      border-color: rgba(5, 150, 105, 0.25);
    }

    /* 主工程视口大卡片 (完全还原图1布局) */
    .main-blueprint-card {
      background: #ffffff;
      border: var(--border-dashed);
      border-radius: 24px;
      padding: 24px;
      display: flex;
      flex-direction: column;
      gap: 20px;
      box-shadow: 0 10px 25px -5px rgba(15, 23, 42, 0.04);
    }

    /* 顶部标题区 (图1同款 01: EFFECTIVE RIM DIAMETER SCHEMATIC) */
    .schematic-header-bar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding-bottom: 8px;
      border-bottom: 1px dashed rgba(15, 23, 42, 0.1);
    }

    .schematic-header-title {
      font-size: 11.5px;
      font-weight: 900;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      color: var(--text-main);
    }

    .schematic-header-badge {
      font-family: var(--tz-font-ui);
      font-size: 8.5px;
      font-weight: 800;
      padding: 2px 8px;
      border-radius: 4px;
      background: rgba(15, 23, 42, 0.05);
      color: var(--text-muted);
    }

    /* 主网格：左侧图纸视口 + 右侧技术序号索引表 (对齐图1标准 55% / 45%) */
    .blueprint-split-grid {
      display: grid;
      grid-template-columns: 1.15fr 0.85fr;
      gap: 24px;
      align-items: start;
    }

    @media (max-width: 1100px) {
      .blueprint-split-grid {
        grid-template-columns: 1fr;
      }
    }

    /* 左侧矢量画布视口容器 */
    .svg-viewport-wrapper {
      background: #f8fafc;
      border: 1px solid rgba(15, 23, 42, 0.08);
      border-radius: 18px;
      padding: 16px;
      display: flex;
      align-items: center;
      justify-content: center;
      position: relative;
      overflow: hidden;
    }

    .svg-viewport-wrapper svg {
      width: 100%;
      height: auto;
      max-height: 480px;
      display: block;
    }

    /* 右侧结构化技术序号索引表 (图1右侧同款) */
    .schematic-legend-panel {
      background: #ffffff;
      border: 1px solid rgba(15, 23, 42, 0.08);
      border-radius: 18px;
      padding: 20px 22px;
      display: flex;
      flex-direction: column;
      gap: 14px;
      box-shadow: 0 2px 8px rgba(15, 23, 42, 0.02);
    }

    .legend-panel-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      border-bottom: 1px dashed rgba(15, 23, 42, 0.1);
      padding-bottom: 10px;
    }

    .legend-panel-title {
      font-size: 10.5px;
      font-weight: 900;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      color: var(--text-main);
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .legend-panel-badge {
      font-family: var(--tz-font-ui);
      font-size: 8px;
      font-weight: 800;
      padding: 2px 7px;
      border-radius: 4px;
      background: rgba(15, 23, 42, 0.05);
      color: var(--text-dim);
    }

    .legend-list {
      display: flex;
      flex-direction: column;
      gap: 10px;
    }

    .legend-item {
      display: flex;
      align-items: flex-start;
      gap: 12px;
      background: #f8fafc;
      border: 1px solid rgba(15, 23, 42, 0.05);
      border-radius: 12px;
      padding: 11px 14px;
      transition: all 0.2s ease;
      cursor: default;
    }

    .legend-item:hover, .legend-item.highlighted {
      background: #ffffff;
      border-color: rgba(5, 150, 105, 0.4);
      box-shadow: 0 4px 12px -2px rgba(5, 150, 105, 0.08);
      transform: translateX(2px);
    }

    .legend-badge-num {
      width: 22px;
      height: 22px;
      border-radius: 9999px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-family: var(--tz-font-ui);
      font-size: 10px;
      font-weight: 900;
      flex-shrink: 0;
      margin-top: 1px;
    }

    .num-emerald {
      background: var(--accent-primary);
      color: #ffffff;
      box-shadow: 0 2px 5px rgba(5, 150, 105, 0.25);
    }

    .num-amber {
      background: var(--accent-amber);
      color: #ffffff;
      box-shadow: 0 2px 5px rgba(217, 119, 6, 0.25);
    }

    .num-sky {
      background: var(--accent-sky);
      color: #ffffff;
      box-shadow: 0 2px 5px rgba(2, 132, 199, 0.25);
    }

    .num-slate {
      background: #475569;
      color: #ffffff;
    }

    .num-rose {
      background: var(--accent-rose);
      color: #ffffff;
      box-shadow: 0 2px 5px rgba(225, 29, 72, 0.25);
    }

    .legend-content {
      display: flex;
      flex-direction: column;
      gap: 3px;
    }

    .legend-item-title {
      font-size: 10px;
      font-weight: 900;
      color: var(--text-main);
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .legend-item-desc {
      font-size: 8.5px;
      color: var(--text-muted);
      line-height: 1.45;
      font-weight: 500;
    }

    /* 底部图解状态条 (图1同款 diagram-info-bar) */
    .diagram-info-bar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      background: #f8fafc;
      border: 1px solid rgba(15, 23, 42, 0.06);
      border-radius: 14px;
      padding: 12px 18px;
      gap: 16px;
    }

    .diagram-info-text strong {
      font-size: 10px;
      font-weight: 900;
      color: var(--text-main);
      display: block;
      margin-bottom: 2px;
    }

    .diagram-info-text p {
      font-size: 8.5px;
      color: var(--text-muted);
      line-height: 1.4;
    }

    /* 计算器第7步实战计算指导看板 */
    .calculator-action-banner {
      background: linear-gradient(135deg, rgba(217, 119, 6, 0.05), rgba(5, 150, 105, 0.03));
      border: 1px solid rgba(217, 119, 6, 0.2);
      border-radius: 18px;
      padding: 18px 24px;
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 20px;
      align-items: center;
    }

    @media (max-width: 900px) {
      .calculator-action-banner {
        grid-template-columns: 1fr;
      }
    }

    .banner-rule-card {
      display: flex;
      align-items: flex-start;
      gap: 14px;
    }

    .banner-rule-badge {
      width: 32px;
      height: 32px;
      border-radius: 10px;
      background: var(--accent-amber);
      color: #ffffff;
      display: flex;
      align-items: center;
      justify-content: center;
      font-family: var(--tz-font-ui);
      font-size: 13px;
      font-weight: 900;
      flex-shrink: 0;
      box-shadow: 0 4px 10px rgba(217, 119, 6, 0.25);
    }

    .banner-rule-title {
      font-size: 10.5px;
      font-weight: 900;
      color: #92400e;
      text-transform: uppercase;
      letter-spacing: 0.04em;
      margin-bottom: 4px;
    }

    .banner-rule-text {
      font-size: 8.5px;
      color: #78350f;
      line-height: 1.5;
    }

    /* 常见内置条帽规格对账表 */
    .spec-table-container {
      background: #ffffff;
      border: var(--border-dashed);
      border-radius: 20px;
      padding: 20px 24px;
      overflow-x: auto;
    }

    .spec-table {
      width: 100%;
      border-collapse: collapse;
      font-size: 9px;
      text-align: left;
    }

    .spec-table th {
      background: #f8fafc;
      padding: 10px 14px;
      font-weight: 900;
      text-transform: uppercase;
      letter-spacing: 0.06em;
      color: var(--text-muted);
      border-bottom: 1px dashed rgba(15, 23, 42, 0.12);
      font-family: var(--tz-font-ui);
    }

    .spec-table td {
      padding: 12px 14px;
      border-bottom: 1px solid rgba(15, 23, 42, 0.05);
      color: var(--text-secondary);
    }

    .spec-table tr:hover td {
      background: #f8fafc;
    }

    .spec-table code {
      font-family: var(--tz-font-ui);
      background: rgba(15, 23, 42, 0.05);
      padding: 2px 6px;
      border-radius: 4px;
      font-size: 8.5px;
      color: #0f172a;
      font-weight: 800;
    }

/* The reference document is embedded below the Step 5 input panel. */
.spoke-nipple-guide {
  width: 100%;
  max-width: none;
  min-height: 0;
  margin: 1.5rem 0 0;
  padding: 0;
  gap: 12px;
  overflow: hidden;
  border: 0;
  border-radius: 0;
  isolation: isolate;
}

@media (max-width: 767px) {
  .spoke-nipple-guide {
    margin-top: 1rem;
  }
}
</style>
