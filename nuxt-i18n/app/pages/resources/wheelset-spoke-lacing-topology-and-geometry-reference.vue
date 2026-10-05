<template>
<div class="spoke-lacing-page">
<!-- ================= 顶部工程页眉 ================= -->
  <header class="blueprint-header">
    <div class="header-title-group">
      <h1>
        <span>{{ t('wheelsetLacingTopology.title') }}</span>
        <span class="header-badge badge-primary">{{ t('wheelsetLacingTopology.versionBadge') }}</span>
        <span class="header-badge">{{ t('wheelsetLacingTopology.topologyTag') }}</span>
      </h1>
      <div class="header-subtitle">
        {{ t('wheelsetLacingTopology.subtitle') }}
      </div>
    </div>
    <div class="header-controls">
      <span class="header-badge" style="background: rgba(5, 150, 105, 0.06); color: #047857; border-color: rgba(5, 150, 105, 0.2);">
        {{ t('wheelsetLacingTopology.standaloneBadge') }}
      </span>
    </div>
  </header>

  <!-- ================= 核心工作台主网格 ================= -->
  <div class="blueprint-grid" :aria-label="t('wheelsetLacingTopology.workbenchLabel')">

    <!-- ========== 左侧列：参数选择器与图层控制 ========== -->
    <section class="uds-card controls-column">
      <div class="card-header">
        <h2 class="card-title">
            <span>{{ t('wheelsetLacingTopology.controls.title') }}</span>
        </h2>
          <span class="card-tag">{{ t('wheelsetLacingTopology.controls.tag') }}</span>
      </div>

      <!-- 1. 轮组总孔数选择 (Hole Count) -->
      <div>
        <div class="selector-group-label">
          <span>{{ t('wheelsetLacingTopology.controls.holeLabel') }}</span>
          <span id="current-hole-label" style="font-family: var(--tz-font-ui); color: var(--accent-primary);">{{ t('wheelsetLacingTopology.holes.24') }} (1:1)</span>
        </div>
        <div class="hole-pill-grid">
          <button type="button" class="uds-pill-btn" data-holes="16" aria-pressed="false" @click="setHoleCount(16)">
            <span>{{ t('wheelsetLacingTopology.holes.16') }}</span>
          </button>
          <button type="button" class="uds-pill-btn" data-holes="20" aria-pressed="false" @click="setHoleCount(20)">
            <span>{{ t('wheelsetLacingTopology.holes.20') }}</span>
          </button>
          <button type="button" class="uds-pill-btn" data-holes="21" aria-pressed="false" @click="setHoleCount(21)">
            <span>{{ t('wheelsetLacingTopology.holes.21') }}</span>
            <span class="sub-text">{{ t('wheelsetLacingTopology.holes.21Sub') }}</span>
          </button>
          <button type="button" class="uds-pill-btn active" data-holes="24" aria-pressed="true" @click="setHoleCount(24)">
            <span>{{ t('wheelsetLacingTopology.holes.24') }}</span>
          </button>
          <button type="button" class="uds-pill-btn" data-holes="28" aria-pressed="false" @click="setHoleCount(28)">
            <span>{{ t('wheelsetLacingTopology.holes.28') }}</span>
          </button>
          <button type="button" class="uds-pill-btn" data-holes="32" aria-pressed="false" @click="setHoleCount(32)">
            <span>{{ t('wheelsetLacingTopology.holes.32') }}</span>
          </button>
          <button type="button" class="uds-pill-btn" data-holes="36" aria-pressed="false" @click="setHoleCount(36)">
            <span>{{ t('wheelsetLacingTopology.holes.36') }}</span>
          </button>
          <button type="button" class="uds-pill-btn" data-holes="24_2to1" aria-pressed="false" @click="setHoleCount('24_2to1')">
            <span>{{ t('wheelsetLacingTopology.holes.24_2to1') }}</span>
            <span class="sub-text">{{ t('wheelsetLacingTopology.holes.24_2to1Sub') }}</span>
          </button>
        </div>
      </div>

      <!-- 2. 编法交叉数选择 (Cross Pattern) -->
      <div>
        <div class="selector-group-label">
          <span>{{ t('wheelsetLacingTopology.controls.crossLabel') }}</span>
          <span id="current-cross-label" style="font-family: var(--tz-font-ui); color: var(--accent-primary);">{{ t('wheelsetLacingTopology.cross.2') }}</span>
        </div>
        <div class="cross-pill-grid" id="cross-btn-container">
          <button type="button" class="uds-pill-btn" data-cross="0" aria-pressed="false" @click="setCrossCount(0)">
            <span>0X</span>
            <span class="sub-text">{{ t('wheelsetLacingTopology.cross.radial') }}</span>
          </button>
          <button type="button" class="uds-pill-btn" data-cross="1" aria-pressed="false" @click="setCrossCount(1)">
            <span>1X</span>
            <span class="sub-text">{{ t('wheelsetLacingTopology.cross.one') }}</span>
          </button>
          <button type="button" class="uds-pill-btn active" data-cross="2" aria-pressed="true" @click="setCrossCount(2)">
            <span>2X</span>
            <span class="sub-text">{{ t('wheelsetLacingTopology.cross.two') }}</span>
          </button>
          <button type="button" class="uds-pill-btn" data-cross="3" aria-pressed="false" @click="setCrossCount(3)">
            <span>3X</span>
            <span class="sub-text">{{ t('wheelsetLacingTopology.cross.three') }}</span>
          </button>
          <button type="button" class="uds-pill-btn" data-cross="4" aria-pressed="false" @click="setCrossCount(4)">
            <span>4X</span>
            <span class="sub-text">{{ t('wheelsetLacingTopology.cross.four') }}</span>
          </button>
        </div>
      </div>

      <!-- 3. 双法兰几何参数；径向值只控制主视图，轴向值控制下方剖面图 -->
      <div class="flange-geometry-control-group">
        <div class="selector-group-label">
          <span>{{ t('wheelsetLacingTopology.controls.flangeGeometryLabel') }}</span>
          <span class="control-unit-label">{{ t('wheelsetLacingTopology.controls.displayAndMm') }}</span>
        </div>
        <div class="flange-geometry-grid">
          <label class="flange-geometry-field">
            <span>{{ t('wheelsetLacingTopology.controls.driveFlangeRadius') }}</span>
            <span class="flange-geometry-input-line">
              <input id="flange-radius-a-input" type="number" :min="MINIMUM_FLANGE_DISPLAY_RADIUS" :max="MAXIMUM_FLANGE_DISPLAY_RADIUS" step="1" :value="state.flangeRadiusA" @input="updateWheelsetLacingGeometryInput('flangeRadiusA', $event)" @change="normalizeWheelsetLacingGeometryInput('flangeRadiusA', $event)">
              <span>SVG</span>
            </span>
          </label>
          <label class="flange-geometry-field">
            <span>{{ t('wheelsetLacingTopology.controls.nonDriveFlangeRadius') }}</span>
            <span class="flange-geometry-input-line">
              <input id="flange-radius-b-input" type="number" :min="MINIMUM_FLANGE_DISPLAY_RADIUS" :max="MAXIMUM_FLANGE_DISPLAY_RADIUS" step="1" :value="state.flangeRadiusB" @input="updateWheelsetLacingGeometryInput('flangeRadiusB', $event)" @change="normalizeWheelsetLacingGeometryInput('flangeRadiusB', $event)">
              <span>SVG</span>
            </span>
          </label>
          <label class="flange-geometry-field">
            <span>{{ t('wheelsetLacingTopology.controls.driveFlangeOffset') }}</span>
            <span class="flange-geometry-input-line">
              <input id="flange-offset-a-input" type="number" :min="MINIMUM_FLANGE_OFFSET_MM" :max="MAXIMUM_FLANGE_OFFSET_MM" step="0.5" :value="state.flangeOffsetAMm" @input="updateWheelsetLacingGeometryInput('flangeOffsetAMm', $event)" @change="normalizeWheelsetLacingGeometryInput('flangeOffsetAMm', $event)">
              <span>mm</span>
            </span>
          </label>
          <label class="flange-geometry-field">
            <span>{{ t('wheelsetLacingTopology.controls.nonDriveFlangeOffset') }}</span>
            <span class="flange-geometry-input-line">
              <input id="flange-offset-b-input" type="number" :min="MINIMUM_FLANGE_OFFSET_MM" :max="MAXIMUM_FLANGE_OFFSET_MM" step="0.5" :value="state.flangeOffsetBMm" @input="updateWheelsetLacingGeometryInput('flangeOffsetBMm', $event)" @change="normalizeWheelsetLacingGeometryInput('flangeOffsetBMm', $event)">
              <span>mm</span>
            </span>
          </label>
        </div>
        <p class="flange-geometry-help">{{ t('wheelsetLacingTopology.controls.flangeGeometryHelp') }}</p>
      </div>

      <!-- G3 仅在 21H 选择时显示；A-B-A 三孔组由两段独立角间距定义 -->
      <div id="g3-geometry-control-group" class="flange-geometry-control-group g3-geometry-control-group" hidden aria-hidden="true">
        <div class="selector-group-label">
          <span>{{ t('wheelsetLacingTopology.controls.g3GroupSpacingLabel') }}</span>
          <span class="control-unit-label">{{ t('wheelsetLacingTopology.controls.displayDegrees') }}</span>
        </div>
        <div class="flange-geometry-grid">
          <label class="flange-geometry-field">
            <span>{{ t('wheelsetLacingTopology.controls.g3SpacingAToB') }}</span>
            <span class="flange-geometry-input-line">
              <input id="g3-spacing-a-to-b-input" type="number" :min="MINIMUM_G3_RIM_HOLE_SPACING_DEGREES" :max="MAXIMUM_G3_RIM_HOLE_SPACING_DEGREES" step="0.1" :value="state.g3RimHoleSpacingAToBDegrees" @input="updateWheelsetLacingGeometryInput('g3RimHoleSpacingAToBDegrees', $event)" @change="normalizeWheelsetLacingGeometryInput('g3RimHoleSpacingAToBDegrees', $event)">
              <span>°</span>
            </span>
          </label>
          <label class="flange-geometry-field">
            <span>{{ t('wheelsetLacingTopology.controls.g3SpacingBToA') }}</span>
            <span class="flange-geometry-input-line">
              <input id="g3-spacing-b-to-a-input" type="number" :min="MINIMUM_G3_RIM_HOLE_SPACING_DEGREES" :max="MAXIMUM_G3_RIM_HOLE_SPACING_DEGREES" step="0.1" :value="state.g3RimHoleSpacingBToADegrees" @input="updateWheelsetLacingGeometryInput('g3RimHoleSpacingBToADegrees', $event)" @change="normalizeWheelsetLacingGeometryInput('g3RimHoleSpacingBToADegrees', $event)">
              <span>°</span>
            </span>
          </label>
        </div>
        <p class="flange-geometry-help">{{ t('wheelsetLacingTopology.controls.g3GroupSpacingHelp') }}</p>
      </div>

      <!-- 4. 轮侧视图模式切换 (View Mode) -->
      <div>
        <div class="selector-group-label">
          <span>{{ t('wheelsetLacingTopology.controls.viewLabel') }}</span>
        </div>
        <div class="view-mode-grid">
          <button type="button" class="view-pill-btn" id="btn-view-drive" aria-pressed="false" @click="setViewMode('drive')">
            {{ t('wheelsetLacingTopology.controls.driveView') }}
          </button>
          <button type="button" class="view-pill-btn active" id="btn-view-both" aria-pressed="true" @click="setViewMode('both')">
            {{ t('wheelsetLacingTopology.controls.bothView') }}
          </button>
          <button type="button" class="view-pill-btn" id="btn-view-nondrive" aria-pressed="false" @click="setViewMode('nondrive')">
            {{ t('wheelsetLacingTopology.controls.nonDriveView') }}
          </button>
        </div>
      </div>

      <!-- 5. 图层显示开关 -->
      <div class="layer-toggle-group">
        <span class="selector-group-label" style="margin-bottom: 4px;">{{ t('wheelsetLacingTopology.controls.layersLabel') }}</span>
        
        <label class="layer-item">
          <span><span class="layer-indicator" style="background: var(--accent-primary);"></span>{{ t('wheelsetLacingTopology.controls.leading') }}</span>
          <input type="checkbox" id="toggle-leading" checked @change="updateCanvasLayers">
        </label>
        
        <label class="layer-item">
          <span><span class="layer-indicator" style="background: var(--accent-amber);"></span>{{ t('wheelsetLacingTopology.controls.trailing') }}</span>
          <input type="checkbox" id="toggle-trailing" checked @change="updateCanvasLayers">
        </label>

        <label class="layer-item">
          <span><span class="layer-indicator" style="background: var(--accent-sky);"></span>{{ t('wheelsetLacingTopology.controls.nonDrive') }}</span>
          <input type="checkbox" id="toggle-nondrive" checked @change="updateCanvasLayers">
        </label>

      </div>

    </section>

    <!-- ========== 中央列：高精交互 CAD 矢量轮组画布 ========== -->
    <section class="uds-card canvas-card">
      <div class="canvas-toolbar">
        <div class="card-title">
          <span>{{ t('wheelsetLacingTopology.canvas.title') }}</span>
          <span class="card-tag" id="canvas-status-tag">{{ t('wheelsetLacingTopology.canvas.tag') }}</span>
        </div>
        <div class="canvas-legend-inline">
          <span style="color: var(--accent-primary);">{{ t('wheelsetLacingTopology.canvas.driveLeading') }}</span>
          <span style="color: var(--accent-amber);">{{ t('wheelsetLacingTopology.canvas.driveTrailing') }}</span>
          <span style="color: var(--accent-sky);">{{ t('wheelsetLacingTopology.canvas.nonDriveSpokes') }}</span>
        </div>
      </div>

      <!-- 矢量画布 -->
      <div class="svg-viewport" id="viewport-container">
        <svg
          id="spoke-lacing-svg"
          viewBox="-300 -300 600 600"
          preserveAspectRatio="xMidYMid meet"
          role="img"
          :aria-label="t('wheelsetLacingTopology.canvas.aria')"
        >
          <!-- 动态渲染 SVG 元素 -->
        </svg>
      </div>
      <div class="flange-profile-viewport" id="flange-profile-container">
        <div class="flange-profile-toolbar">
          <div class="card-title">
            <span>{{ t('wheelsetLacingTopology.canvas.profileTitle') }}</span>
            <span class="card-tag" id="flange-profile-status-tag">{{ t('wheelsetLacingTopology.canvas.profileTag') }}</span>
          </div>
          <span class="flange-profile-summary" id="flange-profile-summary">{{ t('wheelsetLacingTopology.canvas.profilePending') }}</span>
        </div>
        <svg
          id="flange-profile-svg"
          viewBox="-220 -90 440 180"
          preserveAspectRatio="xMidYMid meet"
          role="img"
          :aria-label="t('wheelsetLacingTopology.canvas.profileAria')"
        >
          <!-- 动态渲染双法兰轴向剖面 -->
        </svg>
      </div>
    </section>

    <!-- ========== 右侧列：实时几何投影与拓扑规则 ========== -->
    <section class="geometry-metrics-column">

       <!-- 实时几何指标卡片 -->
      <div class="uds-card">
        <div class="card-header">
          <h2 class="card-title">
            <span>{{ t('wheelsetLacingTopology.telemetry.title') }}</span>
          </h2>
          <span class="card-tag">{{ t('wheelsetLacingTopology.telemetry.tag') }}</span>
        </div>

        <!-- 指标 1：出条切线角 α -->
        <div class="telemetry-metric-tile">
          <div class="metric-top">
            <span class="metric-label">{{ t('wheelsetLacingTopology.telemetry.tangentLabel') }}</span>
            <span class="metric-unit">{{ t('wheelsetLacingTopology.telemetry.projectionAngle') }}</span>
          </div>
          <div class="metric-value">
            <span id="metric-tangential-projection-angle">{{ formatBackendDisplayGeometryMetric(backendDisplayGeometry?.metrics.aggregate_mean_absolute_projection_angle_degrees) }}</span>
            <span class="metric-unit">DEG (°)</span>
          </div>
          <div class="metric-bar">
            <div
              id="bar-tangential-projection-angle"
              class="metric-bar-fill"
              :style="{ width: `${formatBackendDisplayGeometryBarWidth(backendDisplayGeometry?.metrics.aggregate_mean_absolute_projection_angle_degrees, 90)}%`, backgroundColor: 'var(--accent-primary)' }"
            ></div>
          </div>
          <p id="metric-tangential-projection-angle-desc" class="metric-desc">{{ t('wheelsetLacingTopology.telemetry.tangentDesc') }}</p>
        </div>

        <!-- 指标 2：切向几何投影 (sin α) -->
        <div class="telemetry-metric-tile">
          <div class="metric-top">
            <span class="metric-label">{{ t('wheelsetLacingTopology.telemetry.tangentialLabel') }}</span>
            <span class="metric-unit">{{ t('wheelsetLacingTopology.telemetry.tangentialUnit') }}</span>
          </div>
          <div class="metric-value">
            <span id="metric-tangential-projection">{{ formatBackendDisplayGeometryMetric(backendDisplayGeometry?.metrics.mean_absolute_tangential_projection_percent) }}</span>
            <span class="metric-unit">%</span>
          </div>
          <div class="metric-bar">
            <div
              id="bar-tangential-projection"
              class="metric-bar-fill"
              :style="{ width: `${formatBackendDisplayGeometryBarWidth(backendDisplayGeometry?.metrics.mean_absolute_tangential_projection_percent, 100)}%`, backgroundColor: 'var(--accent-primary)' }"
            ></div>
          </div>
          <p id="metric-tangential-projection-desc" class="metric-desc">{{ t('wheelsetLacingTopology.telemetry.tangentialDesc') }}</p>
        </div>

        <!-- 指标 3：径向几何投影 (cos α) -->
        <div class="telemetry-metric-tile">
          <div class="metric-top">
            <span class="metric-label">{{ t('wheelsetLacingTopology.telemetry.radialLabel') }}</span>
            <span class="metric-unit">{{ t('wheelsetLacingTopology.telemetry.radialUnit') }}</span>
          </div>
          <div class="metric-value">
            <span id="metric-radial-projection">{{ formatBackendDisplayGeometryMetric(backendDisplayGeometry?.metrics.mean_absolute_radial_projection_percent) }}</span>
            <span class="metric-unit">%</span>
          </div>
          <div class="metric-bar">
            <div
              id="bar-radial-projection"
              class="metric-bar-fill"
              :style="{ width: `${formatBackendDisplayGeometryBarWidth(backendDisplayGeometry?.metrics.mean_absolute_radial_projection_percent, 100)}%`, backgroundColor: 'var(--accent-steel)' }"
            ></div>
          </div>
          <p id="metric-radial-projection-desc" class="metric-desc">{{ t('wheelsetLacingTopology.telemetry.radialDesc') }}</p>
        </div>

      </div>

      <!-- 法兰孔几何规则提示 -->
      <div class="uds-card">
        <div class="card-header">
          <h2 class="card-title">
            <span>{{ t('wheelsetLacingTopology.review.title') }}</span>
          </h2>
          <span class="card-tag">{{ t('wheelsetLacingTopology.review.tag') }}</span>
        </div>

        <div id="topology-status-box" class="topology-status-card topology-status-preview">
          <div class="status-header-line">
           <span id="status-title-text">{{ t('wheelsetLacingTopology.review.previewTitle') }}</span>
          </div>
          <p id="status-detail-text" class="status-detail-text">
            {{ t('wheelsetLacingTopology.review.generalTip') }}
          </p>
        </div>

        <div style="background: var(--bg-inset); border-radius: 14px; padding: 12px; font-size: 9.5px; color: var(--text-secondary); line-height: 1.45;">
          <strong style="color: var(--text-main); display: block; margin-bottom: 4px;">{{ t('wheelsetLacingTopology.review.noteLabel') }}</strong>
          <span id="builder-tip-text">
            {{ t('wheelsetLacingTopology.review.noteTip') }}
          </span>
        </div>
      </div>

    </section>

  </div>

  <!-- ================= SSR 可读的拓扑事实摘要 ================= -->
  <section class="uds-card geo-fact-summary" aria-labelledby="geo-fact-title">
    <div class="card-header">
      <h2 id="geo-fact-title" class="card-title">
        <span>{{ t('wheelsetLacingTopology.facts.title') }}</span>
      </h2>
      <span class="card-tag">{{ t('wheelsetLacingTopology.facts.tag') }}</span>
    </div>
    <p class="geo-fact-summary__intro">
      {{ t('wheelsetLacingTopology.facts.intro') }}
    </p>
    <dl class="geo-fact-summary__grid">
      <div>
        <dt>{{ t('wheelsetLacingTopology.facts.topologyScope') }}</dt>
        <dd>{{ t('wheelsetLacingTopology.facts.topologyScopeValue') }}</dd>
      </div>
      <div>
        <dt>{{ t('wheelsetLacingTopology.facts.units') }}</dt>
        <dd>{{ t('wheelsetLacingTopology.facts.unitsValue') }}</dd>
      </div>
      <div>
        <dt>{{ t('wheelsetLacingTopology.facts.displayCoordinates') }}</dt>
        <dd>{{ t('wheelsetLacingTopology.facts.displayCoordinatesValue') }}</dd>
      </div>
      <div>
        <dt>{{ t('wheelsetLacingTopology.facts.boundary') }}</dt>
        <dd>{{ t('wheelsetLacingTopology.facts.boundaryValue') }}</dd>
      </div>
    </dl>
  </section>

  <!-- ================= 底部拓扑规则说明 ================= -->
  <footer class="cad-legend-section">
    <div class="card-header" style="padding-bottom: 8px;">
      <h3 class="card-title">
        <span>{{ t('wheelsetLacingTopology.notes.title') }}</span>
      </h3>
      <span class="header-badge">{{ t('wheelsetLacingTopology.notes.tag') }}</span>
    </div>

    <div class="legend-table-grid">
      <div class="legend-box">
        <div class="legend-box-title">
          <span class="legend-box-num">01</span>
          <span>{{ t('wheelsetLacingTopology.notes.radialTitle') }}</span>
        </div>
        <div class="legend-box-body">
            {{ t('wheelsetLacingTopology.notes.radialBody') }}
        </div>
      </div>

      <div class="legend-box">
        <div class="legend-box-title">
          <span class="legend-box-num">02</span>
          <span>{{ t('wheelsetLacingTopology.notes.mappingTitle') }}</span>
        </div>
        <div class="legend-box-body">
            {{ t('wheelsetLacingTopology.notes.mappingBody') }}
        </div>
      </div>

      <div class="legend-box">
        <div class="legend-box-title">
          <span class="legend-box-num">03</span>
          <span>{{ t('wheelsetLacingTopology.notes.asymmetricTitle') }}</span>
        </div>
        <div class="legend-box-body">
            {{ t('wheelsetLacingTopology.notes.asymmetricBody') }}
        </div>
      </div>
    </div>
  </footer>

  <!-- ================= 核心计算与完整拓扑动态 SVG 渲染引擎 ================= -->
  
