<template>
  <figure class="spoke-wheel">
    <div class="spoke-wheel__header">
      <p class="spoke-wheel__eyebrow">{{ wheelLabel }}</p>
      <h5 class="spoke-wheel__title">{{ sideLabel }}</h5>
    </div>

    <svg
      viewBox="0 0 420 320"
      class="spoke-wheel__svg"
      role="img"
      :aria-label="`${wheelLabel} ${sideLabel} schematic`"
    >
      <title>{{ wheelLabel }} {{ sideLabel }} schematic</title>
      <desc>
        Full wheel vector view with rim, hub, spoke pattern, and numbered dimension callouts for ERD,
        PCDL, PCDR, and rim offset.
      </desc>

      <defs>
        <marker
          :id="arrowMarkerId"
          markerWidth="10"
          markerHeight="10"
          refX="5"
          refY="5"
          orient="auto"
          markerUnits="strokeWidth"
        >
          <path d="M0,0 L10,5 L0,10 z" fill="currentColor" />
        </marker>
      </defs>

      <rect x="0" y="0" width="420" height="320" fill="#fff" />

      <circle cx="210" cy="156" r="126" class="spoke-wheel__rim spoke-wheel__rim--outer" />
      <circle cx="210" cy="156" r="104" class="spoke-wheel__rim spoke-wheel__rim--inner" />

      <g class="spoke-wheel__spokes">
        <line v-for="spoke in spokeLines" :key="spoke.id" :x1="spoke.x1" :y1="spoke.y1" :x2="spoke.x2" :y2="spoke.y2" class="spoke-wheel__spoke" />
      </g>

      <g class="spoke-wheel__hub">
        <ellipse cx="176" cy="156" rx="18" ry="42" class="spoke-wheel__flange" />
        <rect x="176" y="138" width="68" height="36" rx="14" class="spoke-wheel__hub-body" />
        <ellipse cx="244" cy="156" rx="18" ry="42" class="spoke-wheel__flange" />
        <circle cx="210" cy="156" r="13" class="spoke-wheel__hub-core" />
      </g>

      <g v-if="isDiscSide">
        <circle cx="146" cy="156" r="20" class="spoke-wheel__disc-ring" />
        <circle cx="146" cy="156" r="8" class="spoke-wheel__disc-core" />
        <text x="146" y="191" class="spoke-wheel__side-mark">Disc</text>
      </g>
      <g v-else>
        <circle cx="274" cy="156" r="18" class="spoke-wheel__freehub-ring" />
        <circle cx="274" cy="156" r="7" class="spoke-wheel__freehub-core" />
        <text x="274" y="191" class="spoke-wheel__side-mark">{{ sideMarkLabel }}</text>
      </g>

      <rect x="160" y="126" width="100" height="60" rx="14" class="spoke-wheel__hub-zoom" />
      <text x="210" y="146" class="spoke-wheel__hub-zoom-label">Hub / flange zone</text>

      <g class="spoke-wheel__callout">
        <line x1="84" y1="54" x2="336" y2="54" class="spoke-wheel__dimension" :marker-start="`url(#${arrowMarkerId})`" :marker-end="`url(#${arrowMarkerId})`" />
        <text x="210" y="36" class="spoke-wheel__callout-number">1</text>
        <text x="210" y="69" class="spoke-wheel__callout-label">ERD</text>
      </g>

      <g class="spoke-wheel__callout">
        <line x1="96" y1="106" x2="172" y2="138" class="spoke-wheel__dimension" :marker-end="`url(#${arrowMarkerId})`" />
        <text x="92" y="103" class="spoke-wheel__callout-number">2</text>
        <text x="90" y="120" class="spoke-wheel__callout-label">PCDL</text>
      </g>

      <g class="spoke-wheel__callout">
        <line x1="324" y1="106" x2="248" y2="138" class="spoke-wheel__dimension" :marker-end="`url(#${arrowMarkerId})`" />
        <text x="328" y="103" class="spoke-wheel__callout-number">3</text>
        <text x="330" y="120" class="spoke-wheel__callout-label spoke-wheel__callout-label--right">PCDR</text>
      </g>

      <g class="spoke-wheel__callout">
        <line x1="166" y1="272" x2="254" y2="272" class="spoke-wheel__dimension" :marker-start="`url(#${arrowMarkerId})`" :marker-end="`url(#${arrowMarkerId})`" />
        <text x="210" y="255" class="spoke-wheel__callout-number">4</text>
        <text x="210" y="289" class="spoke-wheel__callout-label">Offset</text>
      </g>
    </svg>
  </figure>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type WheelKind = 'front' | 'rear'
