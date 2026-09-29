<template>
  <section class="spoke-head-measurement-guide" aria-labelledby="spoke-head-measurement-guide-title">
    <header class="spoke-head-measurement-guide__header">
      <div>
        <p class="spoke-head-measurement-guide__kicker">{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideKicker', '第一步 · 图示说明') }}</p>
        <h3 id="spoke-head-measurement-guide-title" class="spoke-head-measurement-guide__title">{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideTitle', '弯头与直拉式测长基准') }}</h3>
      </div>
      <span class="spoke-head-measurement-guide__badge">{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideBadge', 'MEASUREMENT DATUM') }}</span>
    </header>

    <div class="spoke-head-measurement-guide__grid">
      <div class="spoke-head-measurement-guide__diagram" role="img" :aria-label="t('resourcesSpokeCalculator.calculator.physicalCorrections.guideAriaLabel', '弯头与直拉式辐条测量基准示意图')">
<svg viewBox="0 0 760 510" preserveAspectRatio="xMidYMid meet">
          <defs>
            <marker id="arrow-emerald" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
              <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#059669" />
            </marker>
            <marker id="arrow-rose" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
              <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#e11d48" />
            </marker>
            <marker id="arrow-amber" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
              <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#d97706" />
            </marker>
          </defs>
          <!-- 背景工程坐标网格 -->
          <pattern id="grid-pattern-main" width="20" height="20" patternUnits="userSpaceOnUse">
            <line x1="0" y1="0" x2="20" y2="0" stroke="rgba(15, 23, 42, 0.035)" stroke-width="1" />
            <line x1="0" y1="0" x2="0" y2="20" stroke="rgba(15, 23, 42, 0.035)" stroke-width="1" />
          </pattern>
          <rect width="100%" height="100%" fill="url(#grid-pattern-main)" />

          <!-- 主标题与规格栏 -->
          <text x="380" y="24" fill="#0f172a" font-size="12" font-weight="900" text-anchor="middle" letter-spacing="0.08em">
            01: SPOKE LENGTH MEASUREMENT DATUM SCHEMATIC
          </text>
          <text x="380" y="39" fill="#64748b" font-size="8.5" text-anchor="middle">
            ISO 9633:2001 / DIN 79100 / J-BEND & STRAIGHTPULL CALIBRATION DATUM
          </text>

          <!-- 上下分界虚线 -->
          <line x1="20" y1="250" x2="740" y2="250" stroke="rgba(15, 23, 42, 0.12)" stroke-dasharray="4 4" />

          <!-- ==============================================================
               SECTION A: 弯头辐条 (J-BEND SPOKE) - 高精等径 CAD 测长投影
               ============================================================== -->
          <text x="35" y="48" fill="#334155" font-size="11" font-weight="900" letter-spacing="0.05em">
            A. 弯头辐条 (J-BEND SPOKE) - 测长基准标定
          </text>

          <!-- 1. 弯头蘑菇圆帽 Head (紧凑真实比例，底在 y=120, 顶在 y=112) -->
          <path d="M 112 120 L 113 116 C 113 112 129 112 129 116 L 130 120 Z" fill="#94a3b8" stroke="#0f172a" stroke-width="1.8" />
          <ellipse cx="121" cy="115" rx="6" ry="2" fill="#cbd5e1" opacity="0.6" />

          <!-- 2. 弯头辐条主体: 严格等径恒厚 10px, 紧凑短颈(仅 10px 垂直长度), 90° 同心圆角平滑过渡 -->
          <!-- 垂直短颈 x从116到126 (y: 120到130), 弯曲圆心(140,130), 内弧R14, 外弧R24, 水平条身y从144到154 -->
          <path d="M 116 120 L 126 120 L 126 130 A 14 14 0 0 0 140 144 L 610 144 L 610 154 L 140 154 A 24 24 0 0 1 116 130 Z"
                fill="#cbd5e1" stroke="#0f172a" stroke-width="1.8" stroke-linejoin="round" />

          <!-- 3. 螺纹段 Rolled Thread (x: 610 到 700, y: 142 到 156, 略粗于条身) -->
          <rect x="610" y="142" width="90" height="14" fill="#fef3c7" stroke="#d97706" stroke-width="1.6" />
          <g stroke="#b45309" stroke-width="1.2">
            <line x1="620" y1="142" x2="617" y2="156" />
            <line x1="630" y1="142" x2="627" y2="156" />
            <line x1="640" y1="142" x2="637" y2="156" />
            <line x1="650" y1="142" x2="647" y2="156" />
            <line x1="660" y1="142" x2="657" y2="156" />
            <line x1="670" y1="142" x2="667" y2="156" />
            <line x1="680" y1="142" x2="677" y2="156" />
            <line x1="690" y1="142" x2="687" y2="156" />
          </g>

          <!-- 4. 游标卡尺内量爪 (精准卡入弯折内角，工作面贴死在 x=126 基准线上) -->
          <g transform="translate(126, 130)">
            <path d="M 0 0 L 14 14 L 14 36 L 0 36 Z" fill="#e2e8f0" stroke="#475569" stroke-width="1.4" />
            <line x1="0" y1="0" x2="0" y2="36" stroke="#059669" stroke-width="2" />
            <line x1="6" y1="24" x2="12" y2="24" stroke="#94a3b8" stroke-width="1" />
            <line x1="6" y1="28" x2="10" y2="28" stroke="#94a3b8" stroke-width="1" />
            <line x1="6" y1="32" x2="12" y2="32" stroke="#94a3b8" stroke-width="1" />
            <text x="18" y="28" fill="#059669" font-size="8" font-weight="900">卡尺内爪贴合面</text>
          </g>

          <!-- 5. 法定基准线与主尺寸线 -->
          <!-- [基准线 1] 弯头内圆弧顶点垂线 (x=126) -->
          <line x1="126" y1="105" x2="126" y2="215" stroke="#059669" stroke-width="1.8" stroke-dasharray="4 2" />
          <!-- [基准线 2] 螺纹端头截止垂线 (x=700) -->
          <line x1="700" y1="85" x2="700" y2="215" stroke="#059669" stroke-width="1.8" stroke-dasharray="4 2" />

          <!-- 【主尺寸线】弯头有效长度 L (双向箭头) -->
          <line x1="126" y1="205" x2="700" y2="205" stroke="#059669" stroke-width="2.5" marker-start="url(#arrow-emerald)" marker-end="url(#arrow-emerald)" />
          <rect x="355" y="194" width="150" height="22" rx="5" fill="#ffffff" stroke="#059669" stroke-width="1.4" />
          <text x="430" y="209" fill="#059669" font-size="10.5" font-weight="900" text-anchor="middle">
            L: 弯头理论全长 (DAT)
          </text>

          <!-- 【螺纹尺寸线】9.5mm 滚压牙标注 -->
          <line x1="610" y1="170" x2="700" y2="170" stroke="#d97706" stroke-width="1.5" marker-start="url(#arrow-amber)" marker-end="url(#arrow-amber)" />
          <text x="655" y="182" fill="#d97706" font-size="8.5" font-weight="800" text-anchor="middle">
            9.5mm 滚压牙
          </text>

          <!-- 6. 避坑警示标注 (彻底错开，绝无文字遮挡) -->
          <!-- 【致命误区 A - 序号 4】条帽顶端测量 -->
          <line x1="121" y1="98" x2="121" y2="112" stroke="#e11d48" stroke-width="1.5" stroke-dasharray="2 2" />
          <path d="M 116 93 L 126 103 M 126 93 L 116 103" stroke="#e11d48" stroke-width="2.2" stroke-linecap="round" />
          <line x1="121" y1="98" x2="126" y2="98" stroke="#e11d48" stroke-width="1.4" />
          <text x="132" y="101" fill="#e11d48" font-size="8.5" font-weight="900">+2.2mm 致命虚长</text>
          <!-- 序号 4 徽章 -->
          <g class="svg-node-highlight" id="svg-badge-4" transform="translate(94, 98)">
            <circle cx="0" cy="0" r="11" fill="#e11d48" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-size="11" font-weight="900" text-anchor="middle">4</text>
          </g>

          <!-- 【致命误区 B - 序号 5】外弯弧测量 -->
          <line x1="116" y1="130" x2="85" y2="160" stroke="#e11d48" stroke-width="1.4" stroke-dasharray="2 2" />
          <circle cx="116" cy="130" r="3.5" fill="#e11d48" />
          <path d="M 80 155 L 90 165 M 90 155 L 80 165" stroke="#e11d48" stroke-width="2" stroke-linecap="round" />
          <text x="75" y="176" fill="#e11d48" font-size="8" font-weight="900" text-anchor="middle">禁以外弧测</text>
          <!-- 序号 5 徽章 -->
          <g class="svg-node-highlight" id="svg-badge-5" transform="translate(52, 160)">
            <circle cx="0" cy="0" r="10" fill="#e11d48" stroke="#ffffff" stroke-width="1.6" />
            <text cx="0" cy="3.5" fill="#ffffff" font-size="10" font-weight="900" text-anchor="middle">5</text>
          </g>

          <!-- 序号 1 徽章 (在内弯顶点右上方，带引线指向切点，绝不遮挡圆弧) -->
          <line x1="155" y1="108" x2="130" y2="128" stroke="#059669" stroke-width="1.5" />
          <circle cx="127" cy="130" r="3" fill="#059669" />
          <g class="svg-node-highlight" id="svg-badge-1" transform="translate(165, 102)">
            <circle cx="0" cy="0" r="11" fill="#059669" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-size="11" font-weight="900" text-anchor="middle">1</text>
          </g>

          <!-- 序号 2 徽章 (螺纹尖端) -->
          <g class="svg-node-highlight" id="svg-badge-2" transform="translate(700, 115)">
            <circle cx="0" cy="0" r="11" fill="#059669" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-size="11" font-weight="900" text-anchor="middle">2</text>
          </g>

          <!-- 序号 3 徽章 (螺纹段) -->
          <g class="svg-node-highlight" id="svg-badge-3" transform="translate(655, 140)">
            <circle cx="0" cy="0" r="10" fill="#d97706" stroke="#ffffff" stroke-width="1.6" />
            <text cx="0" cy="3.5" fill="#ffffff" font-size="10" font-weight="900" text-anchor="middle">3</text>
          </g>


          <!-- ==========================================
               SECTION B: 直头辐条 (STRAIGHTPULL SPOKE) 测长投影
               ========================================== -->
          <text x="35" y="280" fill="#334155" font-size="11" font-weight="900" letter-spacing="0.05em">
            B. 直头辐条 (STRAIGHTPULL SPOKE) - 测长基准标定
          </text>

          <!-- 直头蘑菇圆台 Head (x: 100 到 118, y: 342 到 358, 厚度约 1.8mm) -->
          <path d="M 100 344 C 100 340 106 338 110 338 L 118 345 L 118 355 L 110 362 C 106 362 100 360 100 356 Z" fill="#94a3b8" stroke="#334155" stroke-width="1.8" />

          <!-- 直头承载台阶 (Under-Head Shoulder, x=118 垂直结合面，绿色高亮受力面) -->
          <line x1="118" y1="345" x2="118" y2="355" stroke="#059669" stroke-width="2.5" />

          <!-- 条身 (圆柱段 + 破风扁平段) -->
          <rect x="118" y="347" width="80" height="6" fill="#cbd5e1" stroke="#334155" stroke-width="1.5" />

          <!-- 破风扁平段 Bladed Blade (200 到 600, y: 345 到 355) -->
          <path d="M 198 347 L 210 345 L 590 345 L 602 347 L 602 353 L 590 355 L 210 355 L 198 353 Z" fill="#e2e8f0" stroke="#475569" stroke-width="1.5" />
          <line x1="220" y1="350" x2="580" y2="350" stroke="#94a3b8" stroke-width="1" stroke-dasharray="8 4" />
          <text x="400" y="365" fill="#64748b" font-size="8" font-weight="700" text-anchor="middle">
            AERO BLADED WING (气动破风扁平段 / 纯轴向零弯矩)
          </text>

          <!-- 圆柱过渡段到螺纹 (602 到 610) -->
          <rect x="602" y="347" width="8" height="6" fill="#cbd5e1" stroke="#334155" stroke-width="1.5" />

          <!-- 螺纹段 Thread (610 到 700, y: 343.5 到 356.5) -->
          <rect x="610" y="343.5" width="90" height="13" fill="#fef3c7" stroke="#d97706" stroke-width="1.5" />
          <g stroke="#b45309" stroke-width="1.2">
            <line x1="620" y1="343.5" x2="618" y2="356.5" />
            <line x1="630" y1="343.5" x2="628" y2="356.5" />
            <line x1="640" y1="343.5" x2="638" y2="356.5" />
            <line x1="650" y1="343.5" x2="648" y2="356.5" />
            <line x1="660" y1="343.5" x2="658" y2="356.5" />
            <line x1="670" y1="343.5" x2="668" y2="356.5" />
            <line x1="680" y1="343.5" x2="678" y2="356.5" />
            <line x1="690" y1="343.5" x2="688" y2="356.5" />
          </g>

          <!-- 模拟卡尺外爪 (卡在直头蘑菇头下承载台阶) -->
          <path d="M 118 325 L 118 345 L 110 345 L 110 325 Z" fill="rgba(5, 150, 105, 0.15)" stroke="#059669" stroke-width="1.2" stroke-dasharray="2 2" />
          <text x="114" y="318" fill="#059669" font-size="7.5" font-weight="800" text-anchor="middle">卡尺外爪抵位</text>

          <!-- [基准线 6] 直头头下承载面 (Under-Head Datum) 垂线 x=118 -->
          <line x1="118" y1="330" x2="118" y2="455" stroke="#059669" stroke-width="1.8" stroke-dasharray="4 2" />
          <!-- [基准线 2 贯通] 螺纹端头垂线 x=700 -->
          <line x1="700" y1="330" x2="700" y2="455" stroke="#059669" stroke-width="1.8" stroke-dasharray="4 2" />

          <!-- 【主尺寸线】直头有效长度 L (双向箭头) -->
          <line x1="118" y1="440" x2="700" y2="440" stroke="#059669" stroke-width="2.5" marker-start="url(#arrow-emerald)" marker-end="url(#arrow-emerald)" />
          <rect x="360" y="429" width="135" height="20" rx="4" fill="#ffffff" stroke="#059669" stroke-width="1.2" />
          <text x="427" y="443" fill="#059669" font-size="10.5" font-weight="900" text-anchor="middle">
            L: 直头理论全长 (DAT)
          </text>

          <!-- 【致命误区 C - 序号 7】严禁算入条头厚度 (100 到 118) -->
          <line x1="100" y1="385" x2="118" y2="385" stroke="#e11d48" stroke-width="1.8" marker-start="url(#arrow-rose)" marker-end="url(#arrow-rose)" />
          <path d="M 105 378 L 113 392 M 113 378 L 105 392" stroke="#e11d48" stroke-width="2" />
          <text x="109" y="405" fill="#e11d48" font-size="8" font-weight="900" text-anchor="middle">
            禁计头厚~1.8mm
          </text>

          <!-- 序号徽章节点 (SVG Nodes) -->
          <!-- 序号 6: 直头头下承面 -->
          <g class="svg-node-highlight" id="svg-badge-6" transform="translate(118, 415)">
            <circle cx="0" cy="0" r="11" fill="#059669" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-size="11" font-weight="900" text-anchor="middle">6</text>
          </g>
          <!-- 序号 7: 条头厚度禁算 -->
          <g class="svg-node-highlight" id="svg-badge-7" transform="translate(68, 385)">
            <circle cx="0" cy="0" r="11" fill="#e11d48" stroke="#ffffff" stroke-width="1.8" />
            <text cx="0" cy="4" fill="#ffffff" font-size="11" font-weight="900" text-anchor="middle">7</text>
          </g>
        </svg>
      </div>

      <div class="spoke-head-measurement-guide__legend">
        <div class="spoke-head-measurement-guide__legend-header">
          <span class="spoke-head-measurement-guide__legend-title">{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideLegendTitle', '测量要点') }}</span>
          <span class="spoke-head-measurement-guide__legend-meta">{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideLegendMeta', '按图中编号') }}</span>
        </div>
        <ol class="spoke-head-measurement-guide__legend-list">
          <li class="spoke-head-measurement-guide__legend-item">
            <span class="spoke-head-measurement-guide__number spoke-head-measurement-guide__number--green">1</span>
            <span class="spoke-head-measurement-guide__copy"><strong>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideJDatumTitle', '弯头内侧折点') }}</strong><em>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideDatumTag', '起点基准') }}</em><span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideJDatumDescription', '卡尺贴住 J 弯 90° 内侧喉部的折点，作为有效长度起点。') }}</span></span>
          </li>
          <li class="spoke-head-measurement-guide__legend-item">
            <span class="spoke-head-measurement-guide__number spoke-head-measurement-guide__number--green">2</span>
            <span class="spoke-head-measurement-guide__copy"><strong>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideThreadEndTitle', '螺纹末端端面') }}</strong><em>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideEndTag', '终点基准') }}</em><span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideThreadEndDescription', '从螺纹最末端平齐端面收尺，完整滚压螺纹包含在长度内。') }}</span></span>
          </li>
          <li class="spoke-head-measurement-guide__legend-item">
            <span class="spoke-head-measurement-guide__number spoke-head-measurement-guide__number--amber">3</span>
            <span class="spoke-head-measurement-guide__copy"><strong>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideThreadSectionTitle', '有效螺纹段') }}</strong><em>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideLengthTag', '长度范围') }}</em><span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideThreadSectionDescription', '螺纹是辐条本体的一部分，不能因为有条帽而提前结束测量。') }}</span></span>
          </li>
          <li class="spoke-head-measurement-guide__legend-item spoke-head-measurement-guide__legend-item--warning">
            <span class="spoke-head-measurement-guide__number spoke-head-measurement-guide__number--rose">4</span>
            <span class="spoke-head-measurement-guide__copy"><strong>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideJCapWarningTitle', '弯头外沿误差') }}</strong><em>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideWarningTag', '不要这样量') }}</em><span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideJCapWarningDescription', '从蘑菇头外沿开始会把弯折厚度带入，结果会偏长。') }}</span></span>
          </li>
          <li class="spoke-head-measurement-guide__legend-item spoke-head-measurement-guide__legend-item--warning">
            <span class="spoke-head-measurement-guide__number spoke-head-measurement-guide__number--rose">5</span>
            <span class="spoke-head-measurement-guide__copy"><strong>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideJOuterArcWarningTitle', '外弧不是基准') }}</strong><em>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideWarningTag', '不要这样量') }}</em><span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideJOuterArcWarningDescription', '冷镦外弧会有形状公差，外侧切线不能稳定复现长度。') }}</span></span>
          </li>
          <li class="spoke-head-measurement-guide__legend-item">
            <span class="spoke-head-measurement-guide__number spoke-head-measurement-guide__number--green">6</span>
            <span class="spoke-head-measurement-guide__copy"><strong>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideStraightDatumTitle', '直头头下承载台阶') }}</strong><em>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideDatumTag', '起点基准') }}</em><span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideStraightDatumDescription', '从条头下方与花鼓接触的承载肩面开始，量到螺纹末端。') }}</span></span>
          </li>
          <li class="spoke-head-measurement-guide__legend-item spoke-head-measurement-guide__legend-item--warning">
            <span class="spoke-head-measurement-guide__number spoke-head-measurement-guide__number--rose">7</span>
            <span class="spoke-head-measurement-guide__copy"><strong>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideStraightHeadWarningTitle', '直头条头厚度') }}</strong><em>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideExcludeTag', '不计入长度') }}</em><span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.guideStraightHeadWarningDescription', '条头落在花鼓沉孔内，只量承载面之后的有效部分。') }}</span></span>
          </li>
        </ol>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'

