<template>
  <article class="station controls">
    <header class="controls-header"><b>{{ t('guidesTirePressure.dashboard.controlsTitle') }}</b><span>{{ t('guidesTirePressure.dashboard.totalValue', { value: totalWeight.toFixed(1) }) }}</span></header>
    <div class="controls-grid">
    <div class="control-box control-box--wide">
      <b>{{ t('guidesTirePressure.dashboard.fixedPressure') }}</b>
      <span>{{ t('guidesTirePressure.dashboard.fixedPressureHint') }}</span>
    </div>
    <div class="control-box"><div class="control-head"><b>{{ t('guidesTirePressure.dashboard.frontPressure') }}</b><span>{{ fixedFrontPsi.toFixed(1) }} PSI</span></div><input id="tire-dashboard-fixed-front-pressure" v-model.number="fixedFrontPsi" :aria-label="t('guidesTirePressure.dashboard.frontPressure')" type="range" min="25" max="100" step="0.5"><div class="range-scale"><span>25 PSI</span><span>50 PSI</span><span>75 PSI</span><span>100 PSI</span></div></div>
    <div class="control-box"><div class="control-head"><b>{{ t('guidesTirePressure.dashboard.rearPressure') }}</b><span>{{ fixedRearPsi.toFixed(1) }} PSI</span></div><input id="tire-dashboard-fixed-rear-pressure" v-model.number="fixedRearPsi" :aria-label="t('guidesTirePressure.dashboard.rearPressure')" type="range" min="25" max="100" step="0.5"><div class="range-scale"><span>25 PSI</span><span>50 PSI</span><span>75 PSI</span><span>100 PSI</span></div></div>
    <div class="control-box control-box--wide"><div class="control-head"><b>{{ t('guidesTirePressure.dashboard.leanAngle') }}</b><span>{{ leanAngle }}° · {{ leanDesc }}</span></div><div class="presets lean-presets"><button v-for="preset in leanPresets" :key="preset" type="button" :aria-pressed="leanAngle === preset" :class="{ active: leanAngle === preset }" @click="leanAngle = preset">{{ preset }}° · {{ leanPresetLabel(preset) }}</button></div></div>
    <div class="control-box"><div class="control-head"><b>{{ t('guidesTirePressure.dashboard.speed') }}</b><span>{{ speedKmh.toFixed(0) }} km/h</span></div><input id="tire-dashboard-speed" v-model.number="speedKmh" :aria-label="t('guidesTirePressure.dashboard.speed')" type="range" min="0" max="80" step="1"><div class="range-scale"><span>0</span><span>20</span><span>40</span><span>80 km/h</span></div><small class="speed-hint">{{ t('guidesTirePressure.dashboard.speedHint') }}</small><div class="speed-readout"><span>{{ t('guidesTirePressure.dashboard.lateralAcceleration') }}</span><strong>{{ lateralAccelerationG === null ? '—' : lateralAccelerationG.toFixed(2) }} G</strong><span>{{ t('guidesTirePressure.dashboard.turnRadius') }}</span><strong>{{ formatTurnRadius(equivalentTurnRadiusM) }}</strong></div></div>
    <div class="control-box"><div class="control-head"><b>{{ t('guidesTirePressure.dashboard.riderWeight') }}</b><span>{{ riderWeight.toFixed(1) }} kg</span></div><input id="tire-dashboard-rider-weight" v-model.number="riderWeight" :aria-label="t('guidesTirePressure.dashboard.riderWeight')" type="range" min="45" max="115" step="0.5"><div class="range-scale"><span>45 kg</span><span>70 kg</span><span>90 kg</span><span>115 kg</span></div></div>
    <div class="control-box"><div class="control-head"><b>{{ t('guidesTirePressure.dashboard.bikeGear') }}</b><span>{{ bikeWeight.toFixed(1) }} kg</span></div><input id="tire-dashboard-bike-weight" v-model.number="bikeWeight" :aria-label="t('guidesTirePressure.dashboard.bikeGear')" type="range" min="6" max="18" step="0.5"><div class="range-scale"><span>6 kg</span><span>8.5 kg</span><span>12 kg</span><span>18 kg</span></div></div>
    <div class="control-box">
      <b>{{ t('guidesTirePressure.dashboard.posture') }}</b>
      <span>{{ postureLabel }}</span>
      <div class="load-distribution" aria-live="polite">
        <div class="load-distribution__item">
          <span>{{ t('guidesTirePressure.dashboard.frontWheel') }}</span>
          <strong>{{ frontLoad === null ? '—' : frontLoad.toFixed(1) + ' kg' }}</strong>
          <small>{{ frontRatio === null ? '—' : frontRatio + '%' }}</small>
        </div>
        <div class="load-distribution__item">
          <span>{{ t('guidesTirePressure.dashboard.rearWheel') }}</span>
          <strong>{{ rearLoad === null ? '—' : rearLoad.toFixed(1) + ' kg' }}</strong>
          <small>{{ rearRatio === null ? '—' : rearRatio + '%' }}</small>
        </div>
      </div>
      <div class="presets"><button v-for="posture in postures" :key="posture.id" type="button" :class="{ active: posture.id === postureId }" @click="postureId = posture.id">{{ posture.label }}</button></div>
    </div>
    <div class="control-box">
      <b>{{ t('guidesTirePressure.dashboard.roadBaselineTitle') }}</b>
      <span>{{ t('guidesTirePressure.dashboard.flatRoadBaseline') }}</span>
      <div class="friction-baseline" role="note" aria-live="polite">
        <div class="friction-baseline__head">
          <b>{{ t('guidesTirePressure.dashboard.frictionBaselineTitle') }}</b>
          <strong>{{ t('guidesTirePressure.dashboard.fixedMuLabel', { value: formatFrictionCoefficient(demonstrationFrictionCoefficientNominal) }) }}</strong>
        </div>
        <small>{{ t('guidesTirePressure.dashboard.fixedMuHint') }}</small>
      </div>
    </div>
    <div class="control-box"><div class="control-head"><b>{{ t('guidesTirePressure.dashboard.tireWidth') }}</b><span>{{ tireWidth }}C</span></div><input id="tire-dashboard-tire-width" v-model.number="tireWidth" :aria-label="t('guidesTirePressure.dashboard.tireWidth')" type="range" min="25" max="40" step="1"><div class="range-scale"><span>25C</span><span>28C</span><span>32C</span><span>40C</span></div></div>
    </div>
  </article>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import { inject, type ComputedRef, type Ref } from 'vue'