</div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useAsyncData, useI18n, useSwitchLocalePath } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import { usePageMessages } from '~/composables/usePageMessages'
import {
  useStorefrontSeoLinks,
  useStorefrontSeoRouteOverride,
} from '~/composables/seo/useStorefrontSeoLinks'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import { fetchFaqDataByRoutePath } from '~/data/faq'
import localeManifest from '~/i18n/locales.manifest'
import { getSupportedWheelsetLacingCrossCounts } from '~/utils/wheelsetLacingSelectionContract'
import {
  resolveWheelsetLacingDisplayGeometryTopologySelection,
  validateWheelsetLacingDisplayGeometryResponse,
  WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT,
} from '~/utils/wheelsetLacingDisplayGeometryContract'

definePageMeta({
  layout: 'products',
  footer: false,
  breadcrumbLabelKey: 'wheelsetLacingTopology.breadcrumb',
  breadcrumbLabelFallback: 'Wheelset lacing topology',
})

const { locale, t } = useI18n()
const { request } = useApiRequest()
const switchLocalePath = useSwitchLocalePath()
const { loadPageMessages } = usePageMessages('wheelsetLacingTopology')
const pageMessagesVersion = ref(0)
let isInteractiveBlueprintMounted = false

const wheelsetLacingFaqRoutePath = '/resources/wheelset-spoke-lacing-topology-and-geometry-reference'
const { data: wheelsetLacingFaqData } = await useAsyncData(
  'wheelset-lacing-topology-faq-structured-data',
  () => fetchFaqDataByRoutePath(wheelsetLacingFaqRoutePath),
  { watch: [locale], default: () => null },
)