const { t } = useI18n()
</script>

<style scoped>
.spoke-head-measurement-guide {
  width: 100%;
  margin-top: 18px;
  padding: 18px;
  border: 1px solid rgba(15, 23, 42, 0.12);
  border-radius: 20px;
  background: #ffffff;
  color: #0f172a;
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.05);
}

.spoke-head-measurement-guide__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.spoke-head-measurement-guide__kicker {
  margin: 0 0 3px;
  color: #059669;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.spoke-head-measurement-guide__title {
  margin: 0;
  color: #0f172a;
  font-size: 15px;
  font-weight: 900;
  line-height: 1.35;
}

.spoke-head-measurement-guide__badge,
.spoke-head-measurement-guide__legend-meta {
  display: inline-flex;
  align-items: center;
  border: 1px solid rgba(15, 23, 42, 0.12);
  border-radius: 9999px;
  color: #64748b;
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.06em;
  white-space: nowrap;
}

.spoke-head-measurement-guide__badge {
  padding: 5px 9px;
  background: rgba(5, 150, 105, 0.08);
  border-color: rgba(5, 150, 105, 0.2);
  color: #047857;
}

.spoke-head-measurement-guide__grid {
  display: grid;
  grid-template-columns: minmax(0, 1.25fr) minmax(280px, 0.75fr);
  gap: 14px;
  align-items: stretch;
}

