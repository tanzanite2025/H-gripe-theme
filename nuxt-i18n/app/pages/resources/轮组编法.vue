<template>
<div class="spoke-lacing-page">
<!-- ================= 顶部工程页眉 ================= -->
  <header class="blueprint-header">
    <div class="header-title-group">
      <h1>
        <span>辐条孔数与编法交叉力学工程蓝图</span>
        <span class="header-badge badge-primary">SPOKE LACING V1.0-ENGINEERING</span>
        <span class="header-badge">ISO 128 CAD SPEC</span>
      </h1>
      <div class="header-subtitle">
        FULL TOPOLOGY WHEEL MATRIX • DUAL-FLANGE SYNCHRONIZED MAPPING • INTERFERENCE CLEARANCE ENGINE
      </div>
    </div>
    <div class="header-controls">
      <span class="header-badge" style="background: rgba(5, 150, 105, 0.06); color: #047857; border-color: rgba(5, 150, 105, 0.2);">
        STANDALONE PROTOTYPE • 纯设计预览文件
      </span>
    </div>
  </header>

  <!-- ================= 核心工作台主网格 ================= -->
  <div class="blueprint-grid" aria-label="轮组编法工程工作台">

    <!-- ========== 左侧列：参数选择器与图层控制 ========== -->
    <section class="uds-card controls-column">
      <div class="card-header">
        <h2 class="card-title">
          <span>01. 几何参数控制台</span>
        </h2>
        <span class="card-tag">INPUT MATRIX</span>
      </div>

      <!-- 1. 轮组总孔数选择 (Hole Count) -->
      <div>
        <div class="selector-group-label">
          <span>轮组总孔数 (Hole Count)</span>
          <span id="current-hole-label" style="font-family: var(--tz-font-ui); color: var(--accent-primary);">24H (1:1)</span>
        </div>
        <div class="hole-pill-grid">
          <button type="button" class="uds-pill-btn" data-holes="16" @click="setHoleCount(16)">
            <span>16H</span>
            <span class="sub-text">超轻前轮</span>
          </button>
          <button type="button" class="uds-pill-btn" data-holes="20" @click="setHoleCount(20)">
            <span>20H</span>
            <span class="sub-text">公路圈刹</span>
          </button>
          <button type="button" class="uds-pill-btn" data-holes="21" @click="setHoleCount(21)">
            <span>21H</span>
            <span class="sub-text">G3 2:1</span>
          </button>
          <button type="button" class="uds-pill-btn active" data-holes="24" @click="setHoleCount(24)">
            <span>24H</span>
            <span class="sub-text">现代碟刹</span>
          </button>
          <button type="button" class="uds-pill-btn" data-holes="28" @click="setHoleCount(28)">
            <span>28H</span>
            <span class="sub-text">全地形GR</span>
          </button>
          <button type="button" class="uds-pill-btn" data-holes="32" @click="setHoleCount(32)">
            <span>32H</span>
            <span class="sub-text">经典越野</span>
          </button>
          <button type="button" class="uds-pill-btn" data-holes="36" @click="setHoleCount(36)">
            <span>36H</span>
            <span class="sub-text">重载旅行</span>
          </button>
          <button type="button" class="uds-pill-btn" data-holes="24_2to1" @click="setHoleCount('24_2to1')">
            <span>24H (2:1)</span>
            <span class="sub-text">异对称16/8</span>
          </button>
        </div>
      </div>

      <!-- 2. 编法交叉数选择 (Cross Pattern) -->
      <div>
        <div class="selector-group-label">
          <span>交叉编法 (Cross Pattern)</span>
          <span id="current-cross-label" style="font-family: var(--tz-font-ui); color: var(--accent-primary);">2X CROSS</span>
        </div>
        <div class="cross-pill-grid" id="cross-btn-container">
          <button type="button" class="uds-pill-btn" data-cross="0" @click="setCrossCount(0)">
            <span>0X</span>
            <span class="sub-text">Radial 直拉</span>
          </button>
          <button type="button" class="uds-pill-btn" data-cross="1" @click="setCrossCount(1)">
            <span>1X</span>
            <span class="sub-text">一交叉</span>
          </button>
          <button type="button" class="uds-pill-btn active" data-cross="2" @click="setCrossCount(2)">
            <span>2X</span>
            <span class="sub-text">二交叉</span>
          </button>
          <button type="button" class="uds-pill-btn" data-cross="3" @click="setCrossCount(3)">
            <span>3X</span>
            <span class="sub-text">三交叉</span>
          </button>
          <button type="button" class="uds-pill-btn" data-cross="4" @click="setCrossCount(4)">
            <span>4X</span>
            <span class="sub-text">四交叉</span>
          </button>
        </div>
      </div>

      <!-- 3. 轮侧视图模式切换 (View Mode) -->
      <div>
        <div class="selector-group-label">
          <span>法兰透视模式 (Flange View)</span>
        </div>
        <div class="view-mode-grid">
          <button type="button" class="view-pill-btn" id="btn-view-drive" @click="setViewMode('drive')">
            右侧 (驱动侧)
          </button>
          <button type="button" class="view-pill-btn active" id="btn-view-both" @click="setViewMode('both')">
            整轮双侧透视
          </button>
          <button type="button" class="view-pill-btn" id="btn-view-nondrive" @click="setViewMode('nondrive')">
            左侧 (非驱动)
          </button>
        </div>
      </div>

      <!-- 4. 图层显示开关 -->
      <div class="layer-toggle-group">
        <span class="selector-group-label" style="margin-bottom: 4px;">图层显隐辅助</span>
        
        <label class="layer-item">
          <span><span class="layer-indicator" style="background: var(--accent-primary);"></span>牵引条 (Leading Spokes)</span>
          <input type="checkbox" id="toggle-leading" checked @change="updateCanvasLayers">
        </label>
        
        <label class="layer-item">
          <span><span class="layer-indicator" style="background: var(--accent-amber);"></span>推力条 (Trailing Spokes)</span>
          <input type="checkbox" id="toggle-trailing" checked @change="updateCanvasLayers">
        </label>

        <label class="layer-item">
          <span><span class="layer-indicator" style="background: var(--accent-sky);"></span>左侧/非驱动条 (Non-Drive Spokes)</span>
          <input type="checkbox" id="toggle-nondrive" checked @change="updateCanvasLayers">
        </label>

        <label class="layer-item">
          <span><span class="layer-indicator" style="background: var(--accent-rose);"></span>物理交叉咬合点 (Cross Nodes)</span>
          <input type="checkbox" id="toggle-nodes" checked @change="updateCanvasLayers">
        </label>

      </div>

    </section>

    <!-- ========== 中央列：高精交互 CAD 矢量轮组画布 ========== -->
    <section class="uds-card canvas-card">
      <div class="canvas-toolbar">
        <div class="card-title">
          <span>02. 轮组几何矢量蓝图视口</span>
          <span class="card-tag" id="canvas-status-tag">FULL TOPOLOGY</span>
        </div>
        <div class="canvas-legend-inline">
          <span style="color: var(--accent-primary);">■ 右侧牵引</span>
          <span style="color: var(--accent-amber);">■ 右侧推力</span>
          <span style="color: var(--accent-sky);">■ 左侧辐条</span>
          <span style="color: var(--accent-rose);">● 物理交叉</span>
          <span style="color: var(--accent-indigo);">◆ 气门嘴</span>
        </div>
      </div>

      <!-- 矢量画布 -->
      <div class="svg-viewport" id="viewport-container">
        <svg id="spoke-lacing-svg" viewBox="-300 -300 600 600" preserveAspectRatio="xMidYMid meet">
          <!-- 动态渲染 SVG 元素 -->
        </svg>
      </div>
    </section>

    <!-- ========== 右侧列：实时力学遥测与干涉评估 ========== -->
    <section class="telemetry-column">

      <!-- 实时工程指标卡片 -->
      <div class="uds-card">
        <div class="card-header">
          <h2 class="card-title">
            <span>03. 实时工程力学遥测</span>
          </h2>
          <span class="card-tag">MECHANICS</span>
        </div>

        <!-- 指标 1：出条切线角 α -->
        <div class="telemetry-metric-tile">
          <div class="metric-top">
            <span class="metric-label">法兰出条切线角 (Tangent Angle α)</span>
            <span class="metric-unit">理想: 75°~90°</span>
          </div>
          <div class="metric-value">
            <span id="metric-tangent-angle">64.2</span>
            <span class="metric-unit">DEG (°)</span>
          </div>
          <div class="metric-bar">
            <div id="bar-tangent-angle" class="metric-bar-fill" style="width: 71%; background-color: var(--accent-primary);"></div>
          </div>
          <p class="metric-desc">辐条引出法兰孔时的切向夹角；越接近 90° 纯切线，抗扭刚度越高。</p>
        </div>

        <!-- 指标 2：扭转力矩传递效率 (sin α) -->
        <div class="telemetry-metric-tile">
          <div class="metric-top">
            <span class="metric-label">扭转力矩传递效率 (Torque Ratio)</span>
            <span class="metric-unit">η = sin(α)</span>
          </div>
          <div class="metric-value">
            <span id="metric-torque-efficiency">90.0</span>
            <span class="metric-unit">%</span>
          </div>
          <div class="metric-bar">
            <div id="bar-torque-efficiency" class="metric-bar-fill" style="width: 90%; background-color: var(--accent-primary);"></div>
          </div>
          <p class="metric-desc">加速踏力和碟刹制动时的响应灵敏度；0X 直拉时为 0%（严禁用于碟刹侧）。</p>
        </div>

        <!-- 指标 3：侧向支撑角刚度指数 (cos α) -->
        <div class="telemetry-metric-tile">
          <div class="metric-top">
            <span class="metric-label">侧向抗摇车指数 (Lateral Bracing)</span>
            <span class="metric-unit">γ = cos(α)</span>
          </div>
          <div class="metric-value">
            <span id="metric-lateral-index">43.5</span>
            <span class="metric-unit">%</span>
          </div>
          <div class="metric-bar">
            <div id="bar-lateral-index" class="metric-bar-fill" style="width: 43.5%; background-color: var(--accent-steel);"></div>
          </div>
          <p class="metric-desc">侧向摇车抗倾覆刚度分量；0X 放射时达到理论最大 100%。</p>
        </div>

        <!-- 指标 4：单侧物理交叉点数量 -->
        <div class="telemetry-metric-tile">
          <div class="metric-top">
            <span class="metric-label">当前视角物理交叉节点总数</span>
            <span class="metric-unit">Nodes Count</span>
          </div>
          <div class="metric-value">
            <span id="metric-node-count">12</span>
            <span class="metric-unit">NODES</span>
          </div>
          <p class="metric-desc">两根辐条紧密咬合接触点，能够有效抑制辐条高频微共振并消除异响。</p>
        </div>
      </div>

      <!-- 法兰孔干涉与安全性综合评估 -->
      <div class="uds-card">
        <div class="card-header">
          <h2 class="card-title">
            <span>04. 法兰孔干涉与装配裁定</span>
          </h2>
          <span class="card-tag">SAFETY AUDIT</span>
        </div>

        <div id="clearance-status-box" class="clearance-status-card status-healthy">
          <div class="status-header-line">
            <span id="status-pulse-dot" class="pulse-dot" style="display: none;"></span>
            <span id="status-title-text">HEALTHY: 编法结构健康</span>
          </div>
          <p id="status-detail-text" class="status-detail-text">
            24孔搭配 2X 交叉编法为现代公路碟刹车轮的黄金标准组合。
          </p>
        </div>

        <div style="background: var(--bg-inset); border-radius: 14px; padding: 12px; font-size: 9.5px; color: var(--text-secondary); line-height: 1.45;">
          <strong style="color: var(--text-main); display: block; margin-bottom: 4px;">技师装配实战建议：</strong>
          <span id="builder-tip-text">
            推荐用于公路碟刹轮组前轮驱动侧/刹车侧及后轮两侧。
          </span>
        </div>
      </div>

    </section>

  </div>

  <!-- ================= 底部 ISO / CAD 技术规范说明对照表 ================= -->
  <footer class="cad-legend-section">
    <div class="card-header" style="padding-bottom: 8px;">
      <h3 class="card-title">
        <span>05. 自行车轮组编制工程图纸技术规范与准则索引 (ISO 128 / WHEELBUILDING STANDARDS)</span>
      </h3>
      <span class="header-badge">ENGINEERING SPECIFICATION</span>
    </div>

    <div class="legend-table-grid">
      <div class="legend-box">
        <div class="legend-box-title">
          <span class="legend-box-num">01</span>
          <span>放射编法 (0X Radial)</span>
        </div>
        <div class="legend-box-body">
          辐条沿花鼓轴心法线呈 0° 辐射直接连至车圈。侧向刚度最佳，但抗扭刚度为零，<strong>严禁用于碟刹花鼓刹车侧或后轮驱动侧</strong>，否则法兰孔将承受撕裂剪切力破坏。
        </div>
      </div>

      <div class="legend-box">
        <div class="legend-box-title">
          <span class="legend-box-num">02</span>
          <span>法兰孔邻位干涉 (Hole Overlap)</span>
        </div>
        <div class="legend-box-body">
          当交叉数过大（如 24H 选 3X 或 20H 选 2X）时，辐条引出法兰时会强行压在相邻的穿条孔上，遮挡相邻条帽安装且导致条身死折产生断条隐患，系统将触发 <strong>CRITICAL</strong> 警报。
        </div>
      </div>

      <div class="legend-box">
        <div class="legend-box-title">
          <span class="legend-box-num">03</span>
          <span>异对称 2:1 编法 (Triplet Lacing)</span>
        </div>
        <div class="legend-box-body">
          针对后轮右侧法兰过窄导致的张力极度不平衡问题，右侧采用 14/16 根交叉编织，左侧仅使用 7/8 根直拉，将左右张力平衡比强力拉升至 100:90 卓越区间。
        </div>
      </div>
    </div>
  </footer>

  <!-- ================= 核心计算与完整拓扑动态 SVG 渲染引擎 ================= -->
  