const serverRenderedWheelsetLacingDisplayGeometryTopologyIdentifier = '24h-symmetric-1to1-2x'
const serverRenderedWheelsetLacingDisplayGeometrySelection = resolveWheelsetLacingDisplayGeometryTopologySelection(24, 2)
const { data: serverRenderedWheelsetLacingDisplayGeometry } = await useAsyncData(
  'wheelset-lacing-default-display-geometry-v1-4',
  async () => {
    try {
      const response = await request('/wheelset-lacing/display-geometry', {
        method: 'GET',
        query: {
          topology_id: serverRenderedWheelsetLacingDisplayGeometryTopologyIdentifier,
        },
      })
      return response?.data ?? null
    } catch {
      // The page still exposes the static topology facts if the optional
      // engineering API is temporarily unavailable during SSR. The client
      // retry below will request the same canonical GET contract after mount.
      return null
    }
  },
  { default: () => null },
)

const displayGeometryError = ref(null)
const backendDisplayGeometry = ref(null)
const initialServerRenderedDisplayGeometry = serverRenderedWheelsetLacingDisplayGeometry.value
if (initialServerRenderedDisplayGeometry) {
  try {
    validateWheelsetLacingDisplayGeometryResponse(initialServerRenderedDisplayGeometry, serverRenderedWheelsetLacingDisplayGeometrySelection)
    backendDisplayGeometry.value = initialServerRenderedDisplayGeometry
  } catch (error) {
    displayGeometryError.value = error instanceof Error ? error.message : 'invalid server-rendered display geometry'
  }
}
let displayGeometryRequestSequence = 0
let displayGeometryController = null