type TirePressureControlsModel = {
  totalWeight: ComputedRef<number>
  leanAngle: Ref<number>
  leanDesc: ComputedRef<string>
  leanPresets: number[]
  leanPresetLabel: (preset: number) => string
  riderWeight: Ref<number>
  bikeWeight: Ref<number>
  frontLoad: ComputedRef<number | null>
  rearLoad: ComputedRef<number | null>
  frontRatio: ComputedRef<number | null>
  rearRatio: ComputedRef<number | null>
  postureLabel: ComputedRef<string>
  postures: ComputedRef<Array<{ id: string; key: string; label: string }>>
  postureId: Ref<string>
  tireWidth: Ref<number>
  fixedFrontPsi: Ref<number>
  fixedRearPsi: Ref<number>
  speedKmh: Ref<number>
  equivalentTurnRadiusM: ComputedRef<number | null>
  lateralAccelerationG: ComputedRef<number | null>
  formatTurnRadius: (value: number | null) => string
  demonstrationFrictionCoefficientNominal: ComputedRef<number | null>
}

const { t } = useI18n()
const model = inject<TirePressureControlsModel>('tirePressureModel')
if (!model) throw new Error('TirePressureControls requires tirePressureModel')
const { totalWeight, leanAngle, leanDesc, leanPresets, leanPresetLabel, riderWeight, bikeWeight, frontLoad, rearLoad, frontRatio, rearRatio, postureLabel, postures, postureId, tireWidth, fixedFrontPsi, fixedRearPsi, speedKmh, equivalentTurnRadiusM, lateralAccelerationG, formatTurnRadius, demonstrationFrictionCoefficientNominal } = model

function formatFrictionCoefficient(value: number | null | undefined) {
  return value === null || value === undefined || !Number.isFinite(value) ? '—' : value.toFixed(2)
}

</script>

<style scoped>
.controls {
  min-width: 0;
}

.controls-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 0.65rem;
  color: var(--tz-text-primary);
  font-size: 0.9rem;
}

