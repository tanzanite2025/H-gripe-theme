<template>
  <div class="spoke-physics-diagrams" aria-label="辐条几何与物理图解工程蓝图">
    <button
      type="button"
      class="physics-collapse-toggle"
      aria-controls="spoke-physics-diagrams-content"
      :aria-expanded="!isCollapsed"
      @click="isCollapsed = !isCollapsed"
    >
      <span>辐条几何与物理图解</span>
      <span class="physics-collapse-toggle__state">
        {{ isCollapsed ? '展开图解' : '收起图解' }}
        <span aria-hidden="true">{{ isCollapsed ? '⌄' : '⌃' }}</span>
      </span>
    </button>

    <div id="spoke-physics-diagrams-content" class="physics-collapse-content" v-show="!isCollapsed">
      <!-- 中间图解看板交互区 -->
      <section class="schematic-showcase">
    <!-- 图解快速切换 Tab -->
    <nav class="schematic-nav">
      <button class="schematic-tab" id="tab-btn-erd" :class="{ active: activeDiagram === 'erd' }" @click="activeDiagram = 'erd'">
        <span class="num">01</span>
        <span>轮圈 ERD 定义</span>
      </button>
      <button class="schematic-tab" id="tab-btn-pcd" :class="{ active: activeDiagram === 'pcd' }" @click="activeDiagram = 'pcd'">
        <span class="num">02</span>
        <span>花鼓 PCD 与中心距</span>
      </button>
      <button class="schematic-tab" id="tab-btn-hole" :class="{ active: activeDiagram === 'hole' }" @click="activeDiagram = 'hole'">
        <span class="num">03</span>
        <span>孔径内缘咬合</span>
      </button>
      <button class="schematic-tab" id="tab-btn-sp" :class="{ active: activeDiagram === 'sp' }" @click="activeDiagram = 'sp'">
        <span class="num">04</span>
        <span>直拉切线槽位</span>
      </button>
      <button class="schematic-tab" id="tab-btn-stretch" :class="{ active: activeDiagram === 'stretch' }" @click="activeDiagram = 'stretch'">
        <span class="num">05</span>
        <span>高张力弹性拉伸</span>
      </button>
      <button class="schematic-tab" id="tab-btn-drill" :class="{ active: activeDiagram === 'drill' }" @click="activeDiagram = 'drill'">
        <span class="num">06</span>
        <span>轮圈交错钻孔</span>
      </button>
      <button class="schematic-tab" id="tab-btn-interlace" :class="{ active: activeDiagram === 'interlace' }" @click="activeDiagram = 'interlace'">
        <span class="num">07</span>
        <span>交叉压条折线</span>
      </button>
    </nav>

    <!-- 矢量画布与技术索引视口 -->
    <div class="canvas-viewport">
      <!-- SVG 全局共享 Defs 标记与滤镜 -->
      <svg style="position: absolute; width: 0; height: 0;">
        <defs>
          <marker id="arrow-slate" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
            <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#475569" />
          </marker>
          <marker id="arrow-rose" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
            <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#e11d48" />
          </marker>
          <marker id="arrow-amber" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
            <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#d97706" />
          </marker>
          <marker id="arrow-emerald" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
            <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#059669" />
          </marker>
        </defs>
      </svg>

      <!-- ==============================================
           图解 01: 轮圈有效直径 ERD 全景与截面几何解剖图
           ============================================== -->
      <div class="diagram-panel" id="diagram-erd" :class="{ active: activeDiagram === 'erd' }">
        <svg viewBox="0 0 760 380" preserveAspectRatio="xMidYMid meet">
          <!-- 标题栏 (通用代号) -->
          <text x="380" y="24" fill="#0f172a" font-size="12" font-weight="900" text-anchor="middle" letter-spacing="0.1em">
            01: EFFECTIVE RIM DIAMETER (ERD) SCHEMATIC
          </text>

          <!-- 左右分割虚线 -->
          <line x1="365" y1="35" x2="365" y2="365" stroke="rgba(15, 23, 42, 0.12)" stroke-dasharray="4 4" />

          <!-- ================= 左侧：车圈全景极轴视图 ================= -->
          <!-- 车圈最外径 OD 虚线 -->
          <circle cx="180" cy="205" r="135" fill="none" stroke="#cbd5e1" stroke-width="2" stroke-dasharray="6 4" />
          <!-- 车圈真空胎唇卡槽 (BSD 622mm) -->
          <circle cx="180" cy="205" r="128" fill="none" stroke="#94a3b8" stroke-width="1.8" />
          <!-- 车圈条帽着力底面 (ERD 对应圆周，高亮翡翠绿) -->
          <circle cx="180" cy="205" r="114" fill="rgba(5, 150, 105, 0.03)" stroke="#059669" stroke-width="2.5" />
          <!-- 车圈内壁 -->
          <circle cx="180" cy="205" r="92" fill="#f8fafc" stroke="#cbd5e1" stroke-width="2" />

          <!-- 轮组轴心十字标 -->
          <line x1="180" y1="185" x2="180" y2="225" stroke="#64748b" stroke-width="1.5" />
          <line x1="160" y1="205" x2="200" y2="205" stroke="#64748b" stroke-width="1.5" />
          <circle cx="180" cy="205" r="4" fill="#475569" />

          <!-- 对径两端条帽座定位点 (12点钟与6点钟) -->
          <circle cx="180" cy="91" r="5" fill="#059669" />
          <line x1="165" y1="91" x2="195" y2="91" stroke="#059669" stroke-width="2" />
          <circle cx="180" cy="319" r="5" fill="#059669" />
          <line x1="165" y1="319" x2="195" y2="319" stroke="#059669" stroke-width="2" />

          <!-- 穿过轴心的 ERD 连续贯通尺寸线 -->
          <line x1="180" y1="98" x2="180" y2="312" stroke="#059669" stroke-width="2" marker-start="url(#arrow-emerald)" marker-end="url(#arrow-emerald)" />
          
          <!-- 尺寸代号徽标牌 -->
          <rect x="120" y="193" width="120" height="24" rx="6" fill="#ffffff" stroke="#059669" stroke-width="1.5" />
          <text x="180" y="209" fill="#059669" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">
            ERD
          </text>

          <!-- 序号 ①：ERD 极轴尺寸 -->
          <g transform="translate(180, 150)">
            <circle cx="0" cy="0" r="11" fill="#059669" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">1</text>
          </g>

          <!-- 序号 ②：条帽承托座底 -->
          <g transform="translate(225, 91)">
            <line x1="-20" y1="0" x2="-8" y2="0" stroke="#059669" stroke-width="1.5" />
            <circle cx="0" cy="0" r="11" fill="#059669" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">2</text>
          </g>

          <!-- 序号 ③：车圈最外径 OD -->
          <g transform="translate(325, 175)">
            <line x1="-15" y1="15" x2="0" y2="0" stroke="#94a3b8" stroke-width="1.5" />
            <circle cx="0" cy="0" r="11" fill="#475569" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">3</text>
            <text x="16" y="4" fill="#64748b" font-family="var(--font-mono)" font-size="9" font-weight="800">OD</text>
          </g>

          <!-- 序号 ④：胎唇卡槽 BSD -->
          <g transform="translate(325, 235)">
            <line x1="-18" y1="-8" x2="0" y2="0" stroke="#94a3b8" stroke-width="1.5" />
            <circle cx="0" cy="0" r="11" fill="#475569" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">4</text>
            <text x="16" y="4" fill="#64748b" font-family="var(--font-mono)" font-size="9" font-weight="800">BSD</text>
          </g>

          <!-- ================= 右侧：截面解剖特写 ================= -->
          <!-- 车圈双层壁截面 -->
          <path d="M 405 85 L 430 85 C 440 85 445 95 450 105 L 470 115 L 650 115 L 670 105 C 675 95 680 85 690 85 L 715 85 L 715 100 L 695 115 L 685 160 L 655 230 L 465 230 L 435 160 L 425 115 L 405 100 Z" fill="#e2e8f0" stroke="#475569" stroke-width="2" />
          <rect x="465" y="230" width="70" height="15" fill="#cbd5e1" stroke="#475569" stroke-width="1.5" />
          <rect x="585" y="230" width="70" height="15" fill="#cbd5e1" stroke="#475569" stroke-width="1.5" />

          <!-- 条帽实体 (金铜色) -->
          <path d="M 525 205 L 595 205 C 600 205 605 210 605 218 L 605 230 L 575 230 L 575 295 L 545 295 L 545 230 L 515 230 L 515 218 C 515 210 520 205 525 205 Z" fill="#fef3c7" stroke="#d97706" stroke-width="2" />
          <rect x="550" y="205" width="20" height="10" fill="#ffffff" stroke="#b45309" stroke-width="1" />
          <rect x="552" y="212" width="16" height="85" fill="#ffffff" stroke="#475569" stroke-width="1" stroke-dasharray="3 2" />
          <line x1="560" y1="297" x2="560" y2="365" stroke="#0f172a" stroke-width="3.5" stroke-linecap="round" />

          <!-- 基准线：条帽承托座底面 -->
          <line x1="430" y1="230" x2="690" y2="230" stroke="#059669" stroke-width="2" stroke-dasharray="5 3" />
          <line x1="430" y1="180" x2="480" y2="180" stroke="#059669" stroke-width="1.5" />
          <line x1="480" y1="180" x2="512" y2="228" stroke="#059669" stroke-width="1.5" marker-end="url(#arrow-emerald)" />
          
          <!-- 序号 ② 指向底面 -->
          <g transform="translate(420, 180)">
            <circle cx="0" cy="0" r="11" fill="#059669" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">2</text>
          </g>

          <!-- 螺牙平齐线与 序号 ⑤ -->
          <line x1="605" y1="212" x2="655" y2="212" stroke="#0f172a" stroke-width="1.5" />
          <circle cx="605" cy="212" r="3.5" fill="#0f172a" />
          <g transform="translate(670, 212)">
            <circle cx="0" cy="0" r="11" fill="#0f172a" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">5</text>
          </g>
        </svg>

        <!-- 结构化技术规范与序号索引 (Legend & Annotation Grid) -->
        <div class="schematic-legend-panel">
          <div class="legend-header">
            <span class="legend-title">图纸技术规范与序号索引 (ENGINEERING SPECIFICATION & INDEX)</span>
            <span class="legend-badge">ISO 128 / ERD DATUM</span>
          </div>
          <div class="legend-grid">
            <div class="legend-item">
              <div class="legend-badge-num num-emerald">1</div>
              <div class="legend-content">
                <div class="legend-item-title">ERD 极轴贯通主尺寸 (Effective Rim Diameter)</div>
                <div class="legend-item-desc">两端对径条帽着力底面穿过轴心线的绝对极轴跨距，是精确求解辐条三角长度的唯一基准值。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-emerald">2</div>
              <div class="legend-content">
                <div class="legend-item-title">条帽承托座底面 (Nipple Bed / Seat)</div>
                <div class="legend-item-desc">车圈内壁与条帽头部贴合的物理承载平面，是 ERD 物理测量的唯一起始面与终止面。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-slate">3</div>
              <div class="legend-content">
                <div class="legend-item-title">轮圈最外边缘 OD (Outer Diameter - 非测量基准)</div>
                <div class="legend-item-desc">车圈外轮廓最外径，距内部条帽底座有数毫米物理落差，绝对不可用作 ERD 测量。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-slate">4</div>
              <div class="legend-content">
                <div class="legend-item-title">胎唇卡槽 BSD (Bead Seat Diameter - 非测量基准)</div>
                <div class="legend-item-desc">外胎钢丝胎唇扣合的阶梯面（例如 700c 对应 622mm），与辐条力学三角几何完全无关。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-slate">5</div>
              <div class="legend-content">
                <div class="legend-item-title">螺牙满牙平齐面 (Nipple Slot Base - 满扣基准)</div>
                <div class="legend-item-desc">辐条末端旋入条帽的理想黄金咬合位置（位于一字刀槽底），保证 100% 满牙受力且不顶破胎垫。</div>
              </div>
            </div>
          </div>
        </div>

        <div class="diagram-info-bar">
          <div class="diagram-info-text">
            <strong>基础几何 ①：轮圈有效直径 ERD (Effective Rim Diameter)</strong>
            <p>标明 ERD 物理测量面仅对应序号 ② 条帽座底，杜绝误量序号 ③ 外沿或序号 ④ 胎唇槽。</p>
          </div>
          <span class="header-badge" style="background: rgba(5, 150, 105, 0.08); color: #047857; border-color: rgba(5, 150, 105, 0.25);">ERD 基准已锁定</span>
        </div>
      </div>

      <!-- ==============================================
           图解 02: 花鼓节圆 PCD 与中心法兰距 (WL / WR / OLD) 空间解剖图
           ============================================== -->
      <div class="diagram-panel" id="diagram-pcd" :class="{ active: activeDiagram === 'pcd' }">
        <svg viewBox="0 0 760 380" preserveAspectRatio="xMidYMid meet">
          <!-- 标题栏 -->
          <text x="380" y="24" fill="#0f172a" font-size="12" font-weight="900" text-anchor="middle" letter-spacing="0.1em">
            02: HUB PCD & FLANGE OFFSETS (WL / WR / OLD) SCHEMATIC
          </text>

          <!-- 左右分割虚线 -->
          <line x1="335" y1="35" x2="335" y2="365" stroke="rgba(15, 23, 42, 0.12)" stroke-dasharray="4 4" />

          <!-- ================= 左侧：花鼓法兰盘正视截面 ================= -->
          <!-- 法兰盘实体轮廓 -->
          <circle cx="165" cy="205" r="95" fill="#f1f5f9" stroke="#475569" stroke-width="2.5" />
          
          <!-- 序号 ②：指向法兰盘最外边 (非 PCD) -->
          <line x1="250" y1="150" x2="275" y2="125" stroke="#e11d48" stroke-width="1.5" />
          <circle cx="250" cy="150" r="3" fill="#e11d48" />
          <g transform="translate(290, 115)">
            <circle cx="0" cy="0" r="11" fill="#e11d48" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">2</text>
          </g>

          <!-- 花鼓筒身与轴孔 -->
          <circle cx="165" cy="205" r="38" fill="#e2e8f0" stroke="#334155" stroke-width="1.8" />
          <circle cx="165" cy="205" r="16" fill="#cbd5e1" stroke="#64748b" stroke-width="2" />
          <circle cx="165" cy="205" r="4" fill="#475569" />

          <!-- PCD 核心基准圆 -->
          <circle cx="165" cy="205" r="68" fill="none" stroke="#d97706" stroke-width="3" stroke-dasharray="6 4" />

          <!-- 穿条孔阵列 -->
          <circle cx="233" cy="205" r="6" fill="#ffffff" stroke="#475569" stroke-width="2" />
          <circle cx="233" cy="205" r="2.5" fill="#d97706" />
          <circle cx="213" cy="253" r="6" fill="#ffffff" stroke="#475569" stroke-width="2" />
          <circle cx="213" cy="253" r="2.5" fill="#d97706" />
          <circle cx="165" cy="273" r="6" fill="#ffffff" stroke="#475569" stroke-width="2" />
          <circle cx="165" cy="273" r="2.5" fill="#d97706" />
          <circle cx="117" cy="253" r="6" fill="#ffffff" stroke="#475569" stroke-width="2" />
          <circle cx="117" cy="253" r="2.5" fill="#d97706" />
          <circle cx="97" cy="205" r="6" fill="#ffffff" stroke="#475569" stroke-width="2" />
          <circle cx="97" cy="205" r="2.5" fill="#d97706" />
          <circle cx="117" cy="157" r="6" fill="#ffffff" stroke="#475569" stroke-width="2" />
          <circle cx="117" cy="157" r="2.5" fill="#d97706" />
          <circle cx="165" cy="137" r="6" fill="#ffffff" stroke="#475569" stroke-width="2" />
          <circle cx="165" cy="137" r="2.5" fill="#d97706" />
          <circle cx="213" cy="157" r="6" fill="#ffffff" stroke="#475569" stroke-width="2" />
          <circle cx="213" cy="157" r="2.5" fill="#d97706" />

          <!-- PCD 尺寸标注线与 序号 ① -->
          <line x1="97" y1="205" x2="233" y2="205" stroke="#d97706" stroke-width="2" marker-start="url(#arrow-amber)" marker-end="url(#arrow-amber)" />
          <rect x="115" y="193" width="100" height="24" rx="5" fill="#ffffff" stroke="#d97706" stroke-width="1.5" />
          <text x="165" y="209" fill="#d97706" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">
            PCD
          </text>
          <g transform="translate(165, 120)">
            <line x1="0" y1="12" x2="0" y2="17" stroke="#d97706" stroke-width="1.5" />
            <circle cx="0" cy="0" r="11" fill="#d97706" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">1</text>
          </g>

          <!-- ================= 右侧：花鼓轴向剖面尺寸 ================= -->
          <!-- 序号 ③：花鼓对称中心基准线 -->
          <line x1="535" y1="55" x2="535" y2="330" stroke="#059669" stroke-width="2" stroke-dasharray="6 3" />
          <g transform="translate(535, 45)">
            <circle cx="0" cy="0" r="11" fill="#059669" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">3</text>
            <text x="16" y="4" fill="#059669" font-family="var(--font-mono)" font-size="9" font-weight="900">CL</text>
          </g>

          <!-- 序号 ④：开档 OLD 跨距 -->
          <line x1="390" y1="85" x2="680" y2="85" stroke="#475569" stroke-width="1.8" marker-start="url(#arrow-slate)" marker-end="url(#arrow-slate)" />
          <rect x="495" y="73" width="80" height="20" rx="4" fill="#ffffff" stroke="#cbd5e1" stroke-width="1" />
          <text x="535" y="87" fill="#0f172a" font-family="var(--font-mono)" font-size="9.5" font-weight="800" text-anchor="middle">
            OLD
          </text>
          <g transform="translate(475, 83)">
            <circle cx="0" cy="0" r="10" fill="#0f172a" stroke="#ffffff" stroke-width="1.5" />
            <text cx="0" cy="3.5" fill="#ffffff" font-family="var(--font-mono)" font-size="9.5" font-weight="900" text-anchor="middle">4</text>
          </g>
          <line x1="390" y1="80" x2="390" y2="260" stroke="#94a3b8" stroke-dasharray="3 3" />
          <line x1="680" y1="80" x2="680" y2="260" stroke="#94a3b8" stroke-dasharray="3 3" />

          <!-- 花鼓轴心本体 -->
          <rect x="390" y="202" width="25" height="36" rx="3" fill="#e2e8f0" stroke="#64748b" stroke-width="1.5" />
          <rect x="456" y="140" width="8" height="160" rx="3" fill="#334155" stroke="#0f172a" stroke-width="1.5" />
          <circle cx="460" cy="160" r="5" fill="#d97706" />
          <circle cx="460" cy="280" r="5" fill="#d97706" />
          <rect x="464" y="208" width="112" height="24" fill="#cbd5e1" stroke="#475569" stroke-width="1.5" />
          <rect x="572" y="150" width="8" height="140" rx="3" fill="#334155" stroke="#0f172a" stroke-width="1.5" />
          <circle cx="576" cy="170" r="5" fill="#d97706" />
          <circle cx="576" cy="270" r="5" fill="#d97706" />

          <!-- 序号 ⑦：塔基机构 -->
          <rect x="580" y="195" width="100" height="50" rx="4" fill="#f1f5f9" stroke="#64748b" stroke-width="1.5" stroke-dasharray="4 2" />
          <g transform="translate(630, 220)">
            <circle cx="0" cy="0" r="11" fill="#475569" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">7</text>
          </g>

          <!-- 序号 ⑤：WL 尺寸线 -->
          <line x1="460" y1="112" x2="535" y2="112" stroke="#0f172a" stroke-width="2" marker-start="url(#arrow-slate)" marker-end="url(#arrow-slate)" />
          <line x1="460" y1="107" x2="460" y2="140" stroke="#94a3b8" stroke-dasharray="2 2" />
          <rect x="465" y="100" width="60" height="20" rx="4" fill="#ffffff" stroke="#0f172a" stroke-width="1.2" />
          <text x="495" y="114" fill="#0f172a" font-family="var(--font-mono)" font-size="9.5" font-weight="900" text-anchor="middle">
            WL
          </text>
          <g transform="translate(440, 112)">
            <circle cx="0" cy="0" r="10" fill="#0f172a" stroke="#ffffff" stroke-width="1.5" />
            <text cx="0" cy="3.5" fill="#ffffff" font-family="var(--font-mono)" font-size="9.5" font-weight="900" text-anchor="middle">5</text>
          </g>

          <!-- 序号 ⑥：WR 尺寸线 -->
          <line x1="535" y1="135" x2="576" y2="135" stroke="#d97706" stroke-width="2" marker-start="url(#arrow-amber)" marker-end="url(#arrow-amber)" />
          <line x1="576" y1="130" x2="576" y2="150" stroke="#d97706" stroke-dasharray="2 2" />
          <line x1="555" y1="135" x2="595" y2="135" stroke="#d97706" stroke-width="1.2" />
          <rect x="595" y="125" width="60" height="20" rx="4" fill="#fffbeb" stroke="#d97706" stroke-width="1.2" />
          <text x="625" y="139" fill="#b45309" font-family="var(--font-mono)" font-size="9.5" font-weight="900" text-anchor="middle">
            WR
          </text>
          <g transform="translate(675, 135)">
            <circle cx="0" cy="0" r="10" fill="#d97706" stroke="#ffffff" stroke-width="1.5" />
            <text cx="0" cy="3.5" fill="#ffffff" font-family="var(--font-mono)" font-size="9.5" font-weight="900" text-anchor="middle">6</text>
          </g>

          <!-- 辐条牵引向量 -->
          <circle cx="535" cy="65" r="4" fill="#059669" />
          <line x1="460" y1="160" x2="535" y2="65" stroke="#475569" stroke-width="2" />
          <line x1="576" y1="170" x2="535" y2="65" stroke="#d97706" stroke-width="2" />
        </svg>

        <!-- 结构化技术规范与序号索引 (Legend & Annotation Grid) -->
        <div class="schematic-legend-panel">
          <div class="legend-header">
            <span class="legend-title">图纸技术规范与序号索引 (ENGINEERING SPECIFICATION & INDEX)</span>
            <span class="legend-badge">ISO 128 / HUB GEOMETRY</span>
          </div>
          <div class="legend-grid">
            <div class="legend-item">
              <div class="legend-badge-num num-amber">1</div>
              <div class="legend-content">
                <div class="legend-item-title">花鼓节圆 PCD (Pitch Circle Diameter)</div>
                <div class="legend-item-desc">穿过法兰上全部穿条孔【几何圆心】的高亮虚线基准圆，是计算辐条空间三角的严谨几何节圆。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-rose">2</div>
              <div class="legend-content">
                <div class="legend-item-title">法兰最外边缘 (Flange Outer Edge - 非 PCD)</div>
                <div class="legend-item-desc">法兰盘最外端金属圆周，外缘距孔心有 3~6mm 结构边距，严禁用作 PCD 计算。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-emerald">3</div>
              <div class="legend-content">
                <div class="legend-item-title">轮组轴向对称中心线 CL (Wheel Centerline)</div>
                <div class="legend-item-desc">车轮在车架开档内的绝对物理对称轴，左右辐条拉力与法兰距测量均以该轴线为基准原点。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-slate">4</div>
              <div class="legend-content">
                <div class="legend-item-title">开档总宽跨距 OLD (Over Locknut Dimension)</div>
                <div class="legend-item-desc">花鼓左右两侧端盖与车架爪片贴合面之间的安装总跨距（如公路 142mm / 山地 Boost 148mm）。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-slate">5</div>
              <div class="legend-content">
                <div class="legend-item-title">中心至左法兰距 WL (Center to Left Flange)</div>
                <div class="legend-item-desc">对称中心线至左法兰孔平面的水平跨距，后轮由于无飞轮占用，WL 跨距较大，出条支撑角较平缓。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-amber">6</div>
              <div class="legend-content">
                <div class="legend-item-title">中心至右法兰距 WR (Center to Right Flange)</div>
                <div class="legend-item-desc">对称中心线至右法兰孔平面的水平跨距；因飞轮塔基占用空间，WR 剧烈压缩使右侧辐条极陡直。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-slate">7</div>
              <div class="legend-content">
                <div class="legend-item-title">塔基驱动侧机构 (Freehub Body)</div>
                <div class="legend-item-desc">飞轮片安装部位，其占据的开档物理宽度迫使右法兰向中心收缩，是左右张力不对称的根源。</div>
              </div>
            </div>
          </div>
        </div>

        <div class="diagram-info-bar">
          <div class="diagram-info-text">
            <strong>基础几何 ②：花鼓节圆 PCD 与中心距 (WL / WR / OLD)</strong>
            <p>标明 PCD 是穿过穿孔中心的虚线圆序号 ①；右侧 WR 序号 ⑥ 显著内缩导致出条陡峭。</p>
          </div>
          <span class="header-badge" style="background: rgba(217, 119, 6, 0.08); color: #b45309; border-color: rgba(217, 119, 6, 0.25);">PCD 物理位置已标明</span>
        </div>
      </div>

      <!-- ==============================================
           图解 03: 法兰孔内缘切点扣减 (Spoke Hole Edge)
           ============================================== -->
      <div class="diagram-panel" id="diagram-hole" :class="{ active: activeDiagram === 'hole' }">
        <svg viewBox="0 0 760 380" preserveAspectRatio="xMidYMid meet">
          <defs>
            <pattern id="hatch-metal-hole-p" width="8" height="8" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
              <rect width="8" height="8" fill="#f1f5f9" />
              <line x1="0" y1="0" x2="0" y2="8" stroke="#cbd5e1" stroke-width="1.6" />
            </pattern>
            <pattern id="hatch-gap-hole-p" width="6" height="6" patternUnits="userSpaceOnUse" patternTransform="rotate(-45)">
              <rect width="6" height="6" fill="#fff1f2" />
              <line x1="0" y1="0" x2="0" y2="6" stroke="#e11d48" stroke-width="1.2" stroke-opacity="0.6" />
            </pattern>
            <linearGradient id="grad-spoke-steel-hp" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#94a3b8" />
              <stop offset="30%" stop-color="#cbd5e1" />
              <stop offset="50%" stop-color="#f8fafc" />
              <stop offset="70%" stop-color="#cbd5e1" />
              <stop offset="100%" stop-color="#64748b" />
            </linearGradient>
            <linearGradient id="grad-spoke-head-p" x1="0" y1="0" x2="1" y2="0">
              <stop offset="0%" stop-color="#64748b" />
              <stop offset="40%" stop-color="#cbd5e1" />
              <stop offset="70%" stop-color="#f8fafc" />
              <stop offset="100%" stop-color="#94a3b8" />
            </linearGradient>
          </defs>

          <!-- 标题栏 -->
          <text x="380" y="24" fill="#0f172a" font-size="12" font-weight="900" text-anchor="middle" letter-spacing="0.08em">
            03: SPOKE HOLE CONTACT CUTAWAY & RADIUS DEDUCTION (dh / 2)
          </text>

          <!-- 垂直分栏虚线 -->
          <line x1="210" y1="35" x2="210" y2="365" stroke="rgba(15, 23, 42, 0.12)" stroke-dasharray="4 4" />
          <line x1="495" y1="35" x2="495" y2="365" stroke="rgba(15, 23, 42, 0.12)" stroke-dasharray="4 4" />

          <!-- 视区 1 (左侧)：宏观定位 -->
          <circle cx="110" cy="180" r="58" fill="#f1f5f9" stroke="#475569" stroke-width="2" />
          <circle cx="110" cy="180" r="18" fill="#e2e8f0" stroke="#64748b" stroke-width="1.8" />
          <circle cx="110" cy="180" r="3.5" fill="#475569" />
          <circle cx="110" cy="180" r="42" fill="none" stroke="#d97706" stroke-width="1.5" stroke-dasharray="4 3" />
          <circle cx="110" cy="138" r="4" fill="#334155" />
          <circle cx="140" cy="150" r="4" fill="#334155" />
          <circle cx="152" cy="180" r="4" fill="#334155" />
          <circle cx="140" cy="210" r="4" fill="#334155" />
          <circle cx="110" cy="222" r="4" fill="#334155" />
          <circle cx="80" cy="210" r="4" fill="#334155" />
          <circle cx="68" cy="180" r="4" fill="#334155" />
          <circle cx="80" cy="150" r="4" fill="#334155" />
          <circle cx="110" cy="138" r="9" fill="none" stroke="#e11d48" stroke-width="2" stroke-dasharray="3 2" />
          
          <!-- 放大索引线与 序号 ① -->
          <path d="M 110 138 L 210 95 L 210 265 Z" fill="rgba(225, 29, 72, 0.04)" stroke="rgba(225, 29, 72, 0.2)" stroke-dasharray="3 3" />
          <g transform="translate(110, 100)">
            <circle cx="0" cy="0" r="11" fill="#d97706" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">1</text>
          </g>

          <!-- 视区 2 (中间)：装配剖视 (放大20x) -->
          <path d="M 285 75 L 345 75 L 345 135 Q 345 140 340 140 L 290 140 Q 285 140 285 135 Z" fill="url(#hatch-metal-hole-p)" stroke="#64748b" stroke-width="1.8" />
          <path d="M 285 205 Q 285 200 290 200 L 340 200 Q 345 200 345 205 L 345 265 L 285 265 Z" fill="url(#hatch-metal-hole-p)" stroke="#64748b" stroke-width="1.8" />
          
          <!-- 序号 ⑤：孔壁厚度 t -->
          <line x1="285" y1="66" x2="345" y2="66" stroke="#475569" stroke-width="1.2" marker-start="url(#arrow-slate)" marker-end="url(#arrow-slate)" />
          <text x="315" y="62" fill="#475569" font-size="8" font-family="var(--font-mono)" text-anchor="middle">t = 3.2mm</text>
          <g transform="translate(365, 62)">
            <circle cx="0" cy="0" r="10" fill="#475569" stroke="#ffffff" stroke-width="1.5" />
            <text cx="0" cy="3.5" fill="#ffffff" font-family="var(--font-mono)" font-size="9.5" font-weight="900" text-anchor="middle">5</text>
          </g>

          <!-- 蘑菇头与辐条 -->
          <path d="M 285 125 C 265 135 265 195 285 205 Z" fill="url(#grad-spoke-head-p)" stroke="#475569" stroke-width="1.8" />
          <rect x="285" y="140" width="60" height="44" fill="url(#grad-spoke-steel-hp)" stroke="#475569" stroke-width="1.2" />
          <path d="M 345 140 Q 365 140 375 120 L 440 55" fill="none" stroke="#475569" stroke-width="36" stroke-linecap="butt" />
          <path d="M 345 140 Q 365 140 375 120 L 440 55" fill="none" stroke="url(#grad-spoke-steel-hp)" stroke-width="32" stroke-linecap="butt" />
          
          <!-- 序号 ②：弯头受力咬合点 -->
          <circle cx="345" cy="140" r="7" fill="#059669" stroke="#ffffff" stroke-width="2" />
          <line x1="345" y1="140" x2="385" y2="165" stroke="#059669" stroke-width="1.5" />
          <g transform="translate(405, 175)">
            <circle cx="0" cy="0" r="11" fill="#059669" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">2</text>
          </g>

          <!-- 序号 ④：下方脱空间隙 -->
          <rect x="285" y="184" width="60" height="16" fill="url(#hatch-gap-hole-p)" stroke="#e11d48" stroke-width="1" stroke-dasharray="2 2" />
          <line x1="315" y1="200" x2="315" y2="230" stroke="#e11d48" stroke-width="1.2" />
          <g transform="translate(315, 245)">
            <circle cx="0" cy="0" r="11" fill="#e11d48" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">4</text>
          </g>

          <!-- 视区 3 (右侧)：几何误差解析 -->
          <circle cx="620" cy="180" r="55" fill="#f8fafc" stroke="#475569" stroke-width="2.5" />
          <line x1="620" y1="122" x2="620" y2="238" stroke="#94a3b8" stroke-width="1.2" stroke-dasharray="4 3" />
          <line x1="562" y1="180" x2="678" y2="180" stroke="#94a3b8" stroke-width="1.2" stroke-dasharray="4 3" />
          
          <!-- 孔心点与 序号 ① -->
          <circle cx="620" cy="180" r="4.5" fill="#d97706" />
          <g transform="translate(585, 195)">
            <circle cx="0" cy="0" r="10" fill="#d97706" stroke="#ffffff" stroke-width="1.5" />
            <text cx="0" cy="3.5" fill="#ffffff" font-family="var(--font-mono)" font-size="9.5" font-weight="900" text-anchor="middle">1</text>
          </g>

          <!-- 咬合边缘点与 序号 ② -->
          <circle cx="659" cy="141" r="5.5" fill="#059669" stroke="#ffffff" stroke-width="2" />
          <g transform="translate(685, 135)">
            <circle cx="0" cy="0" r="10" fill="#059669" stroke="#ffffff" stroke-width="1.5" />
            <text cx="0" cy="3.5" fill="#ffffff" font-family="var(--font-mono)" font-size="9.5" font-weight="900" text-anchor="middle">2</text>
          </g>

          <!-- 序号 ③：ΔL 半径扣减线段 -->
          <line x1="620" y1="180" x2="659" y2="141" stroke="#e11d48" stroke-width="2.5" marker-start="url(#arrow-rose)" marker-end="url(#arrow-rose)" />
          <rect x="545" y="255" width="150" height="26" rx="6" fill="#fff1f2" stroke="#e11d48" stroke-width="1.2" />
          <text x="620" y="272" fill="#e11d48" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">
            ΔL = -dh / 2
          </text>
          <g transform="translate(620, 235)">
            <circle cx="0" cy="0" r="11" fill="#e11d48" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">3</text>
          </g>
        </svg>

        <!-- 结构化技术规范与序号索引 (Legend & Annotation Grid) -->
        <div class="schematic-legend-panel">
          <div class="legend-header">
            <span class="legend-title">图纸技术规范与序号索引 (ENGINEERING SPECIFICATION & INDEX)</span>
            <span class="legend-badge">ISO 128 / HOLE DEDUCTION</span>
          </div>
          <div class="legend-grid">
            <div class="legend-item">
              <div class="legend-badge-num num-amber">1</div>
              <div class="legend-content">
                <div class="legend-item-title">PCD 理论孔心 (Theoretical Hole Center)</div>
                <div class="legend-item-desc">经典纯几何公式默认将孔的几何正中心作为辐条起点，虚增了半个孔径的距离。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-emerald">2</div>
              <div class="legend-content">
                <div class="legend-item-title">弯头内缘受力咬合面 (Contact Edge)</div>
                <div class="legend-item-desc">轮圈张力拉紧下，弯头内弧死死咬在法兰孔壁朝向轮圈侧的边缘倒角处，此为物理受力真实起点。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-rose">3</div>
              <div class="legend-content">
                <div class="legend-item-title">孔半径扣减量 (ΔL = -dh / 2 = -1.25mm)</div>
                <div class="legend-item-desc">标准辐条尺是从弯头内侧起量。孔心到孔壁相差整整一个孔半径（2.5÷2），必须全量扣除 1.25mm！</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-rose">4</div>
              <div class="legend-content">
                <div class="legend-item-title">下方脱空间隙 (Clearance Gap)</div>
                <div class="legend-item-desc">孔径（2.5mm）与条身（2.0mm）之间的装配间隙，受力后在下方脱空，与下料扣减量无直接换算关系。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-slate">5</div>
              <div class="legend-content">
                <div class="legend-item-title">法兰孔壁厚度 (Flange Thickness t ≈ 3.2mm)</div>
                <div class="legend-item-desc">花鼓法兰盘的实体金属壁厚，影响弯头嵌入后的倒角贴合度与受力疲劳寿命。</div>
              </div>
            </div>
          </div>
        </div>

        <div class="diagram-info-bar">
          <div class="diagram-info-text">
            <strong>物理场景 ③：法兰辐条孔径（$d_h$）内缘切点扣减力学原理</strong>
            <p>揭示经典几何公式算到孔心序号 ①，真实咬合在孔边缘序号 ②，导致误差恒为孔半径序号 ③。</p>
          </div>
          <span class="header-badge" style="background: rgba(225, 29, 72, 0.08); color: #be123c; border-color: rgba(225, 29, 72, 0.25);">实测扣减: -1.25 MM</span>
        </div>
      </div>

      <!-- ==============================================
           图解 04: 直拉切线槽位 vs 弯头同心圆 (Straight Pull)
           ============================================== -->
      <div class="diagram-panel" id="diagram-sp" :class="{ active: activeDiagram === 'sp' }">
        <svg viewBox="0 0 760 380" preserveAspectRatio="xMidYMid meet">
          <!-- 标题栏 -->
          <text x="380" y="24" fill="#0f172a" font-size="12" font-weight="900" text-anchor="middle" letter-spacing="0.08em">
            04: STRAIGHT PULL TANGENT GEOMETRY VS J-BEND
          </text>
          <line x1="380" y1="35" x2="380" y2="365" stroke="rgba(15, 23, 42, 0.12)" stroke-dasharray="6 4" />
          
          <!-- 左侧：J-Bend 弯头圆盘花鼓 -->
          <circle cx="190" cy="200" r="90" fill="none" stroke="#cbd5e1" stroke-width="2" stroke-dasharray="4 4" />
          <circle cx="190" cy="200" r="14" fill="#f1f5f9" stroke="#475569" stroke-width="2" />
          <circle cx="190" cy="110" r="6" fill="#334155" />
          <circle cx="268" cy="155" r="6" fill="#334155" />
          <circle cx="268" cy="245" r="6" fill="#334155" />
          <circle cx="190" cy="290" r="6" fill="#334155" />
          <circle cx="112" cy="245" r="6" fill="#334155" />
          <circle cx="112" cy="155" r="6" fill="#334155" />
          <line x1="190" y1="110" x2="340" y2="60" stroke="#475569" stroke-width="2.5" />
          <line x1="268" y1="155" x2="340" y2="260" stroke="#475569" stroke-width="2.5" />
          
          <!-- 序号 ①：J-Bend 极坐标孔 -->
          <g transform="translate(190, 80)">
            <circle cx="0" cy="0" r="11" fill="#475569" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">1</text>
            <text x="18" y="4" fill="#475569" font-family="var(--font-mono)" font-size="10" font-weight="800">J-BEND</text>
          </g>

          <!-- 右侧：Straight Pull 直拉切线预制槽 -->
          <circle cx="570" cy="200" r="65" fill="#f1f5f9" stroke="#475569" stroke-width="2" />
          <circle cx="570" cy="200" r="14" fill="#ffffff" stroke="#64748b" stroke-width="2" />
          <rect x="550" y="125" width="40" height="18" rx="6" fill="#e2e8f0" stroke="#475569" stroke-width="1.8" transform="rotate(-25 570 135)" />
          <circle cx="560" cy="132" r="5" fill="#0f172a" />
          
          <!-- 切线偏距与出条线 -->
          <line x1="570" y1="200" x2="570" y2="132" stroke="#d97706" stroke-dasharray="4 3" stroke-width="1.5" />
          <line x1="570" y1="132" x2="740" y2="132" stroke="#0f172a" stroke-width="3" stroke-linecap="round" />
          <circle cx="570" cy="132" r="3" fill="#d97706" />

          <!-- 序号 ②：切线槽位 -->
          <g transform="translate(530, 115)">
            <circle cx="0" cy="0" r="11" fill="#0f172a" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">2</text>
          </g>

          <!-- 序号 ③：切线偏距 r_tangent -->
          <g transform="translate(535, 170)">
            <circle cx="0" cy="0" r="11" fill="#d97706" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">3</text>
            <text x="16" y="4" fill="#d97706" font-family="var(--font-mono)" font-size="9.5" font-weight="800">r_tangent</text>
          </g>

          <!-- 序号 ④：直拉出条矢量 -->
          <g transform="translate(680, 115)">
            <circle cx="0" cy="0" r="11" fill="#059669" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">4</text>
          </g>
        </svg>

        <!-- 结构化技术规范与序号索引 (Legend & Annotation Grid) -->
        <div class="schematic-legend-panel">
          <div class="legend-header">
            <span class="legend-title">图纸技术规范与序号索引 (ENGINEERING SPECIFICATION & INDEX)</span>
            <span class="legend-badge">ISO 128 / STRAIGHT PULL</span>
          </div>
          <div class="legend-grid">
            <div class="legend-item">
              <div class="legend-badge-num num-slate">1</div>
              <div class="legend-content">
                <div class="legend-item-title">传统 J-Bend 圆盘极坐标孔 (Polar Holes)</div>
                <div class="legend-item-desc">传统花鼓圆盘孔，出条角度受交叉数与极坐标三角函数约束，弯头承受集中交变弯折应力。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-slate">2</div>
              <div class="legend-content">
                <div class="legend-item-title">直拉花鼓切线预制槽位 (Tangential Slot)</div>
                <div class="legend-item-desc">花鼓壳体在铸造/CNC时直接加工的切向定位槽，省去脆弱弯头，使辐条直接承受纯拉力。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-amber">3</div>
              <div class="legend-content">
                <div class="legend-item-title">切线偏距基准线 (r_tangent = 18.5mm)</div>
                <div class="legend-item-desc">出条切线至轴心的固定垂直距离，由花鼓厂商模具决定，决定了三维切向空间三角形态。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-emerald">4</div>
              <div class="legend-content">
                <div class="legend-item-title">直拉出条矢量 (Straight Pull Ray)</div>
                <div class="legend-item-desc">辐条沿槽位直接指向车圈孔，必须使用三维切线立体空间几何求解，严禁套用 J-Bend 圆盘公式。</div>
              </div>
            </div>
          </div>
        </div>

        <div class="diagram-info-bar">
          <div class="diagram-info-text">
            <strong>物理场景 ④：直拉花鼓（Straight Pull）切线槽几何独立算法</strong>
            <p>对比传统极坐标孔序号 ① 与直拉切线槽序号 ②，依切线偏距序号 ③ 进行三维矢量空间求解。</p>
          </div>
          <span class="header-badge" style="background: rgba(15, 23, 42, 0.05); color: #0f172a; border-color: rgba(15, 23, 42, 0.15);">直拉模型已激活</span>
        </div>
      </div>

      <!-- ==============================================
           图解 05: 高张力弹性拉伸量 (Spoke Stretch)
           ============================================== -->
      <div class="diagram-panel" id="diagram-stretch" :class="{ active: activeDiagram === 'stretch' }">
        <svg viewBox="0 0 760 380" preserveAspectRatio="xMidYMid meet">
          <defs>
            <linearGradient id="grad-stretch-steel-p" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#cbd5e1" />
              <stop offset="30%" stop-color="#f8fafc" />
              <stop offset="50%" stop-color="#ffffff" />
              <stop offset="70%" stop-color="#cbd5e1" />
              <stop offset="100%" stop-color="#94a3b8" />
            </linearGradient>
            <linearGradient id="grad-stretch-tension-p" x1="0" y1="0" x2="1" y2="0">
              <stop offset="0%" stop-color="#475569" />
              <stop offset="60%" stop-color="#fb7185" />
              <stop offset="100%" stop-color="#e11d48" />
            </linearGradient>
          </defs>

          <!-- 标题栏 -->
          <text x="380" y="24" fill="#0f172a" font-size="12" font-weight="900" text-anchor="middle" letter-spacing="0.08em">
            05: HIGH-TENSION SPOKE ELASTIC STRETCH (HOOKE'S LAW: ΔL = F·L / E·A)
          </text>

          <!-- 序号 ①：花鼓固定端基准 -->
          <line x1="60" y1="40" x2="60" y2="350" stroke="#64748b" stroke-width="1.5" stroke-dasharray="4 3" />
          <g transform="translate(60, 48)">
            <circle cx="0" cy="0" r="11" fill="#475569" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">1</text>
          </g>

          <!-- 序号 ②：车圈条帽承托座物理基准 -->
          <line x1="480" y1="40" x2="480" y2="350" stroke="#059669" stroke-width="2" stroke-dasharray="6 3" />
          <g transform="translate(480, 48)">
            <circle cx="0" cy="0" r="11" fill="#059669" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">2</text>
          </g>

          <!-- 状态 A：0 N 松弛状态 (序号 ③) -->
          <g transform="translate(5, 75)">
            <g transform="translate(30, 20)">
              <circle cx="0" cy="0" r="11" fill="#475569" stroke="#ffffff" stroke-width="1.8" />
              <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">3</text>
              <text x="16" y="4" fill="#475569" font-family="var(--font-mono)" font-size="9" font-weight="800">0 N</text>
            </g>
            <rect x="48" y="14" width="12" height="18" fill="#e2e8f0" stroke="#64748b" stroke-width="1" />
            <path d="M 45 17 C 41 19 41 27 45 29 Z" fill="#64748b" />
            <rect x="60" y="20" width="400" height="6" rx="3" fill="#64748b" />
            <rect x="460" y="19" width="30" height="8" fill="#cbd5e1" stroke="#475569" stroke-dasharray="2 1" />
            <path d="M 480 12 L 520 12 L 520 16 L 495 16 L 495 30 L 520 30 L 520 34 L 480 34 Z" fill="#fef3c7" stroke="#d97706" stroke-width="1.4" />
            <rect x="510" y="19" width="8" height="8" fill="#ffffff" stroke="#b45309" stroke-width="1" />
          </g>

          <!-- 状态 B：1200 N 未扣减拉伸量 (序号 ④) -->
          <g transform="translate(5, 160)">
            <g transform="translate(30, 20)">
              <circle cx="0" cy="0" r="11" fill="#e11d48" stroke="#ffffff" stroke-width="1.8" />
              <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">4</text>
              <text x="16" y="4" fill="#e11d48" font-family="var(--font-mono)" font-size="9" font-weight="800">1200 N (OVER)</text>
            </g>
            <rect x="48" y="14" width="12" height="18" fill="#e2e8f0" stroke="#64748b" stroke-width="1" />
            <path d="M 45 17 C 41 19 41 27 45 29 Z" fill="#64748b" />
            <rect x="60" y="21" width="418" height="4.5" rx="2" fill="url(#grad-stretch-tension-p)" />
            <rect x="478" y="20" width="48" height="6.5" fill="#e11d48" />
            <path d="M 480 12 L 520 12 L 520 16 L 495 16 L 495 30 L 520 30 L 520 34 L 480 34 Z" fill="#fee2e2" stroke="#e11d48" stroke-width="1.8" />
            <rect x="510" y="19" width="8" height="8" fill="#ffffff" stroke="#e11d48" stroke-width="1" />
            <line x1="518" y1="23" x2="533" y2="23" stroke="#e11d48" stroke-width="3" />
            <circle cx="533" cy="23" r="3.5" fill="#e11d48" />
          </g>

          <!-- 状态 C：1200 N 提前扣除拉伸量 (序号 ⑤) -->
          <g transform="translate(5, 245)">
            <g transform="translate(30, 20)">
              <circle cx="0" cy="0" r="11" fill="#059669" stroke="#ffffff" stroke-width="1.8" />
              <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">5</text>
              <text x="16" y="4" fill="#059669" font-family="var(--font-mono)" font-size="9" font-weight="800">1200 N (FLUSH)</text>
            </g>
            <rect x="48" y="14" width="12" height="18" fill="#e2e8f0" stroke="#64748b" stroke-width="1" />
            <path d="M 45 17 C 41 19 41 27 45 29 Z" fill="#64748b" />
            <rect x="60" y="21" width="418" height="4.5" rx="2" fill="url(#grad-stretch-steel-p)" stroke="#059669" stroke-width="0.8" />
            <rect x="478" y="20" width="34" height="6.5" fill="#059669" />
            <path d="M 480 12 L 520 12 L 520 16 L 495 16 L 495 30 L 520 30 L 520 34 L 480 34 Z" fill="#ecfdf5" stroke="#059669" stroke-width="1.8" />
            <rect x="510" y="19" width="8" height="8" fill="#ffffff" stroke="#059669" stroke-width="1" />
            <line x1="512" y1="8" x2="512" y2="38" stroke="#059669" stroke-width="1.5" stroke-dasharray="2 2" />
          </g>
        </svg>

        <!-- 结构化技术规范与序号索引 (Legend & Annotation Grid) -->
        <div class="schematic-legend-panel">
          <div class="legend-header">
            <span class="legend-title">图纸技术规范与序号索引 (ENGINEERING SPECIFICATION & INDEX)</span>
            <span class="legend-badge">ISO 128 / HOOKE'S LAW</span>
          </div>
          <div class="legend-grid">
            <div class="legend-item">
              <div class="legend-badge-num num-slate">1</div>
              <div class="legend-content">
                <div class="legend-item-title">花鼓固定端基准 (Hub Anchor Datum)</div>
                <div class="legend-item-desc">辐条头部（弯头蘑菇头或直拉圆柱头）卡在花鼓上的绝对固定零点。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-emerald">2</div>
              <div class="legend-content">
                <div class="legend-item-title">车圈条帽承托座基准 (Rim Bed Datum)</div>
                <div class="legend-item-desc">车圈床身接触平面，空间位置固定不动，作为拉伸量计算的对径端点基准。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-slate">3</div>
              <div class="legend-content">
                <div class="legend-item-title">状态 A：0 N 松弛下料状态 (Relaxed Cutting Length)</div>
                <div class="legend-item-desc">辐条未施加任何拉力时的天然原长 L₀，螺牙旋入条帽深度约 50%，无张力形变。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-rose">4</div>
              <div class="legend-content">
                <div class="legend-item-title">状态 B：未扣除伸长量顶死 (Bottom Out Failure)</div>
                <div class="legend-item-desc">在 1200N 张力下辐条弹性伸长约 1.2mm，螺纹拧到底顶入胎垫孔，导致整组轮无法上紧报废。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-emerald">5</div>
              <div class="legend-content">
                <div class="legend-item-title">状态 C：提前扣除拉伸满牙平齐 (Compensated Flush Assembly)</div>
                <div class="legend-item-desc">下料前提前扣减伸长量（-1.18mm），张力拉满后螺牙刚好平齐一字槽底，达到 100% 满牙受力。</div>
              </div>
            </div>
          </div>
        </div>

        <div class="diagram-info-bar">
          <div class="diagram-info-text">
            <strong>物理场景 ⑤：高张力弹性伸长量（胡克定律: $\Delta L = \frac{F \cdot L}{E \cdot A}$）</strong>
            <p>对比松弛态序号 ③、未扣减顶死态序号 ④ 与提前扣除黄金满牙平齐态序号 ⑤。</p>
          </div>
          <span class="header-badge" style="background: rgba(225, 29, 72, 0.08); color: #be123c; border-color: rgba(225, 29, 72, 0.25);">弹性伸长扣减: -1.18 MM</span>
        </div>
      </div>

      <!-- ==============================================
           图解 06: 轮圈孔位交错偏移 (Alternating Rim Offset)
           ============================================== -->
      <div class="diagram-panel" id="diagram-drill" :class="{ active: activeDiagram === 'drill' }">
        <svg viewBox="0 0 760 380" preserveAspectRatio="xMidYMid meet">
          <!-- 标题栏 -->
          <text x="380" y="24" fill="#0f172a" font-size="12" font-weight="900" text-anchor="middle" letter-spacing="0.08em">
            06: ASYMMETRIC RIM & ALTERNATING DRILLING SCHEMATIC
          </text>

          <!-- 偏心轮圈截面 -->
          <path d="M 120 160 C 120 100 200 70 380 70 C 560 70 640 100 640 160 L 610 240 C 580 320 460 360 380 360 C 300 360 180 320 150 240 Z" fill="#f1f5f9" stroke="#475569" stroke-width="1.8" />
          
          <!-- 序号 ①：轮圈物理对称轴 -->
          <line x1="380" y1="50" x2="380" y2="375" stroke="rgba(15, 23, 42, 0.25)" stroke-dasharray="6 4" stroke-width="1.5" />
          <g transform="translate(380, 42)">
            <circle cx="0" cy="0" r="11" fill="#475569" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">1</text>
          </g>

          <!-- 序号 ②：偏心床面整体偏心距 (+2.5mm) -->
          <line x1="420" y1="50" x2="420" y2="375" stroke="#d97706" stroke-dasharray="4 3" stroke-width="1.5" />
          <g transform="translate(420, 42)">
            <circle cx="0" cy="0" r="11" fill="#d97706" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">2</text>
          </g>

          <!-- 序号 ③：左侧交错孔位 -->
          <circle cx="405" cy="160" r="14" fill="#ffffff" stroke="#475569" stroke-width="2" />
          <circle cx="405" cy="160" r="4" fill="#0f172a" />
          <g transform="translate(365, 160)">
            <circle cx="0" cy="0" r="11" fill="#0f172a" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">3</text>
          </g>

          <!-- 序号 ④：右侧交错孔位 -->
          <circle cx="435" cy="220" r="14" fill="#ffffff" stroke="#475569" stroke-width="2" />
          <circle cx="435" cy="220" r="4" fill="#0f172a" />
          <g transform="translate(475, 220)">
            <circle cx="0" cy="0" r="11" fill="#0f172a" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">4</text>
          </g>

          <!-- 序号 ⑤：交错孔间距尺寸线 -->
          <line x1="405" y1="120" x2="435" y2="120" stroke="#d97706" stroke-width="1.5" stroke-dasharray="3 2" />
          <g transform="translate(420, 110)">
            <circle cx="0" cy="0" r="11" fill="#d97706" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">5</text>
            <text x="16" y="4" fill="#b45309" font-family="var(--font-mono)" font-size="9" font-weight="900">ΔOffset</text>
          </g>

          <!-- 辐条牵引示意 -->
          <line x1="160" y1="340" x2="405" y2="160" stroke="#475569" stroke-width="2" stroke-dasharray="4 2" />
          <line x1="600" y1="340" x2="435" y2="220" stroke="#64748b" stroke-width="2" stroke-dasharray="4 2" />
        </svg>

        <!-- 结构化技术规范与序号索引 (Legend & Annotation Grid) -->
        <div class="schematic-legend-panel">
          <div class="legend-header">
            <span class="legend-title">图纸技术规范与序号索引 (ENGINEERING SPECIFICATION & INDEX)</span>
            <span class="legend-badge">ISO 128 / RIM DRILLING</span>
          </div>
          <div class="legend-grid">
            <div class="legend-item">
              <div class="legend-badge-num num-slate">1</div>
              <div class="legend-content">
                <div class="legend-item-title">轮圈物理几何对称轴 (Rim Symmetry Axis)</div>
                <div class="legend-item-desc">车圈外部轮廓的物理中心线，安装外胎与刹车夹器定位的中心几何基准。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-amber">2</div>
              <div class="legend-content">
                <div class="legend-item-title">床面整体偏心距 (Asymmetric Rim Offset)</div>
                <div class="legend-item-desc">偏心碳圈截面将辐条床整体向非驱动侧偏移（如 +2.5mm），大幅拉开右侧法兰夹角。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-slate">3</div>
              <div class="legend-content">
                <div class="legend-item-title">左侧交错孔位 (-Offset)</div>
                <div class="legend-item-desc">在偏心床面基础上进一步向非驱动侧偏置的钻孔，用于连接左侧法兰，减小出条角度偏折。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-slate">4</div>
              <div class="legend-content">
                <div class="legend-item-title">右侧交错孔位 (+Offset)</div>
                <div class="legend-item-desc">偏向驱动侧的交错孔位，用于连接右侧法兰，改善右侧出条入孔对齐姿态。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-amber">5</div>
              <div class="legend-content">
                <div class="legend-item-title">交错孔横向跨距差 (ΔOffset = 1.50mm)</div>
                <div class="legend-item-desc">相邻左右两孔间的横向偏距差值，对实际有效法兰距产生交错增减补偿。</div>
              </div>
            </div>
          </div>
        </div>

        <div class="diagram-info-bar">
          <div class="diagram-info-text">
            <strong>物理场景 ⑥：偏心轮圈与交错孔位钻孔（Alternating Offset）</strong>
            <p>展示截面对称轴序号 ①、整体偏心距序号 ② 与左右交错钻孔序号 ③、④。</p>
          </div>
          <span class="header-badge" style="background: rgba(15, 23, 42, 0.05); color: #0f172a; border-color: rgba(15, 23, 42, 0.15);">交错差: ±0.75 MM</span>
        </div>
      </div>

      <!-- ==============================================
           图解 07: 编法双条交叠压条物理折线 (Interlacing)
           ============================================== -->
      <div class="diagram-panel" id="diagram-interlace" :class="{ active: activeDiagram === 'interlace' }">
        <svg viewBox="0 0 760 380" preserveAspectRatio="xMidYMid meet">
          <!-- 标题栏 -->
          <text x="380" y="24" fill="#0f172a" font-size="12" font-weight="900" text-anchor="middle" letter-spacing="0.08em">
            07: 2X / 3X CROSS OVER-UNDER INTERLACING SCHEMATIC
          </text>

          <!-- 序号 ①：理想直线轨迹 -->
          <line x1="60" y1="300" x2="700" y2="80" stroke="#94a3b8" stroke-dasharray="5 3" stroke-width="2" />
          <g transform="translate(640, 75)">
            <circle cx="0" cy="0" r="11" fill="#475569" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">1</text>
            <text x="16" y="4" fill="#475569" font-family="var(--font-mono)" font-size="9" font-weight="800">IDEAL</text>
          </g>

          <!-- 压条外层与内层辐条 -->
          <path d="M 60 300 L 340 190 Q 380 160 420 170 L 700 70" fill="none" stroke="#0f172a" stroke-width="7" stroke-linecap="round" />
          <path d="M 80 70 L 340 160 Q 380 200 420 190 L 680 300" fill="none" stroke="#64748b" stroke-width="7" stroke-linecap="round" />

          <!-- 序号 ②：外层辐条弯折 -->
          <g transform="translate(260, 205)">
            <circle cx="0" cy="0" r="11" fill="#0f172a" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">2</text>
          </g>

          <!-- 序号 ③：内层辐条反拱 -->
          <g transform="translate(260, 135)">
            <circle cx="0" cy="0" r="11" fill="#64748b" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">3</text>
          </g>

          <!-- 序号 ④：压条接触咬合点 -->
          <circle cx="380" cy="180" r="32" fill="none" stroke="#e11d48" stroke-width="1.8" stroke-dasharray="4 4" />
          <g transform="translate(380, 235)">
            <circle cx="0" cy="0" r="11" fill="#e11d48" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">4</text>
          </g>

          <!-- 序号 ⑤：拱起位移量 δ -->
          <line x1="380" y1="150" x2="380" y2="210" stroke="#e11d48" stroke-width="1.5" />
          <g transform="translate(425, 175)">
            <circle cx="0" cy="0" r="11" fill="#e11d48" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-family="var(--font-mono)" font-size="11" font-weight="900" text-anchor="middle">5</text>
            <text x="16" y="4" fill="#e11d48" font-family="var(--font-mono)" font-size="9.5" font-weight="900">δ varies</text>
          </g>

          <!-- 计算公式牌 -->
          <rect x="180" y="320" width="400" height="36" rx="8" fill="#f8fafc" stroke="rgba(15, 23, 42, 0.12)" />
          <text x="380" y="342" fill="#059669" font-family="var(--font-mono)" font-size="11.5" font-weight="900" text-anchor="middle">
            ΔL_interlace = √(L1² + δ²) + √(L2² + δ²) - (L1+L2)
          </text>
        </svg>

        <!-- 结构化技术规范与序号索引 (Legend & Annotation Grid) -->
        <div class="schematic-legend-panel">
          <div class="legend-header">
            <span class="legend-title">图纸技术规范与序号索引 (ENGINEERING SPECIFICATION & INDEX)</span>
            <span class="legend-badge">ISO 128 / INTERLACING</span>
          </div>
          <div class="legend-grid">
            <div class="legend-item">
              <div class="legend-badge-num num-slate">1</div>
              <div class="legend-content">
                <div class="legend-item-title">理想纯欧几里得直线 (Uninterlaced Line)</div>
                <div class="legend-item-desc">经典理论计算假设的直线路径，未考虑交叉点压条产生的物理位移。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-slate">2</div>
              <div class="legend-content">
                <div class="legend-item-title">外层辐条压折轨迹 (Over Spoke)</div>
                <div class="legend-item-desc">交叉编法中从外侧越过相邻辐条的条身，装配时手工向内轻微别压。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-slate">3</div>
              <div class="legend-content">
                <div class="legend-item-title">内层辐条反向拱起 (Under Spoke)</div>
                <div class="legend-item-desc">位于下层的辐条被外层别紧，形成对应的微观向上拱弯。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-rose">4</div>
              <div class="legend-content">
                <div class="legend-item-title">压条物理咬合节点 (Interlaced Node)</div>
                <div class="legend-item-desc">两根辐条紧密咬合接触点，能有效传递刹车力矩并消除微小共振。</div>
              </div>
            </div>
            <div class="legend-item">
              <div class="legend-badge-num num-rose">5</div>
              <div class="legend-content">
                <div class="legend-item-title">微观折线拱起位移 (Deflection δ，示意值)</div>
                <div class="legend-item-desc">折线拱起会使辐条物理几何总长增加；具体补偿取决于辐条截面、变径形状和花鼓出条方向，不能用一个固定值代表。</div>
              </div>
            </div>
          </div>
        </div>

        <div class="diagram-info-bar">
          <div class="diagram-info-text">
            <strong>物理场景 ⑦：交叉编法交叠压条（Interlacing）的物理折线补偿</strong>
            <p>展示直线欧氏路径序号 ①、外层下压序号 ②、内层拱起序号 ③ 与压条咬合点序号 ④。</p>
          </div>
          <span class="header-badge" style="background: rgba(5, 150, 105, 0.08); color: #047857; border-color: rgba(5, 150, 105, 0.25);">压条补偿：按实际输入</span>
        </div>
      </div>

    </div>
      </section>
    </div>

  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const activeDiagram = ref('erd')
const isCollapsed = ref(false)

onMounted(() => {
  const mobileQuery = window.matchMedia('(max-width: 767px)')
  const syncCollapseState = () => {
    isCollapsed.value = mobileQuery.matches
  }

  syncCollapseState()
  mobileQuery.addEventListener?.('change', syncCollapseState)

  onBeforeUnmount(() => {
    mobileQuery.removeEventListener?.('change', syncCollapseState)
  })
})
</script>

<style scoped>
    /* ==========================================================================
       ERP UDS 工业设计系统规范 (v1.0) - 国际化工程图纸标准 (ISO / CAD Legend Decoupled)
       ========================================================================== */
    .spoke-physics-diagrams {
      /* 基础淡色工业工程图纸灰阶 (Light Precision Blueprint & Technical Slate) */
      --bg-base: #f8fafc;            /* Slate-50 极清爽明净的工业工程底 */
      --bg-surface: #ffffff;         /* 纯白结构底板 */
      --bg-card: #ffffff;            /* 卡片纯白 */
      --bg-card-hover: #f8fafc;
      --bg-inset: #f1f5f9;           /* Slate-100 浅灰内嵌沉板 */
      --border-line: rgba(15, 23, 42, 0.08);
      --border-dashed: 1px dashed rgba(15, 23, 42, 0.16);
      --border-focus: #0f172a;

      /* 严谨工程语义配色 (Strict ERP UDS Semantics - 浅色高对比版) */
      --accent-primary: #059669;     /* Tanzanite 翡翠精工绿 Emerald-600 */
      --accent-amber: #d97706;       /* 尺寸标注 / PCD 琥珀金 Amber-600 */
      --accent-rose: #e11d48;        /* 临界 / 负向扣减工程红 Rose-600 */
      --accent-emerald: #059669;     /* 正向补偿 / 健康状态绿 */
      --accent-steel: #475569;       /* 机械冷钢中性标尺 Slate-600 */
      
      /* 高可读性工业排版深色色阶 */
      --text-main: #0f172a;          /* Slate-900 锐利深黑高清晰文字 */
      --text-secondary: #334155;     /* Slate-700 次级信息 */
      --text-muted: #64748b;         /* Slate-500 辅助标签 */
      --text-dim: #94a3b8;           /* Slate-400 弱化网格与代码 */
      
      --font-mono: "JetBrains Mono", Consolas, "Courier New", monospace;
      --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
    }

    .spoke-physics-diagrams,
    .spoke-physics-diagrams * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
    }

    .spoke-physics-diagrams {
      background-color: var(--bg-base);
      background-image: 
        linear-gradient(rgba(15, 23, 42, 0.035) 1px, transparent 1px),
        linear-gradient(90deg, rgba(15, 23, 42, 0.035) 1px, transparent 1px);
      background-size: 28px 28px;
      color: var(--text-main);
      font-family: var(--font-sans);
      width: 100%;
      padding: 24px;
      display: flex;
      flex-direction: column;
      gap: 20px;
      max-width: none;
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
      font-family: var(--font-mono);
      font-weight: 800;
      letter-spacing: 0.08em;
    }

    .physics-collapse-toggle {
      display: none;
      width: 100%;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      border: 1px solid rgba(15, 23, 42, 0.12);
      border-radius: 9999px;
      background: #ffffff;
      color: var(--text-main);
      padding: 10px 12px;
      font-family: var(--font-sans);
      font-size: 12px;
      font-weight: 800;
      text-align: left;
      cursor: pointer;
    }

    .physics-collapse-toggle:focus-visible {
      outline: 2px solid #059669;
      outline-offset: 2px;
    }

    .physics-collapse-toggle__state {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      color: #047857;
      font-size: 10px;
      white-space: nowrap;
    }

    .physics-collapse-content {
      display: contents;
    }

    /* 中间图解看板交互区 */
    .schematic-showcase {
      display: flex;
      flex-direction: column;
      gap: 16px;
      min-width: 0;
      width: 100%;
    }

    .schematic-nav {
      display: grid;
      grid-template-columns: repeat(7, 1fr);
      gap: 6px;
      background: var(--bg-surface);
      border: var(--border-dashed);
      border-radius: 16px;
      padding: 6px;
      box-sizing: border-box;
      width: 100%;
      box-shadow: 0 2px 8px rgba(15, 23, 42, 0.03);
    }

    .schematic-tab {
      min-width: 0;
      padding: 6px 4px;
      border: 1px solid transparent;
      background: transparent;
      color: var(--text-muted);
      font-size: 9px;
      font-weight: 800;
      text-transform: uppercase;
      letter-spacing: 0.04em;
      border-radius: 9999px;
      cursor: pointer;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 2px;
      transition: all 0.2s ease;
      white-space: nowrap;
      height: 40px;
      max-height: 40px;
      box-sizing: border-box;
    }

    .schematic-tab span:last-child {
      overflow: hidden;
      text-overflow: ellipsis;
      max-width: 100%;
    }

    .schematic-tab span.num {
      font-family: var(--font-mono);
      font-size: 7.5px;
      padding: 1px 4px;
      background: rgba(15, 23, 42, 0.06);
      color: #334155;
      border-radius: 9999px;
    }

    .schematic-tab:hover {
      color: var(--text-main);
      background: rgba(15, 23, 42, 0.04);
    }

    .schematic-tab.active {
      background: #0f172a;
      color: #ffffff;
      font-weight: 900;
      border-color: #0f172a;
      box-shadow: 0 3px 10px rgba(15, 23, 42, 0.15);
    }

    .schematic-tab.active span.num {
      background: #ffffff;
      color: #0f172a;
      font-weight: 900;
    }

    @media (max-width: 900px) {
      .schematic-nav {
        grid-template-columns: repeat(4, 1fr);
      }
    }

    /* 矢量画布视口 (CAD Blueprint Viewport) */
    .canvas-viewport {
      background: #ffffff;
      border: var(--border-dashed);
      border-radius: 24px;
      position: relative;
      overflow: hidden;
      padding: 24px;
      box-shadow: inset 0 0 30px rgba(15, 23, 42, 0.02), 0 10px 25px -5px rgba(15, 23, 42, 0.04);
    }

    .diagram-panel {
      display: none;
      width: 100%;
      grid-template-columns: minmax(0, 1.15fr) minmax(300px, 0.85fr);
      align-items: start;
      gap: 16px;
      animation: fadeIn 0.25s ease forwards;
    }

    .diagram-panel.active {
      display: grid;
    }

    @keyframes fadeIn {
      from { opacity: 0; transform: scale(0.99); }
      to { opacity: 1; transform: scale(1); }
    }

    .diagram-panel svg {
      grid-column: 1;
      grid-row: 1;
      width: 100%;
      height: auto;
      max-height: 380px;
    }

    .diagram-panel .schematic-legend-panel {
      grid-column: 2;
      grid-row: 1;
      min-width: 0;
    }

    .diagram-panel .legend-grid {
      grid-template-columns: minmax(0, 1fr);
    }

    .diagram-panel .diagram-info-bar {
      grid-column: 1 / -1;
      grid-row: 2;
    }

    /* 结构化技术规范与序号索引表 (Legend & Annotation Grid) */
    .schematic-legend-panel {
      background: #ffffff;
      border: 1px solid rgba(15, 23, 42, 0.08);
      border-radius: 16px;
      padding: 16px 20px;
      display: flex;
      flex-direction: column;
      gap: 12px;
      box-shadow: 0 2px 8px rgba(15, 23, 42, 0.02);
    }

    .legend-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      border-bottom: 1px dashed rgba(15, 23, 42, 0.1);
      padding-bottom: 8px;
    }

    .legend-title {
      font-size: 10.5px;
      font-weight: 900;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      color: var(--text-main);
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .legend-badge {
      font-family: var(--font-mono);
      font-size: 8px;
      font-weight: 800;
      padding: 2px 7px;
      border-radius: 4px;
      background: rgba(15, 23, 42, 0.05);
      color: var(--text-dim);
    }

    .legend-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
      gap: 10px 16px;
    }

    .legend-item {
      display: flex;
      align-items: flex-start;
      gap: 10px;
      background: #f8fafc;
      border: 1px solid rgba(15, 23, 42, 0.05);
      border-radius: 10px;
      padding: 10px 12px;
      transition: all 0.15s ease;
    }

    .legend-item:hover {
      background: #f1f5f9;
      border-color: rgba(15, 23, 42, 0.12);
    }

    .legend-badge-num {
      width: 22px;
      height: 22px;
      border-radius: 9999px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-family: var(--font-mono);
      font-size: 10.5px;
      font-weight: 900;
      flex-shrink: 0;
      margin-top: 1px;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    }

    .num-slate {
      background: #0f172a;
      color: #ffffff;
    }

    .num-emerald {
      background: #059669;
      color: #ffffff;
    }

    .num-amber {
      background: #d97706;
      color: #ffffff;
    }

    .num-rose {
      background: #e11d48;
      color: #ffffff;
    }

    .legend-content {
      display: flex;
      flex-direction: column;
      gap: 3px;
      min-width: 0;
    }

    .legend-item-title {
      font-size: 10.5px;
      font-weight: 800;
      color: var(--text-main);
      display: flex;
      align-items: center;
      gap: 6px;
      line-height: 1.3;
    }

    .legend-item-desc {
      font-size: 9.5px;
      color: var(--text-secondary);
      line-height: 1.45;
    }

    /* 底部物理总结条 */
    .diagram-info-bar {
      background: var(--bg-surface);
      border: 1px solid rgba(15, 23, 42, 0.08);
      border-radius: 14px;
      padding: 12px 18px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      box-shadow: 0 2px 6px rgba(15, 23, 42, 0.02);
    }

    .diagram-info-text strong {
      color: var(--text-main);
      font-size: 11px;
      font-weight: 800;
    }

    .diagram-info-text p {
      font-size: 9.5px;
      color: var(--text-secondary);
      margin-top: 2px;
    }