await loadPageMessages(locale.value)
watch(locale, async nextLocale => {
  await loadPageMessages(nextLocale)
  // useHead can reevaluate immediately after the locale changes, before the
  // lazy page namespace has merged. Touch a local reactive version after the
  // merge so title, description, and JSON-LD never remain as translation keys.
  pageMessagesVersion.value += 1
  if (!isInteractiveBlueprintMounted) return
  await nextTick()
  renderControls()
  renderBlueprint()
})

const supportedSeoLocaleCodes = new Set(['en', 'zh_cn'])
const localizedSeoRoutes = computed(() => localeManifest
  .filter(entry => supportedSeoLocaleCodes.has(entry.code))
  .map(({ code }) => ({
    code,
    path: switchLocalePath(code) || '/resources/wheelset-spoke-lacing-topology-and-geometry-reference',
  })))

const isPublicSeoLocale = computed(() => supportedSeoLocaleCodes.has(locale.value))
const { canonicalUrl } = useStorefrontSeoLinks()
useStorefrontSeoRouteOverride(localizedSeoRoutes)

const topologyFacts = () => [
  {
    '@type': 'PropertyValue',
    name: t('wheelsetLacingTopology.seo.factStandardName'),
    value: t('wheelsetLacingTopology.seo.standardFactValue'),
  },
  {
    '@type': 'PropertyValue',
    name: t('wheelsetLacingTopology.seo.factG3Name'),
    value: t('wheelsetLacingTopology.review.preview21Detail'),
  },
  {
    '@type': 'PropertyValue',
    name: t('wheelsetLacingTopology.seo.factUniformName'),
    value: t('wheelsetLacingTopology.review.uniformTip'),
  },
  {
    '@type': 'PropertyValue',
    name: t('wheelsetLacingTopology.facts.units'),
    value: t('wheelsetLacingTopology.facts.unitsValue'),
  },
]

const wheelsetLacingFaqSchema = () => {
  const items = wheelsetLacingFaqData.value?.items || []
  if (items.length === 0) return null

  return {
    '@type': 'FAQPage',
    '@id': `${canonicalUrl.value}#faq`,
    url: `${canonicalUrl.value}#faq`,
    mainEntity: items.map(item => ({
      '@type': 'Question',
      name: item.question,
      acceptedAnswer: {
        '@type': 'Answer',
        text: item.answer,
      },
    })),
  }
}

const topologySchema = () => ({
  '@context': 'https://schema.org',
  '@graph': [
    {
      '@type': 'TechArticle',
      '@id': `${canonicalUrl.value}#topology-reference`,
      url: canonicalUrl.value,
      mainEntityOfPage: canonicalUrl.value,
      headline: t('wheelsetLacingTopology.seo.headline'),
      description: t('wheelsetLacingTopology.seo.description'),
      inLanguage: localeManifest.find(entry => entry.code === locale.value)?.iso || locale.value,
      articleSection: t('wheelsetLacingTopology.seo.articleSection'),
      keywords: t('wheelsetLacingTopology.seo.keywords'),
      isAccessibleForFree: true,
      about: {
        '@type': 'Thing',
        name: t('wheelsetLacingTopology.seo.about'),
      },
      hasPart: {
        '@id': `${canonicalUrl.value}#topology-facts`,
      },
    },
    {
      '@type': 'Dataset',
      '@id': `${canonicalUrl.value}#topology-facts`,
      url: `${canonicalUrl.value}#geo-fact-title`,
      name: t('wheelsetLacingTopology.seo.datasetName'),
      description: t('wheelsetLacingTopology.seo.datasetDescription'),
      inLanguage: localeManifest.find(entry => entry.code === locale.value)?.iso || locale.value,
      isAccessibleForFree: true,
      measurementTechnique: t('wheelsetLacingTopology.seo.measurementTechnique'),
      variableMeasured: [
        t('wheelsetLacingTopology.seo.variableHoleMapping'),
        t('wheelsetLacingTopology.seo.variableProjection'),
      ],
      additionalProperty: topologyFacts(),
    },
    wheelsetLacingFaqSchema(),
  ].filter(Boolean),
})

// Keep JSON-LD as an SSR/head-entry snapshot for the loaded locale. Updating a
// script text node during a client-side locale switch is blocked by the site's
// Trusted Types policy; crawlers receive the correct localized snapshot from
// the server response for each public route.
useHead({
  script: [createSeoJsonLdScript(topologySchema())],
})

useHead(() => {
  // Keep the head factory reactive to lazy namespace merges on client-side
  // locale switches. The value itself is intentionally not emitted to head.
  void pageMessagesVersion.value
  return {
    title: t('wheelsetLacingTopology.seo.title'),
    meta: [
      {
        name: 'description',
        content: t('wheelsetLacingTopology.seo.description'),
        key: 'description',
      },
      {
        name: 'robots',
        content: isPublicSeoLocale.value ? 'index,follow' : 'noindex,follow,noarchive',
        key: 'robots',
      },
      {
        property: 'og:type',
        content: 'article',
        key: 'og:type',
      },
      {
        property: 'og:title',
        content: t('wheelsetLacingTopology.seo.title'),
        key: 'og:title',
      },
      {
        property: 'og:description',
        content: t('wheelsetLacingTopology.seo.description'),
        key: 'og:description',
      },
      {
        property: 'og:url',
        content: canonicalUrl.value,
        key: 'og:url',
      },
    ],
  }
})