</div>
</template>

<script setup>
import { onMounted } from 'vue'

definePageMeta({
  layout: 'products',
  footer: false,
  breadcrumb: '轮组编法（内部预览）',
})

useHead({
  title: '轮组编法交互工程蓝图',
  meta: [
    {
      name: 'description',
      content: '轮组孔数、交叉编法、拓扑和法兰孔干涉的内部交互工程预览。',
    },
  ],
})

// 基础几何尺寸 (SVG 像素)
    const RIM_ERD_RADIUS = 232;      // 车圈孔位圆半径
    const RIM_OUTER_RADIUS = 260;    // 车圈外边缘
    const RIM_INNER_RADIUS = 214;    // 车圈内壁
    const HUB_PCD_RADIUS_A = 66;     // 右侧法兰节圆 PCD (Drive Side)
    const HUB_PCD_RADIUS_B = 54;     // 左侧法兰节圆 PCD (Non-Drive Side，略小以区分)
    const HUB_OUTER_RADIUS = 78;     // 花鼓法兰盘金属外缘
    const HUB_AXLE_RADIUS = 20;      // 轴心轴承壳

    let state = {
      holes: 24,            // 16, 20, 21, 24, 28, 32, 36, '24_2to1'
      cross: 2,             // 0, 1, 2, 3, 4
      viewMode: 'both',     // 默认全景双侧透视，确保所有孔位100%全满严整
      showLeading: true,
      showTrailing: true,
      showNonDrive: true,
      showNodes: true
    };

    const HOLE_RULES = {
      16: {
        label: '16H',
        allowedCross: [0, 1],
        recommended: 0,
        tips: '16孔通常用于超轻公路圈刹前轮（0X 放射）；无法支持 2X 及以上（法兰孔严重干涉）。'
      },
      20: {
        label: '20H',
        allowedCross: [0, 1, 2],
        recommended: 1,
        tips: '20孔常用于轻量公路前轮；2X 存在轻度法兰孔边缘擦碰风险，需确认法兰直径。'
      },
      21: {
        label: '21H (Campagnolo G3 2:1)',
        allowedCross: [2],
        recommended: 2,
        tips: 'Campagnolo G3 驱动侧切向 X 型交叉编法：花鼓耳引出两根交叉辐条分别锚固相邻两组车圈孔，形成坚固的抗扭三角桁架，彻底消除单侧松条失圆隐患！'
      },
      24: {
        label: '24H (1:1 标准)',
        allowedCross: [0, 1, 2, 3],
        recommended: 2,
        tips: '24孔搭配 2X 交叉是现代公路碟刹车轮的黄金标准；3X 会产生严重法兰孔干涉遮挡，严禁使用！'
      },
      '24_2to1': {
        label: '24H (2:1 异对称)',
        allowedCross: [2],
        recommended: 2,
        tips: '碟刹高阶轮组主流异对称：刹车侧/驱动侧 16 根 2X 编织，非高受力侧 8 根 0X 直拉。'
      },
      28: {
        label: '28H',
        allowedCross: [0, 1, 2, 3],
        recommended: 2,
        tips: '28孔是 Gravel 全地形、轻量山地 XC 的优选；2X 兼顾轻量与侧向刚度，3X 提供更强抗扭极限。'
      },
      32: {
        label: '32H',
        allowedCross: [0, 1, 2, 3, 4],
        recommended: 3,
        tips: '32孔 3X 交叉是自行车历史上最经典的“黄金编法”！出条切线角接近 78°，抗扭寿命极高。'
      },
      36: {
        label: '36H',
        allowedCross: [0, 1, 2, 3, 4],
        recommended: 3,
        tips: '36孔广泛用于重载旅行车、电助力货运车及高强度下坡车；支持 3X 与 4X 大角度交叉编制。'
      }
    };

    

    function setHoleCount(holes) {
      state.holes = holes;
      const rule = HOLE_RULES[holes];
      if (!rule.allowedCross.includes(state.cross)) {
        state.cross = rule.recommended;
      }
      renderControls();
      renderBlueprint();
    }

    function setCrossCount(cross) {
      state.cross = cross;
      renderControls();
      renderBlueprint();
    }

    function setViewMode(mode) {
      state.viewMode = mode;
      document.querySelectorAll('.view-mode-grid .view-pill-btn').forEach(btn => btn.classList.remove('active'));
      const activeBtn = document.getElementById(`btn-view-${mode}`);
      if (activeBtn) activeBtn.classList.add('active');
      renderBlueprint();
    }

    function updateCanvasLayers() {
      state.showLeading = document.getElementById('toggle-leading').checked;
      state.showTrailing = document.getElementById('toggle-trailing').checked;
      state.showNonDrive = document.getElementById('toggle-nondrive').checked;
      state.showNodes = document.getElementById('toggle-nodes').checked;
      renderBlueprint();
    }

    function renderControls() {
      const rule = HOLE_RULES[state.holes];

      document.querySelectorAll('.hole-pill-grid .uds-pill-btn').forEach(btn => {
        const h = btn.getAttribute('data-holes');
        if (h === String(state.holes)) {
          btn.classList.add('active');
        } else {
          btn.classList.remove('active');
        }
      });
      document.getElementById('current-hole-label').innerText = rule.label;

      document.querySelectorAll('.cross-pill-grid .uds-pill-btn').forEach(btn => {
        const c = parseInt(btn.getAttribute('data-cross'), 10);
        const isAllowed = rule.allowedCross.includes(c);
        btn.disabled = !isAllowed;
        if (c === state.cross && isAllowed) {
          btn.classList.add('active');
        } else {
          btn.classList.remove('active');
        }
      });

      const crossNames = ['0X RADIAL (直拉)', '1X CROSS (一交叉)', '2X CROSS (二交叉)', '3X CROSS (三交叉)', '4X CROSS (四交叉)'];
      document.getElementById('current-cross-label').innerText = crossNames[state.cross] || `${state.cross}X`;
    }

    const SVG_NS = 'http://www.w3.org/2000/svg';

    function appendSvgElement(parent, tagName, attributes, textContent) {
      const element = document.createElementNS(SVG_NS, tagName);
      Object.entries(attributes).forEach(([name, value]) => {
        element.setAttribute(name, String(value));
      });
      if (textContent !== undefined) element.textContent = textContent;
      parent.appendChild(element);
      return element;
    }

    // 核心渲染：完整的轮组双法兰全拓扑构建
    function renderBlueprint() {
      const svg = document.getElementById('spoke-lacing-svg');
      svg.replaceChildren();

      const holes = state.holes;
      const cross = state.cross;
      const viewMode = state.viewMode;

      // 1. 底层几何同心圆
      const bgGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
      appendSvgElement(bgGroup, 'circle', { cx: 0, cy: 0, r: RIM_OUTER_RADIUS, fill: 'none', stroke: '#cbd5e1', 'stroke-width': 2.5 });
      appendSvgElement(bgGroup, 'circle', { cx: 0, cy: 0, r: RIM_OUTER_RADIUS - 12, fill: 'none', stroke: '#e2e8f0', 'stroke-width': 1.2, 'stroke-dasharray': '6 4' });
      appendSvgElement(bgGroup, 'circle', { cx: 0, cy: 0, r: RIM_ERD_RADIUS, fill: 'rgba(5, 150, 105, 0.015)', stroke: '#059669', 'stroke-width': 1.6, 'stroke-dasharray': '4 4' });
      appendSvgElement(bgGroup, 'circle', { cx: 0, cy: 0, r: RIM_INNER_RADIUS, fill: 'none', stroke: '#e2e8f0', 'stroke-width': 1.8 });
      appendSvgElement(bgGroup, 'line', { x1: -280, y1: 0, x2: 280, y2: 0, stroke: 'rgba(15, 23, 42, 0.08)', 'stroke-width': 1, 'stroke-dasharray': '4 4' });
      appendSvgElement(bgGroup, 'line', { x1: 0, y1: -280, x2: 0, y2: 280, stroke: 'rgba(15, 23, 42, 0.08)', 'stroke-width': 1, 'stroke-dasharray': '4 4' });
      svg.appendChild(bgGroup);

      // 2. 严密的轮组拓扑连线数据结构
      const rimHoles = [];      // 车圈所有穿条孔
      const hubHolesA = [];     // 侧 A (右法兰 / 驱动侧) 孔位
      const hubHolesB = [];     // 侧 B (左法兰 / 非驱动侧) 孔位
      const spokes = [];        // 全部辐条线段集合
      const crossNodes = [];    // 交叉相交节点

      // ==========================================
      // 分支一：21H (Campagnolo G3 异对称) 专属精确拓扑
      // ==========================================
      if (holes === 21) {
        const groups = 7;
        const delta = 0.085; // 组内三孔跨距

        // 车圈 21 孔：7 组紧凑三孔
        for (let g = 0; g < groups; g++) {
          const centerAngle = (g * 2 * Math.PI) / groups - Math.PI / 2;
          
          // 孔 0: 驱动侧推力条孔 (左)
          const a0 = centerAngle - delta;
          const r0 = { id: g * 3, side: 'A', role: 'trailing', angle: a0, x: RIM_ERD_RADIUS * Math.cos(a0), y: RIM_ERD_RADIUS * Math.sin(a0) };
          rimHoles.push(r0);

          // 孔 1: 非驱动侧直拉孔 (中)
          const a1 = centerAngle;
          const r1 = { id: g * 3 + 1, side: 'B', role: 'radial', angle: a1, x: RIM_ERD_RADIUS * Math.cos(a1), y: RIM_ERD_RADIUS * Math.sin(a1) };
          rimHoles.push(r1);

          // 孔 2: 驱动侧牵引条孔 (右)
          const a2 = centerAngle + delta;
          const r2 = { id: g * 3 + 2, side: 'A', role: 'leading', angle: a2, x: RIM_ERD_RADIUS * Math.cos(a2), y: RIM_ERD_RADIUS * Math.sin(a2) };
          rimHoles.push(r2);
        }

        // 驱动侧右法兰 14 孔 (在法兰圆上严格 14 等分均分，间距恒定 25.71°)
        for (let g = 0; g < groups; g++) {
          const centerAngle = (g * 2 * Math.PI) / groups - Math.PI / 2;
          const midAngle = centerAngle + (Math.PI / groups);
          const angleLeft = midAngle - (Math.PI / 14);
          const angleRight = midAngle + (Math.PI / 14);

          hubHolesA.push({
            id: g * 2,
            angle: angleLeft,
            x: HUB_PCD_RADIUS_A * Math.cos(angleLeft),
            y: HUB_PCD_RADIUS_A * Math.sin(angleLeft)
          });
          hubHolesA.push({
            id: g * 2 + 1,
            angle: angleRight,
            x: HUB_PCD_RADIUS_A * Math.cos(angleRight),
            y: HUB_PCD_RADIUS_A * Math.sin(angleRight)
          });
        }

        // 左法兰 7 孔 (严格位于 7 组中心线上，保证直拉条 100% 穿过圆心)
        for (let k = 0; k < 7; k++) {
          const angle = (k * 2 * Math.PI) / 7 - Math.PI / 2;
          hubHolesB.push({ id: k, angle, x: HUB_PCD_RADIUS_B * Math.cos(angle), y: HUB_PCD_RADIUS_B * Math.sin(angle) });
        }

        // 1. 非驱动侧：7 根直拉辐条 (蓝线)，沿中心线 100% 直达车圈中孔
        for (let g = 0; g < groups; g++) {
          spokes.push({
            side: 'B',
            type: 'nondrive',
            hub: hubHolesB[g],
            rim: rimHoles[g * 3 + 1]
          });
        }

        // 2. 驱动侧：14 根交叉辐条 (严格对齐用户最新截图红箭头：每个花鼓耳引出两根交叉 X 辐条！)
        // - 花鼓耳左孔 (2g): 跨向顺时针车圈组 (g+1) 的左孔 rimHoles[((g+1)%7)*3]
        // - 花鼓耳右孔 (2g+1): 跨向逆时针车圈组 g 的右孔 rimHoles[g*3+2]
        // 两根红箭头在正下方 (6 点钟) 与全轮 7 处完美形成“X”形交叉！
        for (let g = 0; g < groups; g++) {
          const nextGroup = (g + 1) % groups;
          spokes.push({
            side: 'A',
            type: 'trailing',
            hub: hubHolesA[g * 2],
            rim: rimHoles[nextGroup * 3]
          });
          spokes.push({
            side: 'A',
            type: 'leading',
            hub: hubHolesA[g * 2 + 1],
            rim: rimHoles[g * 3 + 2]
          });
        }
      }
      // ==========================================
      // 分支二：24H (2:1 异对称) 专属精确拓扑
      // ==========================================
      else if (holes === '24_2to1') {
        const total = 24;
        // 24 孔车圈均布，每 3 孔为一单元 (2 个驱动侧 A，1 个非驱动侧 B)
        for (let i = 0; i < total; i++) {
          const angle = (i * 2 * Math.PI) / total - Math.PI / 2 + (Math.PI / total);
          const isNonDrive = (i % 3 === 1);
          rimHoles.push({
            id: i,
            side: isNonDrive ? 'B' : 'A',
            angle,
            x: RIM_ERD_RADIUS * Math.cos(angle),
            y: RIM_ERD_RADIUS * Math.sin(angle)
          });
        }

        const bRimHoles = rimHoles.filter(r => r.side === 'B');
        const aRimHoles = rimHoles.filter(r => r.side === 'A');

        // 驱动侧 16 孔花鼓
        for (let j = 0; j < 16; j++) {
          const angle = (j * 2 * Math.PI) / 16 - Math.PI / 2;
          hubHolesA.push({ id: j, angle, x: HUB_PCD_RADIUS_A * Math.cos(angle), y: HUB_PCD_RADIUS_A * Math.sin(angle) });
        }

        // 非驱动侧 8 孔花鼓 (0X 直拉)：极角严格对齐车圈侧 B 孔位，保证直通圆心
        for (let k = 0; k < 8; k++) {
          const angle = bRimHoles[k].angle;
          hubHolesB.push({ id: k, angle, x: HUB_PCD_RADIUS_B * Math.cos(angle), y: HUB_PCD_RADIUS_B * Math.sin(angle) });
        }

        // 非驱动侧 8 根 0X 直拉
        for (let k = 0; k < 8; k++) {
          spokes.push({ side: 'B', type: 'nondrive', hub: hubHolesB[k], rim: bRimHoles[k] });
        }
        for (let j = 0; j < 16; j++) {
          const isEven = (j % 2 === 0);
          const hub = hubHolesA[j];
          if (isEven) {
            const targetIdx = (j + 2) % 16;
            spokes.push({ side: 'A', type: 'leading', hub, rim: aRimHoles[targetIdx] });
          } else {
            const targetIdx = (j - 2 + 16) % 16;
            spokes.push({ side: 'A', type: 'trailing', hub, rim: aRimHoles[targetIdx] });
          }
        }
      }
      // ==========================================
      // 分支三：标准等分对称轮组 (16H, 20H, 24H, 28H, 32H, 36H)
      // ==========================================
      else {
        const total = parseInt(holes, 10);
        const flangeCount = total / 2; // 单侧法兰孔数

        // 车圈均布 total 个孔，交替分配给 A 侧 (偶数) 和 B 侧 (奇数)
        for (let i = 0; i < total; i++) {
          const angle = (i * 2 * Math.PI) / total - Math.PI / 2 + (Math.PI / total);
          const isSideA = (i % 2 === 0);
          rimHoles.push({
            id: i,
            side: isSideA ? 'A' : 'B',
            angle,
            x: RIM_ERD_RADIUS * Math.cos(angle),
            y: RIM_ERD_RADIUS * Math.sin(angle)
          });
        }

        const rimA = rimHoles.filter(r => r.side === 'A');
        const rimB = rimHoles.filter(r => r.side === 'B');

        // 右法兰 A (外圈 PCD) - 极角严格对齐车圈侧 A 孔位，保证 0X 放射时完全沿半径直穿圆心
        for (let j = 0; j < flangeCount; j++) {
          const angle = rimA[j].angle;
          hubHolesA.push({ id: j, angle, x: HUB_PCD_RADIUS_A * Math.cos(angle), y: HUB_PCD_RADIUS_A * Math.sin(angle) });
        }

        // 左法兰 B (内圈 PCD) - 极角严格对齐车圈侧 B 孔位，保证 0X 放射时完全沿半径直穿圆心
        for (let k = 0; k < flangeCount; k++) {
          const angle = rimB[k].angle;
          hubHolesB.push({ id: k, angle, x: HUB_PCD_RADIUS_B * Math.cos(angle), y: HUB_PCD_RADIUS_B * Math.sin(angle) });
        }

        if (cross === 0) {
          // 0X 放射编法：直达对应车圈孔
          for (let j = 0; j < flangeCount; j++) {
            spokes.push({ side: 'A', type: 'leading', hub: hubHolesA[j], rim: rimA[j] });
            spokes.push({ side: 'B', type: 'nondrive', hub: hubHolesB[j], rim: rimB[j] });
          }
        } else {
          // 交叉编法：跨孔步长 span = cross
          const span = cross;

          // 侧 A (右法兰驱动侧)：偶数孔牵引，奇数孔推力
          for (let j = 0; j < flangeCount; j++) {
            const isLeading = (j % 2 === 0);
            const hub = hubHolesA[j];
            if (isLeading) {
              const target = (j + span) % flangeCount;
              spokes.push({ side: 'A', type: 'leading', hub, rim: rimA[target] });
            } else {
              const target = (j - span + flangeCount) % flangeCount;
              spokes.push({ side: 'A', type: 'trailing', hub, rim: rimA[target] });
            }
          }

          // 侧 B (左法兰非驱动侧)：偶数孔牵引，奇数孔推力
          for (let k = 0; k < flangeCount; k++) {
            const isLeading = (k % 2 === 0);
            const hub = hubHolesB[k];
            if (isLeading) {
              const target = (k + span) % flangeCount;
              spokes.push({ side: 'B', type: 'nondrive', hub, rim: rimB[target] });
            } else {
              const target = (k - span + flangeCount) % flangeCount;
              spokes.push({ side: 'B', type: 'nondrive', hub, rim: rimB[target] });
            }
          }
        }
      }

      // 3. 计算可见辐条的真实物理交叉咬合点 (Cross Nodes)
      const visibleSpokes = spokes.filter(s => {
        if (viewMode === 'drive' && s.side !== 'A') return false;
        if (viewMode === 'nondrive' && s.side !== 'B') return false;
        if (s.side === 'A' && s.type === 'leading' && !state.showLeading) return false;
        if (s.side === 'A' && s.type === 'trailing' && !state.showTrailing) return false;
        if (s.side === 'B' && !state.showNonDrive) return false;
        return true;
      });

      // 同侧的异向辐条相交检测
      for (let m = 0; m < visibleSpokes.length; m++) {
        for (let n = m + 1; n < visibleSpokes.length; n++) {
          const s1 = visibleSpokes[m];
          const s2 = visibleSpokes[n];
          if (s1.side === s2.side && s1.type !== s2.type) {
            const pt = getIntersection(s1.hub, s1.rim, s2.hub, s2.rim);
            if (pt) crossNodes.push(pt);
          }
        }
      }

      // 4. 先绘制花鼓底层金属底盘与 PCD 节圆 (置于辐条下方，避免遮挡辐条穿入孔心)
      const hubBaseGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
      hubBaseGroup.setAttribute('id', 'layer-hub-base');
      appendSvgElement(hubBaseGroup, 'circle', { cx: 0, cy: 0, r: HUB_OUTER_RADIUS, fill: '#f1f5f9', stroke: '#475569', 'stroke-width': 2 });
      appendSvgElement(hubBaseGroup, 'circle', { cx: 0, cy: 0, r: HUB_PCD_RADIUS_A, fill: 'none', stroke: '#d97706', 'stroke-width': 1.4, 'stroke-dasharray': '3 3' });
      appendSvgElement(hubBaseGroup, 'circle', { cx: 0, cy: 0, r: HUB_PCD_RADIUS_B, fill: 'none', stroke: '#0284c7', 'stroke-width': 1.2, 'stroke-dasharray': '2 2' });
      appendSvgElement(hubBaseGroup, 'circle', { cx: 0, cy: 0, r: HUB_AXLE_RADIUS, fill: '#0f172a', stroke: '#cbd5e1', 'stroke-width': 1.8 });
      appendSvgElement(hubBaseGroup, 'circle', { cx: 0, cy: 0, r: 5, fill: '#38bdf8' });
      svg.appendChild(hubBaseGroup);

      // 5. 绘制辐条图层 (从 PCD 均分孔心到车圈孔心，完整暴露无遮挡)
      const spokeGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
      spokeGroup.setAttribute('id', 'layer-spokes');
      visibleSpokes.forEach((spk, idx) => {
        let strokeColor = '#059669';
        let strokeWidth = 2.4;
        let opacity = 1.0;

        if (spk.side === 'A') {
          strokeColor = (spk.type === 'leading') ? '#059669' : '#d97706';
        } else {
          strokeColor = '#0284c7';
          if (viewMode === 'both') {
            strokeWidth = 1.8;
            opacity = 0.65;
          }
        }

        const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        line.setAttribute('x1', spk.hub.x);
        line.setAttribute('y1', spk.hub.y);
        line.setAttribute('x2', spk.rim.x);
        line.setAttribute('y2', spk.rim.y);
        line.setAttribute('stroke', strokeColor);
        line.setAttribute('stroke-width', strokeWidth);
        line.setAttribute('opacity', opacity);
        line.setAttribute('stroke-linecap', 'round');
        line.setAttribute('class', 'spoke-line');

        const title = document.createElementNS('http://www.w3.org/2000/svg', 'title');
        title.textContent = `辐条 #${idx + 1} [${spk.side === 'A' ? '驱动侧/右' : '非驱动侧/左'}]: 花鼓孔(${spk.hub.x.toFixed(1)}, ${spk.hub.y.toFixed(1)}) -> 车圈孔(${spk.rim.x.toFixed(1)}, ${spk.rim.y.toFixed(1)})`;
        line.appendChild(title);

        spokeGroup.appendChild(line);
      });
      svg.appendChild(spokeGroup);

      // 6. 绘制物理交叉咬合点 (Cross Nodes)
      if (state.showNodes && cross > 0) {
        const nodeGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        crossNodes.forEach(pt => {
          const c = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
          c.setAttribute('cx', pt.x);
          c.setAttribute('cy', pt.y);
          c.setAttribute('r', 3.6);
          c.setAttribute('class', 'spoke-node-circle');
          const title = document.createElementNS('http://www.w3.org/2000/svg', 'title');
          title.textContent = `交叉咬合节点 (Interlacing Crossing Point)`;
          c.appendChild(title);
          nodeGroup.appendChild(c);
        });
        svg.appendChild(nodeGroup);
      }

      // 7. 绘制顶层花鼓穿条孔与铆钉 (条头压在辐条上方，呈现清晰穿孔咬合质感)
      const hubHolesTopGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
      hubHolesTopGroup.setAttribute('id', 'layer-hub-holes-top');

      // 绘制右法兰 PCD 均分孔 (橙心)
      if (viewMode === 'drive' || viewMode === 'both') {
        hubHolesA.forEach(h => {
          appendSvgElement(hubHolesTopGroup, 'circle', { cx: h.x, cy: h.y, r: 3.4, fill: '#ffffff', stroke: '#0f172a', 'stroke-width': 1.6 });
          appendSvgElement(hubHolesTopGroup, 'circle', { cx: h.x, cy: h.y, r: 1.6, fill: '#d97706' });
        });
      }

      // 绘制左法兰 PCD 均分孔 (蓝心)
      if (viewMode === 'nondrive' || viewMode === 'both') {
        hubHolesB.forEach(h => {
          appendSvgElement(hubHolesTopGroup, 'circle', { cx: h.x, cy: h.y, r: 3.0, fill: '#ffffff', stroke: '#0284c7', 'stroke-width': 1.4 });
          appendSvgElement(hubHolesTopGroup, 'circle', { cx: h.x, cy: h.y, r: 1.3, fill: '#0284c7' });
        });
      }
      svg.appendChild(hubHolesTopGroup);

      // 8. 绘制车圈穿条孔 (全量着色，覆盖辐条外端)
      const rimHoleGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
      rimHoleGroup.setAttribute('id', 'layer-rim-holes');
      rimHoles.forEach((r, idx) => {
        let holeFill = (r.side === 'A') ? '#059669' : '#0284c7';
        appendSvgElement(rimHoleGroup, 'circle', { cx: r.x, cy: r.y, r: 4.2, fill: '#ffffff', stroke: holeFill, 'stroke-width': 2 });
        appendSvgElement(rimHoleGroup, 'circle', { cx: r.x, cy: r.y, r: 1.8, fill: '#0f172a' });
      });

      svg.appendChild(rimHoleGroup);

      // 8. 实时遥测力学与干涉审计
      updateTelemetryAndAudit(holes, cross, crossNodes.length);
    }

    function getIntersection(p0, p1, p2, p3) {
      const s1_x = p1.x - p0.x;
      const s1_y = p1.y - p0.y;
      const s2_x = p3.x - p2.x;
      const s2_y = p3.y - p2.y;

      const denom = (-s2_x * s1_y + s1_x * s2_y);
      if (Math.abs(denom) < 0.0001) return null;

      const s = (-s1_y * (p0.x - p2.x) + s1_x * (p0.y - p2.y)) / denom;
      const t = ( s2_x * (p0.y - p2.y) - s2_y * (p0.x - p2.x)) / denom;

      if (s >= 0.08 && s <= 0.92 && t >= 0.08 && t <= 0.92) {
        return {
          x: p0.x + (t * s1_x),
          y: p0.y + (t * s1_y)
        };
      }
      return null;
    }

    function updateTelemetryAndAudit(holes, cross, nodeCount) {
      let numHoles = (holes === '24_2to1' || holes === 21) ? (holes === 21 ? 21 : 24) : parseInt(holes, 10);
      let tangentAngle = 0;

      if (holes === 21) {
        tangentAngle = 45.0;
      } else if (cross > 0) {
        const betaRad = (cross * 2 * 2 * Math.PI) / numHoles;
        const L = Math.sqrt(
          RIM_ERD_RADIUS * RIM_ERD_RADIUS + HUB_PCD_RADIUS_A * HUB_PCD_RADIUS_A - 
          2 * RIM_ERD_RADIUS * HUB_PCD_RADIUS_A * Math.cos(betaRad)
        );
        const sinAlpha = (RIM_ERD_RADIUS / L) * Math.sin(betaRad);
        tangentAngle = Math.asin(Math.min(1.0, Math.max(0, sinAlpha))) * (180 / Math.PI);
        if (tangentAngle > 88.5) tangentAngle = 88.5;
      }

      const rad = (tangentAngle * Math.PI) / 180;
      const torqueRatio = Math.sin(rad) * 100;
      const lateralRatio = Math.cos(rad) * 100;

      document.getElementById('metric-tangent-angle').innerText = tangentAngle.toFixed(1);
      document.getElementById('bar-tangent-angle').style.width = `${(tangentAngle / 90) * 100}%`;

      document.getElementById('metric-torque-efficiency').innerText = torqueRatio.toFixed(1);
      document.getElementById('bar-torque-efficiency').style.width = `${torqueRatio}%`;

      document.getElementById('metric-lateral-index').innerText = lateralRatio.toFixed(1);
      document.getElementById('bar-lateral-index').style.width = `${lateralRatio}%`;

      document.getElementById('metric-node-count').innerText = nodeCount;

      // 干涉审计
      const statusBox = document.getElementById('clearance-status-box');
      const pulseDot = document.getElementById('status-pulse-dot');
      const statusTitle = document.getElementById('status-title-text');
      const statusDetail = document.getElementById('status-detail-text');
      const builderTip = document.getElementById('builder-tip-text');
      const statusTag = document.getElementById('canvas-status-tag');

      if (holes === 21) {
        statusBox.className = 'clearance-status-card status-healthy';
        pulseDot.style.display = 'none';
        statusTitle.innerText = 'HEALTHY: Campagnolo G3 驱动侧 X 型交叉专利结构';
        statusDetail.innerText = '车圈 7 组独立成束，每个花鼓耳引出两根交叉辐条分别锚固相邻两组车圈孔，形成坚固的抗扭三角桁架，彻底消除单侧松条失圆隐患！';
        builderTip.innerText = '完全对齐 Campagnolo G3 官方后轮专利结构：花鼓耳双孔出条交叉成 X 型桁架，左右张力 1:1 完美均等。';
        statusTag.innerText = 'G3 X-CROSS OPTIMAL';
        statusTag.style.background = 'rgba(5, 150, 105, 0.1)';
        statusTag.style.color = '#059669';
        return;
      }

      const isCritical = (
        (numHoles === 24 && cross >= 3) ||
        (numHoles === 20 && cross >= 2) ||
        (numHoles === 16 && cross >= 2) ||
        (numHoles === 28 && cross >= 4)
      );

      const isAlert = (cross === 0 || (cross === 1 && numHoles <= 24));

      if (isCritical) {
        statusBox.className = 'clearance-status-card status-critical';
        pulseDot.style.display = 'inline-block';
        pulseDot.style.backgroundColor = 'var(--accent-rose)';
        statusTitle.innerText = 'CRITICAL: 法兰孔严重干涉！';
        statusDetail.textContent = `在该孔数（${numHoles}H）下选用 ${cross}X 交叉，辐条出条切线偏角过大，强行压过相邻法兰孔的条头！存在无法就位与法兰撕裂隐患，严禁装配！`;
        builderTip.innerText = `必须将交叉数降低至 ${Math.max(1, cross - 1)}X；或选用小 PCD 节圆花鼓。`;
        statusTag.innerText = 'CRITICAL INTERFERENCE';
        statusTag.style.background = 'rgba(225, 29, 72, 0.1)';
        statusTag.style.color = '#e11d48';
      } else if (isAlert) {
        statusBox.className = 'clearance-status-card status-alert';
        pulseDot.style.display = 'none';
        statusTitle.innerText = cross === 0 ? 'ALERT: 放射编法受限' : 'ALERT: 微力矩过渡编法';
        statusDetail.innerText = cross === 0 ? 
          '0X 放射编法抗扭力矩为 0。仅可用于圈刹公路前轮。绝对严禁在碟刹刹车侧或后轮驱动侧使用，否则将直接撕裂法兰孔！' :
          '1X 交叉出条角较小，扭转响应偏软，通常仅适用于小法兰折叠车或超轻爬坡轮。';
        builderTip.innerText = cross === 0 ? '碟刹轮组必须采用 2X 或 3X 交叉编制。' : '建议优先考虑 2X 交叉以获得更佳扭转刚度。';
        statusTag.innerText = 'CAUTION ADVISED';
        statusTag.style.background = 'rgba(217, 119, 6, 0.1)';
        statusTag.style.color = '#d97706';
      } else {
        statusBox.className = 'clearance-status-card status-healthy';
        pulseDot.style.display = 'none';
        statusTitle.innerText = 'HEALTHY: 黄金工程几何平衡';
        statusDetail.innerText = `${numHoles}孔搭配 ${cross}X 交叉组合为标准成熟结构。切线出条角处于 ${tangentAngle.toFixed(1)}° 的黄金受力区间，无死折与干涉风险。`;
        builderTip.innerText = HOLE_RULES[holes].tips || '左右张力平衡度与耐疲劳寿命表现优异。';
        statusTag.innerText = 'ENGINEERING OPTIMAL';
        statusTag.style.background = 'rgba(5, 150, 105, 0.1)';
        statusTag.style.color = '#059669';
      }
    }

