<template>
  <svg class="cad-svg" viewBox="0 0 540 440" role="img" :aria-label="t('guidesTirePressure.dashboard.svgLabel')">
    <title>{{ t('guidesTirePressure.dashboard.svgLabel') }}</title>
    <desc>{{ t('guidesTirePressure.dashboard.svgDescription') }}</desc>
    <defs>
      <pattern id="tire-ground-grid" width="16" height="16" patternUnits="userSpaceOnUse"><path d="M16 0H0V16" fill="none" stroke="#e2e8f0" stroke-width="1"/></pattern>
      <radialGradient id="tire-patch-front"><stop offset="0" stop-color="#0284c7" stop-opacity=".65"/><stop offset="1" stop-color="#0284c7" stop-opacity=".08"/></radialGradient>
      <radialGradient id="tire-patch-rear"><stop offset="0" stop-color="#d97706" stop-opacity=".65"/><stop offset="1" stop-color="#d97706" stop-opacity=".08"/></radialGradient>
      <marker id="tire-force-arrow" markerWidth="7" markerHeight="7" refX="6" refY="3.5" orient="auto" markerUnits="userSpaceOnUse"><path d="M0 0L7 3.5L0 7Z" fill="#dc2626"/></marker>
      <marker id="tire-load-arrow-front" markerWidth="7" markerHeight="7" refX="6" refY="3.5" orient="auto" markerUnits="userSpaceOnUse"><path d="M0 0L7 3.5L0 7Z" fill="#0284c7"/></marker>
      <marker id="tire-load-arrow-rear" markerWidth="7" markerHeight="7" refX="6" refY="3.5" orient="auto" markerUnits="userSpaceOnUse"><path d="M0 0L7 3.5L0 7Z" fill="#d97706"/></marker>
    </defs>

    <rect x="20" y="220" width="500" height="38" fill="url(#tire-ground-grid)"/>
    <line x1="20" y1="220" x2="520" y2="220" stroke="#0f172a" stroke-width="2.5"/>
    <g v-for="wheel in wheels" :key="wheel.id">
      <ellipse :cx="wheel.geometry.contactX" cy="222" :rx="wheel.geometry.contactShadow" ry="3.5" :fill="wheel.color" opacity=".14"/>
      <g :transform="`rotate(${leanAngle}, ${wheel.x}, 220)`">
        <ellipse :cx="wheel.x" :cy="wheel.geometry.centerY" :rx="wheel.geometry.tireHalfWidth" :ry="wheel.geometry.tireRadius" fill="#1e293b" stroke="#0f172a" stroke-width="1.2"/>
        <ellipse :cx="wheel.x" :cy="wheel.geometry.centerY" :rx="wheel.geometry.airHalfWidth" :ry="wheel.geometry.tireRadius - 8" :fill="wheel.id === 'front' ? '#e0f2fe' : '#fef3c7'" stroke="#64748b" stroke-width="1.2"/>
        <ellipse :cx="wheel.x" :cy="wheel.geometry.centerY" :rx="wheel.geometry.rimHalfWidth" :ry="wheel.geometry.rimRadius" fill="#f8fafc" stroke="#475569" stroke-width="4"/>
        <line :x1="wheel.x - wheel.geometry.axleHalfLength" :y1="wheel.geometry.centerY" :x2="wheel.x + wheel.geometry.axleHalfLength" :y2="wheel.geometry.centerY" stroke="#334155" stroke-width="3"/>
        <circle :cx="wheel.x" :cy="wheel.geometry.centerY" r="6" fill="#64748b" stroke="#334155" stroke-width="1.5"/><circle :cx="wheel.x" :cy="wheel.geometry.centerY" r="2" fill="#e2e8f0"/>
      </g>
      <g class="force-labels">
        <rect :x="wheel.x - 64" y="12" width="128" height="20" rx="6" :fill="wheel.id === 'front' ? '#f0f9ff' : '#fffbeb'" :stroke="wheel.id === 'front' ? '#bae6fd' : '#fde68a'"/>
        <circle :cx="wheel.x - 54" cy="22" r="3" :fill="wheel.color"/><text :x="wheel.x - 46" y="25" :fill="wheel.color" font-size="9" font-weight="800">{{ wheel.id === 'front' ? t('guidesTirePressure.dashboard.frontStation') : t('guidesTirePressure.dashboard.rearStation') }}</text>
        <text :x="wheel.force.resultantX" y="49" text-anchor="middle" :fill="wheel.color" font-size="9" font-weight="800">{{ t('guidesTirePressure.dashboard.resultantLabel', { value: wheel.force.resultantKg.toFixed(1), g: wheel.force.gRatio.toFixed(2) }) }}</text>
        <line :x1="wheel.force.resultantX" y1="57" :x2="wheel.force.resultantX" y2="82" :stroke="wheel.color" stroke-width="2.5" :marker-end="wheel.id === 'front' ? 'url(#tire-load-arrow-front)' : 'url(#tire-load-arrow-rear)'"/>
        <line :x1="wheel.force.resultantX - 20" y1="67" :x2="wheel.force.resultantX - 20" y2="91" :stroke="wheel.color" stroke-width="1.5" stroke-dasharray="3 2"/>
        <text :x="wheel.force.resultantX - 24" y="83" text-anchor="end" :fill="wheel.color" font-size="8" font-weight="700">{{ t('guidesTirePressure.dashboard.verticalLoadLabel', { value: wheel.load.toFixed(1) }) }}</text>
      </g>
      <g>
        <rect v-if="wheel.force.demand > 0" :x="wheel.force.demandLabelX - 55" :y="wheel.force.demandLabelY - 10" width="110" height="14" rx="3" fill="#fff" fill-opacity=".98"/>
        <text v-if="wheel.force.demand > 0" :x="wheel.force.demandLabelX" :y="wheel.force.demandLabelY" text-anchor="middle" fill="#b91c1c" font-size="8.5" font-weight="700">{{ t(wheel.id === 'front' ? 'guidesTirePressure.dashboard.frontLateralDemandLabel' : 'guidesTirePressure.dashboard.rearLateralDemandLabel', { value: Math.round(wheel.force.demand) }) }}</text>
        <line v-if="wheel.force.demand > 0" :x1="wheel.force.demandStartX" :y1="wheel.force.axisY" :x2="wheel.force.demandEndX" :y2="wheel.force.axisY" stroke="#dc2626" stroke-width="2.5" marker-end="url(#tire-force-arrow)"/>
        <line :x1="wheel.x - wheel.force.gripLength" y1="248" :x2="wheel.x + wheel.force.gripLength" y2="248" stroke="#059669" stroke-width="2"/><line :x1="wheel.x - wheel.force.gripLength" y1="244" :x2="wheel.x - wheel.force.gripLength" y2="252" stroke="#059669" stroke-width="2"/><line :x1="wheel.x + wheel.force.gripLength" y1="244" :x2="wheel.x + wheel.force.gripLength" y2="252" stroke="#059669" stroke-width="2"/><line :x1="wheel.x" y1="245" :x2="wheel.x" y2="251" stroke="#059669" stroke-width="1" stroke-dasharray="1 1"/>
        <text :x="wheel.x" y="264" text-anchor="middle" fill="#047857" font-size="8" font-weight="700">{{ t('guidesTirePressure.dashboard.maxGripValue', { value: Math.round(wheel.force.maxGrip) }) }}</text>
      </g>
      <line x1="20" y1="220" x2="520" y2="220" stroke="#0f172a" stroke-width="2.5"/>
      <text :x="wheel.patch.cx" y="350" text-anchor="middle" :fill="wheel.color" font-size="9.5" font-weight="800">{{ t('guidesTirePressure.dashboard.contactValue', { value: wheel.area.toFixed(2) }) }}</text>
      <ellipse :cx="wheel.patch.cx" cy="376" :rx="wheel.patch.rx" :ry="wheel.patch.ry" :transform="`rotate(${wheel.patch.rotate}, ${wheel.patch.cx}, 376)`" :fill="wheel.id === 'front' ? 'url(#tire-patch-front)' : 'url(#tire-patch-rear)'" :stroke="wheel.color" stroke-width="1.5"/>
      <text :x="wheel.patch.cx" y="420" text-anchor="middle" :fill="wheel.color" font-size="8" font-weight="700">{{ t('guidesTirePressure.dashboard.contactPatchDimensions', { length: wheel.patch.lengthCm.toFixed(1), width: wheel.patch.widthCm.toFixed(1), state: wheel.patch.state }) }}</text>
    </g>
    <text x="30" y="242" fill="#64748b" font-size="9" font-weight="700">{{ t('guidesTirePressure.dashboard.roadLevel') }}</text><line x1="20" y1="320" x2="520" y2="320" stroke="#cbd5e1" stroke-width="1" stroke-dasharray="2 2"/><text x="270" y="336" text-anchor="middle" fill="#64748b" font-size="9" font-weight="700">{{ t('guidesTirePressure.dashboard.contactProjectionTitle') }}</text>
  </svg>
</template>

<script setup lang="ts">
import { inject } from 'vue'
const model = inject<any>('tirePressureModel')
if (!model) throw new Error('TirePressureCad requires tirePressureModel')
const { t, leanAngle, wheels } = model
</script>

<style scoped>
.cad-svg{width:100%;height:auto;margin-top:1rem;background:#fff;border:1px solid var(--line);border-radius:.8rem}
</style>