.controls-header span {
  color: var(--tz-text-secondary);
  font-size: 0.72rem;
  font-variant-numeric: tabular-nums;
}

.controls-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.55rem;
}

.control-box {
  min-width: 0;
  margin: 0;
  padding: 0.58rem 0.65rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.65rem;
  background: var(--tz-surface-muted);
  color: var(--tz-text-primary);
}

.control-box--wide {
  grid-column: 1 / -1;
}

.control-box > b {
  display: block;
  font-size: 0.76rem;
  line-height: 1.25;
}

.control-box > span {
  display: block;
  margin-top: 0.18rem;
  overflow: hidden;
  color: var(--tz-text-secondary);
  font-size: 0.68rem;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.friction-baseline {
  margin-top: 0.5rem;
  padding: 0.45rem 0.5rem;
  border: 1px dashed var(--tz-border-strong);
  border-radius: 0.45rem;
  background: var(--tz-card-surface);
}

.friction-baseline__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.45rem;
}

.friction-baseline__head b {
  color: var(--tz-text-secondary);
  font-size: 0.64rem;
  line-height: 1.25;
}

.friction-baseline__head strong {
  color: var(--tz-text-primary);
  font-size: 0.72rem;
  font-variant-numeric: tabular-nums;
  line-height: 1.2;
  white-space: nowrap;
}

.friction-baseline small {
  display: block;
  margin-top: 0.25rem;
  color: var(--tz-text-secondary);
  font-size: 0.59rem;
  line-height: 1.35;
}


.load-distribution {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.35rem;
  margin-top: 0.48rem;
}

.load-distribution__item {
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: baseline;
  gap: 0.1rem 0.35rem;
  min-width: 0;
  padding: 0.4rem 0.45rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.45rem;
  background: var(--tz-card-surface);
}

.load-distribution__item > span {
  grid-column: 1 / -1;
  overflow: hidden;
  color: var(--tz-text-secondary);
  font-size: 0.62rem;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.load-distribution__item strong {
  color: var(--tz-text-primary);
  font-size: 0.76rem;
  font-variant-numeric: tabular-nums;
  line-height: 1.2;
}

.load-distribution__item small {
  color: var(--tz-text-secondary);
  font-size: 0.62rem;
  font-variant-numeric: tabular-nums;
  line-height: 1.2;
}

.control-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem;
  font-size: 0.76rem;
  line-height: 1.25;
}

.control-head span {
  color: var(--tz-text-secondary);
  font-size: 0.68rem;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.control-box input[type='range'] {
  display: block;
  width: 100%;
  height: 1.05rem;
  margin: 0.2rem 0 0;
  accent-color: var(--tz-text-primary);
  cursor: pointer;
}

.range-scale {
  display: flex;
  justify-content: space-between;
  gap: 0.25rem;
  margin-top: 0.08rem;
  color: var(--tz-text-secondary);
  font-size: 0.58rem;
  line-height: 1.2;
}

.speed-readout {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 0.15rem 0.5rem;
  margin-top: 0.42rem;
  color: var(--tz-text-secondary);
  font-size: 0.61rem;
  line-height: 1.25;
}

.speed-hint {
  display: block;
  margin-top: 0.2rem;
  color: var(--tz-text-secondary);
  font-size: 0.59rem;
  line-height: 1.35;
}

.speed-readout strong {
  color: var(--tz-text-primary);
  font-variant-numeric: tabular-nums;
  text-align: right;
}

.presets {
  display: flex;
  flex-wrap: wrap;
  gap: 0.28rem;
  margin-top: 0.38rem;
}

.presets button {
  min-height: 1.7rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.4rem;
  background: var(--tz-card-surface);
  color: var(--tz-text-primary);
  cursor: pointer;
  padding: 0.26rem 0.42rem;
  font-size: 0.66rem;
  line-height: 1.15;
  white-space: nowrap;
}

.presets button:hover,
.presets button:focus-visible {
  border-color: var(--tz-site-accent);
  outline: none;
}

.presets button.active {
  border-color: var(--tz-text-primary);
  background: var(--tz-text-primary);
  color: var(--tz-card-surface);
}

@media (max-width: 600px) {
  .controls-grid {
    grid-template-columns: 1fr;
  }

  .control-box--wide {
    grid-column: auto;
  }

}
</style>