type WheelSide = 'disc' | 'non-disc'

const props = defineProps<{
  wheel: WheelKind
  side: WheelSide
}>()

const wheelLabel = computed(() => (props.wheel === 'front' ? 'Front wheel' : 'Rear wheel'))
const sideLabel = computed(() => (props.side === 'disc' ? 'Disc side' : 'Non-disc side'))
const sideMarkLabel = computed(() => {
  if (props.side === 'disc') return 'Disc'
  return props.wheel === 'front' ? 'Non-disc' : 'NDS'
})
const isDiscSide = computed(() => props.side === 'disc')
const arrowMarkerId = computed(() => `spoke-wheel-${props.wheel}-${props.side}-arrow`)

const spokeLines = [
  { id: 'spoke-1', x1: 176, y1: 156, x2: 110, y2: 92 },
  { id: 'spoke-2', x1: 176, y1: 156, x2: 86, y2: 156 },
  { id: 'spoke-3', x1: 176, y1: 156, x2: 110, y2: 220 },
  { id: 'spoke-4', x1: 176, y1: 156, x2: 148, y2: 248 },
  { id: 'spoke-5', x1: 244, y1: 156, x2: 310, y2: 92 },
  { id: 'spoke-6', x1: 244, y1: 156, x2: 334, y2: 156 },
  { id: 'spoke-7', x1: 244, y1: 156, x2: 310, y2: 220 },
  { id: 'spoke-8', x1: 244, y1: 156, x2: 272, y2: 248 },
]
</script>

<style scoped>
.spoke-wheel {
  display: grid;
  gap: 0.6rem;
  min-width: 0;
}

.spoke-wheel__header {
  display: flex;
  min-width: 0;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.6rem;
}

.spoke-wheel__eyebrow {
  margin: 0;
  color: rgba(5, 150, 105, 0.82);
  font-size: 0.62rem;
  font-weight: 800;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.spoke-wheel__title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.9rem;
  font-weight: 750;
  line-height: 1.2;
  text-align: right;
}

.spoke-wheel__svg {
  display: block;
  width: 100%;
  height: auto;
}

.spoke-wheel__rim {
  fill: none;
}

.spoke-wheel__rim--outer {
  stroke: #2b2b31;
  stroke-width: 26;
}

.spoke-wheel__rim--inner {
  stroke: rgba(148, 163, 184, 0.5);
  stroke-width: 2;
}

.spoke-wheel__spoke {
  fill: none;
  stroke: rgba(31, 41, 55, 0.72);
  stroke-width: 2;
  stroke-linecap: round;
}

.spoke-wheel__hub-body {
  fill: #d1d5db;
  stroke: #6b7280;
  stroke-width: 1.5;
}

.spoke-wheel__flange {
  fill: #374151;
  stroke: #111827;
  stroke-width: 1.2;
}

.spoke-wheel__hub-core {
  fill: #f8fafc;
  stroke: #6b7280;
  stroke-width: 1.2;
}

.spoke-wheel__disc-ring {
  fill: none;
  stroke: #ef4444;
  stroke-width: 4;
}

.spoke-wheel__disc-core {
  fill: rgba(239, 68, 68, 0.14);
}

.spoke-wheel__freehub-ring {
  fill: none;
  stroke: #94a3b8;
  stroke-width: 4;
}

.spoke-wheel__freehub-core {
  fill: rgba(226, 232, 240, 0.95);
}

.spoke-wheel__hub-zoom {
  fill: rgba(239, 68, 68, 0.04);
  stroke: #ef4444;
  stroke-width: 2;
}

.spoke-wheel__hub-zoom-label {
  fill: #ef4444;
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-anchor: middle;
  text-transform: uppercase;
}

.spoke-wheel__side-mark {
  fill: rgba(100, 116, 139, 0.92);
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-anchor: middle;
  text-transform: uppercase;
}

.spoke-wheel__dimension {
  fill: none;
  stroke: #ef4444;
  stroke-width: 2;
}

.spoke-wheel__callout-number {
  fill: #ef4444;
  font-size: 12px;
  font-weight: 900;
  text-anchor: middle;
}

.spoke-wheel__callout-label {
  fill: #ef4444;
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.1em;
  text-anchor: middle;
  text-transform: uppercase;
}

.spoke-wheel__callout-label--right {
  text-anchor: start;
}

@media (max-width: 767px) {
  .spoke-wheel__svg {
    width: calc(100% + 0.5rem);
    margin-inline: -0.25rem;
  }
}
</style>