.spoke-head-measurement-guide__diagram,
.spoke-head-measurement-guide__legend {
  min-width: 0;
  border: 1px solid rgba(15, 23, 42, 0.09);
  border-radius: 16px;
  background: #f8fafc;
}

.spoke-head-measurement-guide__diagram {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  padding: 8px;
}

.spoke-head-measurement-guide__diagram svg {
  display: block;
  width: 100%;
  height: auto;
  max-height: 520px;
  font-family: inherit;
}

.spoke-head-measurement-guide__diagram svg text {
  font-family: inherit;
}

.spoke-head-measurement-guide__legend {
  display: flex;
  flex-direction: column;
  padding: 13px;
  overflow: hidden;
}

.spoke-head-measurement-guide__legend-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding-bottom: 9px;
  border-bottom: 1px dashed rgba(15, 23, 42, 0.12);
}

.spoke-head-measurement-guide__legend-title {
  color: #0f172a;
  font-size: 12px;
  font-weight: 900;
}

.spoke-head-measurement-guide__legend-meta {
  padding: 3px 7px;
  font-size: 8px;
}

.spoke-head-measurement-guide__legend-list {
  display: grid;
  gap: 7px;
  margin: 10px 0 0;
  padding: 0;
  list-style: none;
}

.spoke-head-measurement-guide__legend-item {
  display: grid;
  grid-template-columns: 23px minmax(0, 1fr);
  gap: 8px;
  align-items: start;
  min-width: 0;
}