// Display-only SVG coordinates. These constants are not measured ERD/PCD
// values and must never be passed to spoke-length or mechanical calculations.
    const SVG_RIM_HOLE_RING_DISPLAY_RADIUS = 232;             // Canvas radius for rim-hole positions
    const SVG_RIM_OUTER_EDGE_DISPLAY_RADIUS = 260;            // Canvas radius for outer rim edge
    const SVG_RIM_INNER_EDGE_DISPLAY_RADIUS = 214;            // Canvas radius for inner rim edge
    const SVG_HUB_FLANGE_HOLE_RING_DISPLAY_RADIUS_A = 66;     // Canvas radius for flange A holes
    const SVG_HUB_FLANGE_HOLE_RING_DISPLAY_RADIUS_B = 54;     // Canvas radius for flange B holes
    const SVG_HUB_AXLE_HOUSING_DISPLAY_RADIUS = 20;            // Canvas radius for axle housing
    const MINIMUM_FLANGE_DISPLAY_RADIUS = 1;
    const MAXIMUM_FLANGE_DISPLAY_RADIUS = 280;
    const MINIMUM_FLANGE_OFFSET_MM = 0;
    const MAXIMUM_FLANGE_OFFSET_MM = 100;
    const DEFAULT_G3_RIM_HOLE_SPACING_A_TO_B_DEGREES = 4.87;
    const DEFAULT_G3_RIM_HOLE_SPACING_B_TO_A_DEGREES = 4.87;
    const MINIMUM_G3_RIM_HOLE_SPACING_DEGREES = 0.1;
    const MAXIMUM_G3_RIM_HOLE_SPACING_DEGREES = 20;
    const WHEELSET_LACING_DISPLAY_GEOMETRY_REFRESH_DEBOUNCE_MS = 120;

    const resolveWheelsetLacingBackendTopologySelection = (holes, cross) => (
      resolveWheelsetLacingDisplayGeometryTopologySelection(holes, cross)
    );

    const resolveWheelsetLacingBackendTopologyIdentifier = (holes, cross) => (
      resolveWheelsetLacingBackendTopologySelection(holes, cross).topologyId
    );

    const formatBackendDisplayGeometryMetric = (value) => (
      Number.isFinite(value) ? Number(value).toFixed(1) : '—'
    );

    const formatBackendDisplayGeometryBarWidth = (value, maximum) => {
      if (!Number.isFinite(value) || maximum <= 0) return 0;
      return Math.min(100, Math.max(0, (value / maximum) * 100));
    };

    const resolveDisplayGeometryRadius = (points, fallbackRadius) => {
      const radii = (points || [])
        .map(point => Math.hypot(Number(point.x), Number(point.y)))
        .filter(Number.isFinite);
      if (radii.length === 0) return fallbackRadius;
      return Math.round(Math.max(...radii) * 100) / 100;
    };

    const resolveInitialFlangeOffset = (profile, field, fallbackOffset) => {
      const offset = Number(profile?.[field]);
      return Number.isFinite(offset) && offset >= MINIMUM_FLANGE_OFFSET_MM && offset <= MAXIMUM_FLANGE_OFFSET_MM
        ? offset
        : fallbackOffset;
    };

    const resolveInitialG3RimHoleSpacing = (spacingProfile, field, fallbackSpacing) => {
      const spacing = Number(spacingProfile?.[field]);
      return Number.isFinite(spacing) && spacing >= MINIMUM_G3_RIM_HOLE_SPACING_DEGREES && spacing <= MAXIMUM_G3_RIM_HOLE_SPACING_DEGREES
        ? spacing
        : fallbackSpacing;
    };

    const initialServerRenderedFlangeProfile = initialServerRenderedDisplayGeometry?.flange_profile;
    const initialServerRenderedG3GroupSpacing = initialServerRenderedDisplayGeometry?.g3_group_spacing;

    let displayGeometryRefreshTimer = null;

    const buildWheelsetLacingDisplayGeometryRequestBody = () => {
      const requestBody = {
        topology_id: resolveWheelsetLacingBackendTopologySelection(state.holes, state.cross).topologyId,
        rim_radius: SVG_RIM_HOLE_RING_DISPLAY_RADIUS,
        flange_radius_a: state.flangeRadiusA,
        flange_radius_b: state.flangeRadiusB,
        flange_offset_a_mm: state.flangeOffsetAMm,
        flange_offset_b_mm: state.flangeOffsetBMm,
      };
      if (state.holes === 21) {
        requestBody.g3_rim_hole_spacing_a_to_b_degrees = state.g3RimHoleSpacingAToBDegrees;
        requestBody.g3_rim_hole_spacing_b_to_a_degrees = state.g3RimHoleSpacingBToADegrees;
      }
      return requestBody;
    };

    const refreshWheelsetLacingDisplayGeometryFromBackend = async () => {
      const requestId = ++displayGeometryRequestSequence;
      const requestBody = buildWheelsetLacingDisplayGeometryRequestBody();
      const selectedTopologyId = requestBody.topology_id;
      const selectedTopology = resolveWheelsetLacingBackendTopologySelection(state.holes, state.cross);
      displayGeometryController?.abort();
      displayGeometryController = new AbortController();
      backendDisplayGeometry.value = null;
      displayGeometryError.value = null;
      renderBlueprint();
      try {
        const response = await request('/wheelset-lacing/display-geometry', {
          method: 'POST',
          signal: displayGeometryController.signal,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(requestBody),
        });
        if (requestId !== displayGeometryRequestSequence || selectedTopologyId !== resolveWheelsetLacingBackendTopologyIdentifier(state.holes, state.cross)) return;
        validateWheelsetLacingDisplayGeometryResponse(response?.data, selectedTopology);
        backendDisplayGeometry.value = response.data;
        renderBlueprint();
      } catch (error) {
        if (requestId !== displayGeometryRequestSequence || (error instanceof DOMException && error.name === 'AbortError')) return;
        backendDisplayGeometry.value = null;
        displayGeometryError.value = error instanceof Error ? error.message : 'wheelset lacing display geometry request failed';
        renderBlueprint();
      }
    };

    const scheduleWheelsetLacingDisplayGeometryRefresh = () => {
      if (displayGeometryRefreshTimer !== null) {
        window.clearTimeout(displayGeometryRefreshTimer);
      }
      displayGeometryRefreshTimer = window.setTimeout(() => {
        displayGeometryRefreshTimer = null;
        void refreshWheelsetLacingDisplayGeometryFromBackend();
      }, WHEELSET_LACING_DISPLAY_GEOMETRY_REFRESH_DEBOUNCE_MS);
    };

    const cancelScheduledWheelsetLacingDisplayGeometryRefresh = () => {
      if (displayGeometryRefreshTimer === null) return;
      window.clearTimeout(displayGeometryRefreshTimer);
      displayGeometryRefreshTimer = null;
    };

    let state = {
      holes: 24,            // 16, 20, 21, 24, 28, 32, 36, '24_2to1'
      cross: 2,             // 0, 1, 2, 3, 4
      viewMode: 'both',     // 默认全景双侧透视，确保所有孔位100%全满严整
      flangeRadiusA: resolveDisplayGeometryRadius(initialServerRenderedDisplayGeometry?.hub_holes_a, SVG_HUB_FLANGE_HOLE_RING_DISPLAY_RADIUS_A),
      flangeRadiusB: resolveDisplayGeometryRadius(initialServerRenderedDisplayGeometry?.hub_holes_b, SVG_HUB_FLANGE_HOLE_RING_DISPLAY_RADIUS_B),
      flangeOffsetAMm: resolveInitialFlangeOffset(initialServerRenderedFlangeProfile, 'flange_offset_a_mm', 20),
      flangeOffsetBMm: resolveInitialFlangeOffset(initialServerRenderedFlangeProfile, 'flange_offset_b_mm', 35),
      g3RimHoleSpacingAToBDegrees: resolveInitialG3RimHoleSpacing(initialServerRenderedG3GroupSpacing, 'spacing_a_to_b_degrees', DEFAULT_G3_RIM_HOLE_SPACING_A_TO_B_DEGREES),
      g3RimHoleSpacingBToADegrees: resolveInitialG3RimHoleSpacing(initialServerRenderedG3GroupSpacing, 'spacing_b_to_a_degrees', DEFAULT_G3_RIM_HOLE_SPACING_B_TO_A_DEGREES),
      showLeading: true,
      showTrailing: true,
      showNonDrive: true,
    };

    // Default values only choose the first useful preview for a selection.
    // Supported combinations themselves come from the pure topology contract.
    const DEFAULT_PREVIEW_CROSS_COUNT_BY_SELECTION = {
      16: 0,
      20: 1,
      21: 2,
      24: 2,
      '24_2to1': 2,
      28: 2,
      32: 3,
      36: 3,
    };

    const getHoleRule = (holes) => {
      const translatedLabel = holes === 21
        ? `${t('wheelsetLacingTopology.holes.21')} (${t('wheelsetLacingTopology.holes.21Sub')})`
        : holes === '24_2to1'
          ? t('wheelsetLacingTopology.holes.24_2to1')
          : holes === 24
            ? `${t('wheelsetLacingTopology.holes.24')} (1:1)`
            : t(`wheelsetLacingTopology.holes.${holes}`)
      return {
        label: translatedLabel,
        allowedCross: getSupportedWheelsetLacingCrossCounts(holes),
        recommended: DEFAULT_PREVIEW_CROSS_COUNT_BY_SELECTION[holes],
      }
    }

    function setHoleCount(holes) {
      state.holes = holes;
      const rule = getHoleRule(holes);
      if (!rule.allowedCross.includes(state.cross)) {
        state.cross = rule.recommended;
      }
      renderControls();
      cancelScheduledWheelsetLacingDisplayGeometryRefresh();
      void refreshWheelsetLacingDisplayGeometryFromBackend();
    }

    function setCrossCount(cross) {
      state.cross = cross;
      renderControls();
      cancelScheduledWheelsetLacingDisplayGeometryRefresh();
      void refreshWheelsetLacingDisplayGeometryFromBackend();
    }

    const getWheelsetLacingGeometryInputLimits = field => {
      if (field.startsWith('flangeRadius')) {
        return { min: MINIMUM_FLANGE_DISPLAY_RADIUS, max: MAXIMUM_FLANGE_DISPLAY_RADIUS };
      }
      if (field.startsWith('g3RimHoleSpacing')) {
        return { min: MINIMUM_G3_RIM_HOLE_SPACING_DEGREES, max: MAXIMUM_G3_RIM_HOLE_SPACING_DEGREES };
      }
      return { min: MINIMUM_FLANGE_OFFSET_MM, max: MAXIMUM_FLANGE_OFFSET_MM };
    };

    function updateWheelsetLacingGeometryInput(field, event) {
      const rawValue = String(event?.target?.value ?? '').trim();
      const limits = getWheelsetLacingGeometryInputLimits(field);
      const inputValue = Number(rawValue);
      if (!rawValue || !Number.isFinite(inputValue) || inputValue < limits.min || inputValue > limits.max) {
        cancelScheduledWheelsetLacingDisplayGeometryRefresh();
        return;
      }
      state[field] = inputValue;
      scheduleWheelsetLacingDisplayGeometryRefresh();
    }

    function normalizeWheelsetLacingGeometryInput(field, event) {
      const target = event?.target;
      const rawValue = String(target?.value ?? '').trim();
      const limits = getWheelsetLacingGeometryInputLimits(field);
      const previousValue = state[field];
      const inputValue = Number(rawValue);
      const normalizedValue = rawValue && Number.isFinite(inputValue)
        ? Math.min(limits.max, Math.max(limits.min, inputValue))
        : previousValue;
      state[field] = normalizedValue;
      if (target) target.value = String(normalizedValue);
      if (normalizedValue !== previousValue) {
        scheduleWheelsetLacingDisplayGeometryRefresh();
      }
    }

    function setViewMode(mode) {
      state.viewMode = mode;
      document.querySelectorAll('.view-mode-grid .view-pill-btn').forEach(btn => {
        btn.classList.remove('active');
        btn.setAttribute('aria-pressed', 'false');
      });
      const activeBtn = document.getElementById(`btn-view-${mode}`);
      if (activeBtn) {
        activeBtn.classList.add('active');
        activeBtn.setAttribute('aria-pressed', 'true');
      }
      renderBlueprint();
    }

    function updateCanvasLayers() {
      state.showLeading = document.getElementById('toggle-leading').checked;
      state.showTrailing = document.getElementById('toggle-trailing').checked;
      state.showNonDrive = document.getElementById('toggle-nondrive').checked;
      renderBlueprint();
    }

    function renderControls() {
      const rule = getHoleRule(state.holes);

      document.querySelectorAll('.hole-pill-grid .uds-pill-btn').forEach(btn => {
        const h = btn.getAttribute('data-holes');
        if (h === String(state.holes)) {
          btn.classList.add('active');
          btn.setAttribute('aria-pressed', 'true');
        } else {
          btn.classList.remove('active');
          btn.setAttribute('aria-pressed', 'false');
        }
      });
      document.getElementById('current-hole-label').innerText = rule.label;

      document.querySelectorAll('.cross-pill-grid .uds-pill-btn').forEach(btn => {
        const c = parseInt(btn.getAttribute('data-cross'), 10);
        const isAllowed = rule.allowedCross.includes(c);
        btn.disabled = !isAllowed;
        if (c === state.cross && isAllowed) {
          btn.classList.add('active');
          btn.setAttribute('aria-pressed', 'true');
        } else {
          btn.classList.remove('active');
          btn.setAttribute('aria-pressed', 'false');
        }
      });

      const crossNames = [
        t('wheelsetLacingTopology.cross.0'),
        t('wheelsetLacingTopology.cross.1'),
        t('wheelsetLacingTopology.cross.2'),
        t('wheelsetLacingTopology.cross.3'),
        t('wheelsetLacingTopology.cross.4'),
      ];
      document.getElementById('current-cross-label').innerText = crossNames[state.cross] || t('wheelsetLacingTopology.cross.label', { count: state.cross });
      document.querySelectorAll('.view-mode-grid .view-pill-btn').forEach(btn => {
        btn.setAttribute('aria-pressed', String(btn.id === `btn-view-${state.viewMode}`));
      });
      const g3GeometryControlGroup = document.getElementById('g3-geometry-control-group');
      if (g3GeometryControlGroup) {
        const isG3Selection = state.holes === 21;
        g3GeometryControlGroup.hidden = !isG3Selection;
        g3GeometryControlGroup.setAttribute('aria-hidden', String(!isG3Selection));
      }
      document.getElementById('flange-radius-a-input').value = String(state.flangeRadiusA);
      document.getElementById('flange-radius-b-input').value = String(state.flangeRadiusB);
      document.getElementById('flange-offset-a-input').value = String(state.flangeOffsetAMm);
      document.getElementById('flange-offset-b-input').value = String(state.flangeOffsetBMm);
      document.getElementById('g3-spacing-a-to-b-input').value = String(state.g3RimHoleSpacingAToBDegrees);
      document.getElementById('g3-spacing-b-to-a-input').value = String(state.g3RimHoleSpacingBToADegrees);
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
      const selectionLabel = getHoleRule(holes).label;
      const crossLabel = t(`wheelsetLacingTopology.cross.${cross}`);
      const svgDescription = t('wheelsetLacingTopology.canvas.ariaRuntime', {
        selection: selectionLabel,
        cross: crossLabel,
      });
      svg.setAttribute('aria-label', svgDescription);
      appendSvgElement(svg, 'desc', { id: 'spoke-lacing-svg-description' }, svgDescription);

      // 1. 底层几何同心圆
      const bgGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
      appendSvgElement(bgGroup, 'circle', { cx: 0, cy: 0, r: SVG_RIM_OUTER_EDGE_DISPLAY_RADIUS, fill: 'none', stroke: '#cbd5e1', 'stroke-width': 2.5 });
      appendSvgElement(bgGroup, 'circle', { cx: 0, cy: 0, r: SVG_RIM_OUTER_EDGE_DISPLAY_RADIUS - 12, fill: 'none', stroke: '#e2e8f0', 'stroke-width': 1.2, 'stroke-dasharray': '6 4' });
      appendSvgElement(bgGroup, 'circle', { cx: 0, cy: 0, r: SVG_RIM_HOLE_RING_DISPLAY_RADIUS, fill: 'rgba(5, 150, 105, 0.015)', stroke: '#059669', 'stroke-width': 1.6, 'stroke-dasharray': '4 4' });
      appendSvgElement(bgGroup, 'circle', { cx: 0, cy: 0, r: SVG_RIM_INNER_EDGE_DISPLAY_RADIUS, fill: 'none', stroke: '#e2e8f0', 'stroke-width': 1.8 });
      appendSvgElement(bgGroup, 'line', { x1: -280, y1: 0, x2: 280, y2: 0, stroke: 'rgba(15, 23, 42, 0.08)', 'stroke-width': 1, 'stroke-dasharray': '4 4' });
      appendSvgElement(bgGroup, 'line', { x1: 0, y1: -280, x2: 0, y2: 280, stroke: 'rgba(15, 23, 42, 0.08)', 'stroke-width': 1, 'stroke-dasharray': '4 4' });
      svg.appendChild(bgGroup);

      // The backend owns topology generation and geometry projection. The
      // browser only filters the returned line segments for the selected view
      // and creates SVG nodes.
      const geometry = backendDisplayGeometry.value;
      if (!geometry) {
        appendSvgElement(
          svg,
          'text',
          { x: 0, y: 0, 'text-anchor': 'middle', fill: displayGeometryError.value ? '#b91c1c' : '#64748b', 'font-size': 12 },
          displayGeometryError.value
            ? t('wheelsetLacingTopology.telemetry.backendRejected')
            : t('wheelsetLacingTopology.telemetry.backendPending'),
        );
        renderFlangeProfile(null);
        updateGeometryProjectionMetricsAndTopologyReview(null, null, null);
        return;
      }
      const topology = geometry.topology;
      const rimHoles = geometry.rim_holes;
      const hubHolesA = geometry.hub_holes_a;
      const hubHolesB = geometry.hub_holes_b;
      const spokes = geometry.spokes;
      const flangeRadiusA = resolveDisplayGeometryRadius(hubHolesA, state.flangeRadiusA);
      const flangeRadiusB = resolveDisplayGeometryRadius(hubHolesB, state.flangeRadiusB);
      const hubOuterRadius = Math.min(280, Math.max(flangeRadiusA, flangeRadiusB) + 12);

      // 3. Select the visible spoke line segments for the blueprint.
      const visibleSpokes = spokes.filter(s => {
        if (viewMode === 'drive' && s.side !== 'A') return false;
        if (viewMode === 'nondrive' && s.side !== 'B') return false;
        if (s.side === 'A' && s.type === 'leading' && !state.showLeading) return false;
        if (s.side === 'A' && s.type === 'trailing' && !state.showTrailing) return false;
        if (s.side === 'B' && !state.showNonDrive) return false;
        return true;
      });

      // 4. 先绘制花鼓底层金属底盘与 PCD 节圆 (置于辐条下方，避免遮挡辐条穿入孔心)
      const hubBaseGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
      hubBaseGroup.setAttribute('id', 'layer-hub-base');
      appendSvgElement(hubBaseGroup, 'circle', { cx: 0, cy: 0, r: hubOuterRadius, fill: '#f1f5f9', stroke: '#475569', 'stroke-width': 2 });
      appendSvgElement(hubBaseGroup, 'circle', { cx: 0, cy: 0, r: flangeRadiusA, fill: 'none', stroke: '#d97706', 'stroke-width': 1.4, 'stroke-dasharray': '3 3' });
      appendSvgElement(hubBaseGroup, 'circle', { cx: 0, cy: 0, r: flangeRadiusB, fill: 'none', stroke: '#0284c7', 'stroke-width': 1.2, 'stroke-dasharray': '2 2' });
      appendSvgElement(hubBaseGroup, 'circle', { cx: 0, cy: 0, r: SVG_HUB_AXLE_HOUSING_DISPLAY_RADIUS, fill: '#0f172a', stroke: '#cbd5e1', 'stroke-width': 1.8 });
      appendSvgElement(hubBaseGroup, 'circle', { cx: 0, cy: 0, r: 5, fill: '#38bdf8' });
      svg.appendChild(hubBaseGroup);

      // 5. 绘制辐条图层 (从 PCD 均分孔心到车圈孔心，完整暴露无遮挡)
      const spokeGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
      spokeGroup.setAttribute('id', 'layer-spokes');
      visibleSpokes.forEach(spk => {
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
        title.textContent = t('wheelsetLacingTopology.svg.spokeTitle', {
          index: spk.id + 1,
          side: spk.side === 'A'
            ? t('wheelsetLacingTopology.svg.driveSide')
            : t('wheelsetLacingTopology.svg.nonDriveSide'),
          hubX: spk.hub.x.toFixed(1),
          hubY: spk.hub.y.toFixed(1),
          rimX: spk.rim.x.toFixed(1),
          rimY: spk.rim.y.toFixed(1),
        });
        line.appendChild(title);

        spokeGroup.appendChild(line);
      });
      svg.appendChild(spokeGroup);

      // 6. Draw the top hub holes and rim holes above the spoke lines.
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
      rimHoles.forEach(r => {
        let holeFill = (r.side === 'A') ? '#059669' : '#0284c7';
        appendSvgElement(rimHoleGroup, 'circle', { cx: r.x, cy: r.y, r: 4.2, fill: '#ffffff', stroke: holeFill, 'stroke-width': 2 });
        appendSvgElement(rimHoleGroup, 'circle', { cx: r.x, cy: r.y, r: 1.8, fill: '#0f172a' });
      });

      svg.appendChild(rimHoleGroup);

      // 7. Geometry telemetry and topology rule review.
      renderFlangeProfile(geometry.flange_profile);
      updateGeometryProjectionMetricsAndTopologyReview(topology, geometry.metrics, geometry.g3_group_spacing);
    }

    function renderFlangeProfile(profile) {
      const svg = document.getElementById('flange-profile-svg');
      const summary = document.getElementById('flange-profile-summary');
      const statusTag = document.getElementById('flange-profile-status-tag');
      svg.replaceChildren();
      const profileCoordinateKeys = ['centerline_x', 'flange_a_x', 'flange_b_x', 'axle_left_x', 'axle_right_x', 'flange_offset_a_mm', 'flange_offset_b_mm', 'total_flange_span_mm'];
      const hasValidProfile = profile && profileCoordinateKeys.every(key => Number.isFinite(Number(profile[key])));
      if (!hasValidProfile) {
        appendSvgElement(svg, 'text', { x: 0, y: 8, 'text-anchor': 'middle', fill: '#64748b', 'font-size': 10 }, t('wheelsetLacingTopology.canvas.profilePending'));
        summary.innerText = t('wheelsetLacingTopology.canvas.profilePending');
        statusTag.innerText = t('wheelsetLacingTopology.canvas.profileTag');
        return;
      }

      const profileDescription = t('wheelsetLacingTopology.canvas.profileRuntime', {
        driveOffset: formatBackendDisplayGeometryMetric(profile.flange_offset_a_mm),
        nonDriveOffset: formatBackendDisplayGeometryMetric(profile.flange_offset_b_mm),
        total: formatBackendDisplayGeometryMetric(profile.total_flange_span_mm),
      });
      svg.setAttribute('aria-label', profileDescription);
      appendSvgElement(svg, 'desc', { id: 'flange-profile-svg-description' }, profileDescription);
      const defs = document.createElementNS(SVG_NS, 'defs');
      const createProfileArrowMarker = (id, color) => {
        const marker = appendSvgElement(defs, 'marker', { id, viewBox: '0 0 10 10', refX: 5, refY: 5, markerWidth: 5, markerHeight: 5, orient: 'auto-start-reverse' });
        appendSvgElement(marker, 'path', { d: 'M 0 1 L 8 5 L 0 9 z', fill: color });
      };
      createProfileArrowMarker('flange-profile-arrow-sky', '#0284c7');
      createProfileArrowMarker('flange-profile-arrow-amber', '#d97706');
      svg.appendChild(defs);
      appendSvgElement(svg, 'line', { x1: profile.axle_left_x, y1: 0, x2: profile.axle_right_x, y2: 0, stroke: '#475569', 'stroke-width': 8, 'stroke-linecap': 'round' });
      appendSvgElement(svg, 'line', { x1: profile.centerline_x, y1: -62, x2: profile.centerline_x, y2: 62, stroke: '#059669', 'stroke-width': 1.5, 'stroke-dasharray': '5 4' });
      appendSvgElement(svg, 'text', { x: profile.centerline_x + 5, y: -66, fill: '#059669', 'font-size': 8, 'font-weight': 800 }, t('wheelsetLacingTopology.canvas.profileCenterline'));

      appendSvgElement(svg, 'line', { x1: profile.flange_b_x, y1: -34, x2: profile.flange_b_x, y2: 34, stroke: '#0284c7', 'stroke-width': 6, 'stroke-linecap': 'round' });
      appendSvgElement(svg, 'line', { x1: profile.flange_a_x, y1: -34, x2: profile.flange_a_x, y2: 34, stroke: '#d97706', 'stroke-width': 6, 'stroke-linecap': 'round' });
      appendSvgElement(svg, 'text', { x: profile.flange_b_x, y: 52, fill: '#0284c7', 'font-size': 8, 'font-weight': 800, 'text-anchor': 'middle' }, `${t('wheelsetLacingTopology.canvas.profileNonDrive')} ${profile.flange_offset_b_mm} mm`);
      appendSvgElement(svg, 'text', { x: profile.flange_a_x, y: -48, fill: '#b45309', 'font-size': 8, 'font-weight': 800, 'text-anchor': 'middle' }, `${t('wheelsetLacingTopology.canvas.profileDrive')} ${profile.flange_offset_a_mm} mm`);

      appendSvgElement(svg, 'line', { x1: profile.flange_b_x, y1: 70, x2: profile.centerline_x, y2: 70, stroke: '#0284c7', 'stroke-width': 1.2, 'marker-start': 'url(#flange-profile-arrow-sky)', 'marker-end': 'url(#flange-profile-arrow-sky)' });
      appendSvgElement(svg, 'line', { x1: profile.centerline_x, y1: -70, x2: profile.flange_a_x, y2: -70, stroke: '#d97706', 'stroke-width': 1.2, 'marker-start': 'url(#flange-profile-arrow-amber)', 'marker-end': 'url(#flange-profile-arrow-amber)' });
      summary.innerText = profileDescription;
      statusTag.innerText = t('wheelsetLacingTopology.canvas.profileTag');
    }

    function updateGeometryProjectionMetricsAndTopologyReview(topology, geometryProjectionMetrics, g3GroupSpacing) {
      const holes = topology?.selection ?? state.holes;
      const displayLayout = topology?.display_layout ?? null;
      const isG3Topology = displayLayout === WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.g3Triplet2To1;
      const isUniformTwoToOneTopology = displayLayout === WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform2To1;
      const numHoles = Number(topology?.hole_count ?? (holes === '24_2to1' ? 24 : holes));
      const aggregateMeanAbsoluteProjectionAngleDegrees = geometryProjectionMetrics?.aggregate_mean_absolute_projection_angle_degrees ?? null;

      document.getElementById('metric-tangential-projection-angle').innerText = formatBackendDisplayGeometryMetric(aggregateMeanAbsoluteProjectionAngleDegrees);
      document.getElementById('bar-tangential-projection-angle').style.width = `${formatBackendDisplayGeometryBarWidth(aggregateMeanAbsoluteProjectionAngleDegrees, 90)}%`;
      document.getElementById('metric-tangential-projection-angle-desc').innerText = t('wheelsetLacingTopology.telemetry.tangentRuntime', {
        count: geometryProjectionMetrics?.drive_side_spoke_count ?? '—',
        min: formatBackendDisplayGeometryMetric(geometryProjectionMetrics?.minimum_absolute_drive_side_projection_angle_degrees),
        max: formatBackendDisplayGeometryMetric(geometryProjectionMetrics?.maximum_absolute_drive_side_projection_angle_degrees),
      });

      document.getElementById('metric-tangential-projection').innerText = formatBackendDisplayGeometryMetric(geometryProjectionMetrics?.mean_absolute_tangential_projection_percent);
      document.getElementById('bar-tangential-projection').style.width = `${formatBackendDisplayGeometryBarWidth(geometryProjectionMetrics?.mean_absolute_tangential_projection_percent, 100)}%`;
      document.getElementById('metric-tangential-projection-desc').innerText = t('wheelsetLacingTopology.telemetry.tangentialRuntime');

      document.getElementById('metric-radial-projection').innerText = formatBackendDisplayGeometryMetric(geometryProjectionMetrics?.mean_absolute_radial_projection_percent);
      document.getElementById('bar-radial-projection').style.width = `${formatBackendDisplayGeometryBarWidth(geometryProjectionMetrics?.mean_absolute_radial_projection_percent, 100)}%`;
      document.getElementById('metric-radial-projection-desc').innerText = t('wheelsetLacingTopology.telemetry.radialRuntime');

      if (!topology || !geometryProjectionMetrics) {
        const statusBox = document.getElementById('topology-status-box');
        const statusTitle = document.getElementById('status-title-text');
        const statusDetail = document.getElementById('status-detail-text');
        const builderTip = document.getElementById('builder-tip-text');
        const statusTag = document.getElementById('canvas-status-tag');
        const hasDisplayGeometryError = Boolean(displayGeometryError.value);
        statusBox.className = hasDisplayGeometryError
          ? 'topology-status-card topology-status-error'
          : 'topology-status-card topology-status-preview';
        statusTitle.innerText = hasDisplayGeometryError
          ? t('wheelsetLacingTopology.review.backendRejectedTitle')
          : t('wheelsetLacingTopology.review.backendPendingTitle');
        statusDetail.innerText = hasDisplayGeometryError
          ? t('wheelsetLacingTopology.review.backendRejectedDetail')
          : t('wheelsetLacingTopology.review.backendPendingDetail');
        builderTip.innerText = hasDisplayGeometryError
          ? t('wheelsetLacingTopology.review.backendRejectedTip')
          : t('wheelsetLacingTopology.review.noteTip');
        statusTag.innerText = hasDisplayGeometryError
          ? t('wheelsetLacingTopology.review.tagUnavailable')
          : t('wheelsetLacingTopology.review.tagPreview');
        statusTag.style.background = hasDisplayGeometryError ? 'rgba(185, 28, 28, 0.1)' : 'rgba(5, 150, 105, 0.1)';
        statusTag.style.color = hasDisplayGeometryError ? '#b91c1c' : '#059669';
        return;
      }

      // This status confirms only that the supported topology was generated.
      // It does not infer flange clearance, assembly safety, or mechanics.
      const statusBox = document.getElementById('topology-status-box');
      const statusTitle = document.getElementById('status-title-text');
      const statusDetail = document.getElementById('status-detail-text');
      const builderTip = document.getElementById('builder-tip-text');
      const statusTag = document.getElementById('canvas-status-tag');

      statusBox.className = 'topology-status-card topology-status-preview';
      if (isG3Topology) {
        statusTitle.innerText = t('wheelsetLacingTopology.review.preview21Title');
        statusDetail.innerText = t('wheelsetLacingTopology.review.preview21DetailRuntime', {
          spacingAToB: formatBackendDisplayGeometryMetric(g3GroupSpacing?.spacing_a_to_b_degrees),
          spacingBToA: formatBackendDisplayGeometryMetric(g3GroupSpacing?.spacing_b_to_a_degrees),
        });
        builderTip.innerText = t('wheelsetLacingTopology.review.preview21Tip');
        statusTag.innerText = t('wheelsetLacingTopology.review.tagG3');
        statusTag.style.background = 'rgba(5, 150, 105, 0.1)';
        statusTag.style.color = '#059669';
        return;
      }

      statusTitle.innerText = t('wheelsetLacingTopology.review.previewTitle');
      statusDetail.innerText = t('wheelsetLacingTopology.review.previewDetail', {
        holes: numHoles,
        cross: topology.cross,
        angle: formatBackendDisplayGeometryMetric(aggregateMeanAbsoluteProjectionAngleDegrees),
      });
      builderTip.innerText = isUniformTwoToOneTopology
        ? t('wheelsetLacingTopology.review.uniformTip')
        : t('wheelsetLacingTopology.review.generalTip');
      statusTag.innerText = t('wheelsetLacingTopology.review.tagPreview');
      statusTag.style.background = 'rgba(5, 150, 105, 0.1)';
      statusTag.style.color = '#059669';
    }