/* The reference document is embedded as a full-width card inside the
 * calculator shell, so its internal canvas remains independent of page chrome. */
.spoke-physics-diagrams {
  width: 100%;
  max-width: none;
  margin: 0 auto;
  border: 1px dashed rgba(15, 23, 42, 0.16);
  border-radius: 24px;
  overflow: hidden;
  isolation: isolate;
}

@media (max-width: 767px) {
  .spoke-physics-diagrams {
    padding: 16px;
    gap: 14px;
  }

  .spoke-physics-diagrams .physics-collapse-toggle {
    display: flex;
  }

  .spoke-physics-diagrams .physics-collapse-content {
    display: block;
  }

  .spoke-physics-diagrams .canvas-viewport {
    padding: 12px;
    border-radius: 18px;
  }

  .spoke-physics-diagrams .diagram-panel {
    grid-template-columns: 1fr;
  }

  .spoke-physics-diagrams .diagram-panel svg,
  .spoke-physics-diagrams .diagram-panel .schematic-legend-panel,
  .spoke-physics-diagrams .diagram-panel .diagram-info-bar {
    grid-column: 1;
    grid-row: auto;
  }

  .spoke-physics-diagrams .diagram-info-bar {
    align-items: flex-start;
    flex-direction: column;
    gap: 8px;
  }

  .spoke-physics-diagrams .schematic-legend-panel {
    padding: 12px;
  }

  .spoke-physics-diagrams .legend-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 6px;
  }
}
</style>