.spoke-head-measurement-guide__number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 23px;
  height: 23px;
  border-radius: 9999px;
  color: #ffffff;
  font-size: 10px;
  font-weight: 900;
}

.spoke-head-measurement-guide__number--green {
  background: #059669;
}

.spoke-head-measurement-guide__number--amber {
  background: #d97706;
}

.spoke-head-measurement-guide__number--rose {
  background: #e11d48;
}

.spoke-head-measurement-guide__copy {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 2px 6px;
  min-width: 0;
  color: #475569;
  font-size: 10.5px;
  line-height: 1.4;
}

.spoke-head-measurement-guide__copy strong {
  min-width: 0;
  color: #0f172a;
  font-size: 11px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-head-measurement-guide__copy em {
  align-self: start;
  padding: 1px 5px;
  border: 1px solid rgba(5, 150, 105, 0.2);
  border-radius: 9999px;
  background: rgba(5, 150, 105, 0.08);
  color: #047857;
  font-size: 8px;
  font-style: normal;
  font-weight: 800;
  white-space: nowrap;
}

.spoke-head-measurement-guide__copy > span {
  grid-column: 1 / -1;
}

.spoke-head-measurement-guide__legend-item--warning .spoke-head-measurement-guide__copy em {
  border-color: rgba(225, 29, 72, 0.2);
  background: rgba(225, 29, 72, 0.08);
  color: #be123c;
}

@media (max-width: 960px) {
  .spoke-head-measurement-guide__grid {
    grid-template-columns: 1fr;
  }

  .spoke-head-measurement-guide__diagram svg {
    max-height: none;
  }
}

@media (max-width: 767px) {
  .spoke-head-measurement-guide {
    padding: 14px;
    border-radius: 18px;
  }

  .spoke-head-measurement-guide__header {
    align-items: flex-start;
    flex-direction: column;
  }

  .spoke-head-measurement-guide__copy {
    grid-template-columns: 1fr;
  }

  .spoke-head-measurement-guide__copy em {
    justify-self: start;
  }
}
</style>