onMounted(() => {
  renderControls()
  renderBlueprint()
})

</script>

<style scoped>
/* ==========================================================================
       ERP UDS 工业设计系统规范 (v1.0) - 浅色工业工程图纸标准 (Light Blueprint)
       ========================================================================== */
    .spoke-lacing-page {
      --bg-base: #f8fafc;            /* Slate-50 */
      --bg-surface: #ffffff;         /* 纯白 */
      --bg-card: #ffffff;
      --bg-card-hover: #f1f5f9;
      --bg-inset: #f1f5f9;           /* Slate-100 */
      --border-line: rgba(15, 23, 42, 0.08);
      --border-dashed: 1px dashed rgba(15, 23, 42, 0.16);
      --border-focus: #0f172a;

      /* 语义色 */
      --accent-primary: #059669;     /* Tanzanite 翡翠绿 (驱动侧牵引条) */
      --accent-amber: #d97706;       /* 琥珀金 (驱动侧推力条) */
      --accent-sky: #0284c7;         /* 天蓝 (非驱动侧直拉/交叉) */
      --accent-rose: #e11d48;        /* 玫瑰红 (物理交叉咬合点 / 干涉告警) */
      --accent-steel: #475569;       /* 机械冷钢 (基准尺寸) */
      --accent-indigo: #4f46e5;      /* 气门嘴紫蓝 */

      /* 文字色阶 */
      --text-main: #0f172a;
      --text-secondary: #334155;
      --text-muted: #64748b;
      --text-dim: #94a3b8;
    }

    .spoke-lacing-page * {
      box-sizing: border-box;
      margin: 0;
      padding: 0;
    }

    .spoke-lacing-page {
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
      max-width: 1440px;
      margin: 0 auto;
    }

    /* 顶部工程页眉 */
    .blueprint-header {
      background: var(--bg-surface);
      border: var(--border-dashed);
      border-radius: 32px;
      padding: 22px 36px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      position: relative;
      overflow: hidden;
      box-shadow: 0 10px 25px -5px rgba(15, 23, 42, 0.05);
    }

    .blueprint-header::before {
      content: "";
      position: absolute;
      inset: 0;
      background: linear-gradient(135deg, rgba(5, 150, 105, 0.03), transparent 60%);
      pointer-events: none;
    }

    .header-title-group h1 {
      font-size: 1.15rem;
      font-weight: 900;
      letter-spacing: -0.05em;
      font-style: italic;
      text-transform: uppercase;
      color: var(--text-main);
      display: flex;
      align-items: center;
      gap: 12px;
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
      color: var(--accent-primary);
      border-color: rgba(5, 150, 105, 0.25);
    }

    .header-subtitle {
      font-size: 9px;
      font-weight: 900;
      text-transform: uppercase;
      letter-spacing: 0.14em;
      color: var(--text-muted);
      opacity: 0.75;
      margin-top: 4px;
    }

    /* 顶层主网格排版 */
    .blueprint-grid {
      display: grid;
      grid-template-columns: 360px 1fr 340px;
      gap: 20px;
      align-items: start;
    }

    @media (max-width: 1280px) {
      .blueprint-grid {
        grid-template-columns: 320px 1fr;
      }
      .telemetry-column {
        grid-column: 1 / -1;
      }
    }

    @media (max-width: 960px) {
      .blueprint-grid {
        grid-template-columns: 1fr;
      }
    }

    /* 统一卡片样式 */
    .uds-card {
      background: var(--bg-surface);
      border: var(--border-dashed);
      border-radius: 24px;
      padding: 20px;
      box-shadow: 0 4px 12px rgba(15, 23, 42, 0.02);
      display: flex;
      flex-direction: column;
      gap: 16px;
      position: relative;
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      border-bottom: 1px solid var(--border-line);
      padding-bottom: 12px;
    }

    .card-title {
      font-size: 0.85rem;
      font-weight: 900;
      letter-spacing: -0.03em;
      font-style: italic;
      text-transform: uppercase;
      color: var(--text-main);
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .card-tag {
      font-family: var(--tz-font-ui);
      font-size: 8px;
      font-weight: 800;
      padding: 2px 6px;
      border-radius: 9999px;
      background: var(--bg-inset);
      color: var(--text-muted);
    }

    /* 控制面板按钮网格 */
    .selector-group-label {
      font-size: 10px;
      font-weight: 900;
      text-transform: uppercase;
      letter-spacing: 0.1em;
      color: var(--text-muted);
      margin-bottom: 8px;
      display: flex;
      justify-content: space-between;
    }

    .hole-pill-grid {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 8px;
    }

    .cross-pill-grid {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 8px;
    }

    .view-mode-grid {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 6px;
      background: var(--bg-inset);
      padding: 4px;
      border-radius: 14px;
    }

    .uds-pill-btn {
      height: 42px;
      border-radius: 12px;
      border: 1px solid var(--border-line);
      background: var(--bg-inset);
      color: var(--text-secondary);
      font-family: var(--tz-font-ui);
      font-size: 11px;
      font-weight: 800;
      cursor: pointer;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 2px;
      transition: all 0.18s ease;
    }

    .uds-pill-btn:hover:not(:disabled) {
      background: #ffffff;
      border-color: rgba(5, 150, 105, 0.4);
      color: var(--accent-primary);
      transform: translateY(-1px);
    }

    .uds-pill-btn.active {
      background: #ffffff;
      border: 1.5px solid var(--accent-primary);
      color: var(--accent-primary);
      box-shadow: 0 4px 12px rgba(5, 150, 105, 0.12);
    }

    .uds-pill-btn:disabled {
      opacity: 0.35;
      cursor: not-allowed;
      border-color: transparent;
      background: rgba(15, 23, 42, 0.04);
      color: var(--text-dim);
    }

    .uds-pill-btn .sub-text {
      font-size: 7.5px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      opacity: 0.75;
    }

    .view-pill-btn {
      height: 32px;
      border-radius: 10px;
      border: 1px solid transparent;
      background: transparent;
      color: var(--text-muted);
      font-size: 9.5px;
      font-weight: 800;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .view-pill-btn.active {
      background: #ffffff;
      color: var(--text-main);
      border-color: rgba(15, 23, 42, 0.1);
      box-shadow: 0 2px 6px rgba(15, 23, 42, 0.04);
    }

    /* 图层控制项 */
    .layer-toggle-group {
      display: flex;
      flex-direction: column;
      gap: 8px;
      background: var(--bg-inset);
      border-radius: 16px;
      padding: 12px;
    }

    .layer-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-size: 11px;
      font-weight: 800;
      color: var(--text-secondary);
      cursor: pointer;
      user-select: none;
    }

    .layer-item input[type="checkbox"] {
      accent-color: var(--accent-primary);
      width: 15px;
      height: 15px;
      cursor: pointer;
    }

    .layer-indicator {
      display: inline-block;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      margin-right: 6px;
    }

    /* 中央图纸画布 */
    .canvas-card {
      padding: 0;
      overflow: hidden;
      display: flex;
      flex-direction: column;
      align-items: center;
    }

    .canvas-toolbar {
      width: 100%;
      padding: 12px 20px;
      border-bottom: 1px solid var(--border-line);
      display: flex;
      justify-content: space-between;
      align-items: center;
      background: var(--bg-surface);
      z-index: 2;
    }

    .canvas-legend-inline {
      display: flex;
      gap: 14px;
      align-items: center;
      font-size: 9.5px;
      font-weight: 800;
      font-family: var(--tz-font-ui);
      flex-wrap: wrap;
    }

    .svg-viewport {
      width: 100%;
      height: 620px;
      position: relative;
      background: radial-gradient(circle at center, rgba(15, 23, 42, 0.015) 0%, rgba(15, 23, 42, 0.04) 100%);
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .svg-viewport svg {
      width: 100%;
      height: 100%;
      max-width: 600px;
      max-height: 600px;
      filter: drop-shadow(0 12px 24px rgba(15, 23, 42, 0.04));
    }

    /* 辐条线交互 */
    .spoke-lacing-page :deep(.spoke-line) {
      transition: stroke 0.15s ease, stroke-width 0.15s ease, opacity 0.2s ease;
      cursor: pointer;
    }

    .spoke-lacing-page :deep(.spoke-line:hover) {
      stroke-width: 4px !important;
      filter: drop-shadow(0 0 6px rgba(5, 150, 105, 0.8));
    }

    .spoke-lacing-page :deep(.spoke-node-circle) {
      fill: #ffffff;
      stroke: var(--accent-rose);
      stroke-width: 1.5;
      transition: r 0.15s ease;
    }

    .spoke-lacing-page :deep(.spoke-node-circle:hover) {
      r: 6;
      fill: var(--accent-rose);
    }

    /* 右侧工程遥测看板 */
    .telemetry-column {
      display: flex;
      flex-direction: column;
      gap: 20px;
    }

    .telemetry-metric-tile {
      background: var(--bg-inset);
      border-radius: 16px;
      padding: 14px 16px;
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    .metric-top {
      display: flex;
      justify-content: space-between;
      align-items: baseline;
    }

    .metric-label {
      font-size: 9.5px;
      font-weight: 900;
      text-transform: uppercase;
      letter-spacing: 0.08em;
      color: var(--text-muted);
    }

    .metric-value {
      font-family: var(--tz-font-ui);
      font-size: 1.45rem;
      font-weight: 900;
      color: var(--text-main);
      display: flex;
      align-items: baseline;
      gap: 4px;
    }

    .metric-unit {
      font-size: 11px;
      font-weight: 700;
      color: var(--text-muted);
    }

    .metric-bar {
      height: 4px;
      width: 100%;
      background: rgba(15, 23, 42, 0.08);
      border-radius: 9999px;
      overflow: hidden;
      margin-top: 6px;
    }

    .metric-bar-fill {
      height: 100%;
      border-radius: 9999px;
      transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1), background-color 0.3s ease;
    }

    .metric-desc {
      font-size: 9px;
      color: var(--text-muted);
      line-height: 1.35;
      margin-top: 2px;
    }

    /* 干涉状态卡片 */
    .clearance-status-card {
      border-radius: 18px;
      padding: 14px 16px;
      border: 1px solid transparent;
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    .status-healthy {
      background: rgba(5, 150, 105, 0.06);
      border-color: rgba(5, 150, 105, 0.25);
      color: #047857;
    }

    .status-alert {
      background: rgba(217, 119, 6, 0.06);
      border-color: rgba(217, 119, 6, 0.25);
      color: #b45309;
    }

    .status-critical {
      background: rgba(225, 29, 72, 0.06);
      border-color: rgba(225, 29, 72, 0.3);
      color: #be123c;
    }

    .status-header-line {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 11px;
      font-weight: 900;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }

    .status-detail-text {
      font-size: 9.5px;
      line-height: 1.45;
      color: var(--text-secondary);
    }

    /* 底部技术规范对照表 */
    .cad-legend-section {
      background: var(--bg-surface);
      border: var(--border-dashed);
      border-radius: 24px;
      padding: 24px;
      display: flex;
      flex-direction: column;
      gap: 16px;
    }

    .legend-table-grid {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 14px;
    }

    @media (max-width: 1024px) {
      .legend-table-grid {
        grid-template-columns: repeat(2, 1fr);
      }
    }

    @media (max-width: 640px) {
      .legend-table-grid {
        grid-template-columns: 1fr;
      }
    }

    .legend-box {
      border: 1px solid var(--border-line);
      border-radius: 16px;
      padding: 12px 14px;
      background: var(--bg-base);
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    .legend-box-title {
      font-size: 10px;
      font-weight: 900;
      color: var(--text-main);
      display: flex;
      align-items: center;
      gap: 6px;
    }

    .legend-box-num {
      width: 18px;
      height: 18px;
      border-radius: 6px;
      background: #0f172a;
      color: #ffffff;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      font-family: var(--tz-font-ui);
      font-size: 9px;
      font-weight: 900;
    }

    .legend-box-body {
      font-size: 9.5px;
      color: var(--text-secondary);
      line-height: 1.45;
    }


    .pulse-dot {
      width: 7px;
      height: 7px;
      border-radius: 50%;
      background: currentColor;
      box-shadow: 0 0 0 0 rgba(225, 29, 72, 0.7);
      animation: pulse-ring 1.8s cubic-bezier(0.24, 0, 0.38, 1) infinite;
    }

    @keyframes pulse-ring {
      0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(225, 29, 72, 0.7); }
      70% { transform: scale(1.1); box-shadow: 0 0 0 6px rgba(225, 29, 72, 0); }
      100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(225, 29, 72, 0); }
    }

/* Nodes are inserted into the SVG by the blueprint renderer after mount. */
.spoke-lacing-page :deep(.spoke-line),
.spoke-lacing-page :deep(.spoke-node-circle) {
  vector-effect: non-scaling-stroke;
}
</style>