onMounted(() => {
  isInteractiveBlueprintMounted = true
  renderControls()
  if (backendDisplayGeometry.value) {
    renderBlueprint()
    return
  }
  void refreshWheelsetLacingDisplayGeometryFromBackend()
})

onBeforeUnmount(() => {
  isInteractiveBlueprintMounted = false
  if (displayGeometryRefreshTimer !== null) {
    window.clearTimeout(displayGeometryRefreshTimer)
  }
  displayGeometryController?.abort()
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
      --accent-steel: #475569;       /* 机械冷钢 (基准尺寸) */

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
      font-style: normal;
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
      width: 100%;
      max-width: none;
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
      .geometry-metrics-column {
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

    /* SSR-visible fact block: keep the page's stable reference facts readable
       before the interactive SVG hydrates and easy to scan on narrow screens. */
    .geo-fact-summary {
      gap: 14px;
    }

    .geo-fact-summary__intro {
      max-width: 980px;
      color: var(--text-secondary);
      font-size: 11px;
      line-height: 1.65;
    }

    .geo-fact-summary__grid {
      display: grid;
      grid-template-columns: repeat(4, minmax(0, 1fr));
      gap: 12px;
      margin: 0;
    }

    .geo-fact-summary__grid > div {
      min-width: 0;
      padding: 12px 14px;
      border: 1px solid var(--border-line);
      border-radius: 16px;
      background: var(--bg-inset);
    }

    .geo-fact-summary__grid dt {
      color: var(--text-main);
      font-size: 10px;
      font-weight: 900;
      letter-spacing: 0.04em;
    }

    .geo-fact-summary__grid dd {
      margin: 6px 0 0;
      color: var(--text-secondary);
      font-size: 10px;
      line-height: 1.55;
    }

    @media (max-width: 1024px) {
      .geo-fact-summary__grid {
        grid-template-columns: repeat(2, minmax(0, 1fr));
      }
    }

    @media (max-width: 640px) {
      .geo-fact-summary__grid {
        grid-template-columns: 1fr;
      }
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

    .flange-geometry-control-group {
      display: flex;
      flex-direction: column;
      gap: 8px;
      padding: 12px;
      border: 1px solid var(--border-line);
      border-radius: 16px;
      background: rgba(241, 245, 249, 0.68);
    }

    .g3-geometry-control-group[hidden] {
      display: none !important;
    }

    .control-unit-label {
      color: var(--text-dim);
      font-size: 8px;
      font-weight: 800;
      letter-spacing: 0.04em;
      text-transform: uppercase;
    }

    .flange-geometry-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 8px;
    }

    .flange-geometry-field {
      display: flex;
      min-width: 0;
      flex-direction: column;
      gap: 4px;
      color: var(--text-secondary);
      font-size: 8.5px;
      font-weight: 800;
      line-height: 1.2;
    }

    .flange-geometry-input-line {
      display: flex;
      align-items: center;
      gap: 4px;
      color: var(--text-muted);
      font-family: var(--tz-font-ui);
      font-size: 8px;
    }

    .flange-geometry-input-line input {
      width: 100%;
      min-width: 0;
      height: 28px;
      padding: 0 6px;
      border: 1px solid var(--border-line);
      border-radius: 8px;
      background: #ffffff;
      color: var(--text-main);
      font-family: var(--tz-font-ui);
      font-size: 10px;
      font-weight: 800;
    }

    .flange-geometry-input-line input:focus-visible {
      outline: 2px solid rgba(5, 150, 105, 0.35);
      outline-offset: 1px;
      border-color: var(--accent-primary);
    }

    .flange-geometry-help {
      color: var(--text-muted);
      font-size: 8.5px;
      line-height: 1.4;
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

    .flange-profile-viewport {
      width: 100%;
      padding: 12px 20px 16px;
      border-top: 1px solid var(--border-line);
      background: rgba(248, 250, 252, 0.72);
    }

    .flange-profile-toolbar {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      margin-bottom: 4px;
    }

    .flange-profile-summary {
      max-width: 60%;
      color: var(--text-muted);
      font-size: 8.5px;
      line-height: 1.35;
      text-align: right;
    }

    .flange-profile-viewport svg {
      display: block;
      width: 100%;
      height: 142px;
      max-width: 600px;
      margin: 0 auto;
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

    /* 右侧几何指标看板 */
    .geometry-metrics-column {
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

    /* 拓扑生成状态卡片 */
    .topology-status-card {
      border-radius: 18px;
      padding: 14px 16px;
      border: 1px solid transparent;
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    .topology-status-preview {
      background: rgba(5, 150, 105, 0.06);
      border-color: rgba(5, 150, 105, 0.25);
      color: #047857;
    }

    .topology-status-error {
      background: rgba(185, 28, 28, 0.06);
      border-color: rgba(185, 28, 28, 0.25);
      color: #b91c1c;
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

/* Spoke lines are inserted into the SVG by the blueprint renderer after mount. */
.spoke-lacing-page :deep(.spoke-line) {
  vector-effect: non-scaling-stroke;
}
</style>
