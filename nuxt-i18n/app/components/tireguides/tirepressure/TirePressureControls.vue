<template>
  <article class="station controls">
    <header><b>{{ t('guidesTirePressure.dashboard.controlsTitle') }}</b><span>{{ t('guidesTirePressure.dashboard.totalValue', { value: totalWeight.toFixed(1) }) }}</span></header>
    <div class="control-box">
      <b>{{ t('guidesTirePressure.dashboard.pressureMode') }}</b>
      <div class="presets">
        <button :class="{ active: pressureMode === 'auto' }" @click="pressureMode = 'auto'">{{ t('guidesTirePressure.dashboard.autoPressure') }}</button>
        <button :class="{ active: pressureMode === 'fixed' }" @click="pressureMode = 'fixed'">{{ t('guidesTirePressure.dashboard.fixedPressure') }}</button>
      </div>
      <span>{{ t('guidesTirePressure.dashboard.fixedPressureHint') }}</span>
    </div>
    <template v-if="pressureMode === 'fixed'">
      <div class="control-box"><div class="control-head"><b>{{ t('guidesTirePressure.dashboard.frontPressure') }}</b><span>{{ fixedFrontPsi.toFixed(1) }} PSI</span></div><input id="tire-dashboard-fixed-front-pressure" v-model.number="fixedFrontPsi" :aria-label="t('guidesTirePressure.dashboard.frontPressure')" type="range" min="25" max="100" step="0.5"><div class="range-scale"><span>25 PSI</span><span>50 PSI</span><span>75 PSI</span><span>100 PSI</span></div></div>
      <div class="control-box"><div class="control-head"><b>{{ t('guidesTirePressure.dashboard.rearPressure') }}</b><span>{{ fixedRearPsi.toFixed(1) }} PSI</span></div><input id="tire-dashboard-fixed-rear-pressure" v-model.number="fixedRearPsi" :aria-label="t('guidesTirePressure.dashboard.rearPressure')" type="range" min="25" max="100" step="0.5"><div class="range-scale"><span>25 PSI</span><span>50 PSI</span><span>75 PSI</span><span>100 PSI</span></div></div>
    </template>
    <div class="control-box"><div class="control-head"><b>{{ t('guidesTirePressure.dashboard.leanAngle') }}</b><span>{{ leanAngle }}° · {{ leanDesc }}</span></div><input id="tire-dashboard-lean-angle" v-model.number="leanAngle" :aria-label="t('guidesTirePressure.dashboard.leanAngle')" type="range" min="0" max="45" step="1"><div class="range-scale"><span>0°</span><span>15°</span><span>30°</span><span>45°</span></div></div>
    <div class="presets"><button v-for="preset in leanPresets" :key="preset" :class="{ active: leanAngle === preset }" @click="leanAngle = preset">{{ preset }}° · {{ leanPresetLabel(preset) }}</button></div>
    <div class="control-box"><div class="control-head"><b>{{ t('guidesTirePressure.dashboard.riderWeight') }}</b><span>{{ riderWeight.toFixed(1) }} kg</span></div><input id="tire-dashboard-rider-weight" v-model.number="riderWeight" :aria-label="t('guidesTirePressure.dashboard.riderWeight')" type="range" min="45" max="115" step="0.5"><div class="range-scale"><span>45 kg</span><span>70 kg</span><span>90 kg</span><span>115 kg</span></div></div>
    <div class="control-box"><div class="control-head"><b>{{ t('guidesTirePressure.dashboard.bikeGear') }}</b><span>{{ bikeWeight.toFixed(1) }} kg</span></div><input id="tire-dashboard-bike-weight" v-model.number="bikeWeight" :aria-label="t('guidesTirePressure.dashboard.bikeGear')" type="range" min="6" max="18" step="0.5"><div class="range-scale"><span>6 kg</span><span>8.5 kg</span><span>12 kg</span><span>18 kg</span></div></div>
    <div class="control-box"><b>{{ t('guidesTirePressure.dashboard.posture') }}</b><span>{{ postureLabel }}</span><div class="presets"><button v-for="posture in postures" :key="posture.id" :class="{ active: posture.id === postureId }" @click="postureId = posture.id">{{ posture.label }}</button></div></div>
    <div class="control-box"><b>{{ t('guidesTirePressure.dashboard.surface') }}</b><span>{{ surfaceLabel }}</span><div class="presets"><button v-for="surface in surfaces" :key="surface.id" :class="{ active: surface.id === surfaceId }" @click="surfaceId = surface.id">{{ surface.label }}</button></div></div>
    <div class="split"><div class="control-box"><b>{{ t('guidesTirePressure.dashboard.weather') }}</b><div class="presets"><button :class="{ active: weather === 'dry' }" @click="weather = 'dry'">{{ t('guidesTirePressure.dashboard.dry') }}</button><button :class="{ active: weather === 'wet' }" @click="weather = 'wet'">{{ t('guidesTirePressure.dashboard.wet') }}</button></div></div><div class="control-box"><b>{{ t('guidesTirePressure.dashboard.rimSystem') }}</b><div class="presets"><button :class="{ active: rimSystem === 'hookless' }" @click="rimSystem = 'hookless'">{{ t('guidesTirePressure.dashboard.hookless') }}</button><button :class="{ active: rimSystem === 'hooked' }" @click="rimSystem = 'hooked'">{{ t('guidesTirePressure.dashboard.hooked') }}</button></div></div></div>
    <div class="control-box"><div class="control-head"><b>{{ t('guidesTirePressure.dashboard.tireWidth') }}</b><span>{{ tireWidth }}C</span></div><input id="tire-dashboard-tire-width" v-model.number="tireWidth" :aria-label="t('guidesTirePressure.dashboard.tireWidth')" type="range" min="25" max="40" step="1"><div class="range-scale"><span>25C</span><span>28C</span><span>32C</span><span>40C</span></div></div>
    <div class="control-box"><div class="control-head"><b>{{ t('guidesTirePressure.calculator.rimWidth') }}</b><span>{{ rimWidth }} mm</span></div><input id="tire-dashboard-rim-width" v-model.number="rimWidth" :aria-label="t('guidesTirePressure.calculator.rimWidth')" type="range" min="19" max="30" step="1"><div class="range-scale"><span>19 mm</span><span>23 mm</span><span>25 mm</span><span>30 mm</span></div></div>
  </article>
</template>

<script setup lang="ts">
import { inject } from 'vue'

const model = inject<any>('tirePressureModel')
if (!model) throw new Error('TirePressureControls requires tirePressureModel')
const { t, totalWeight, leanAngle, leanDesc, leanPresets, leanPresetLabel, riderWeight, bikeWeight, postureLabel, postures, postureId, surfaceLabel, surfaces, surfaceId, weather, rimSystem, tireWidth, rimWidth, pressureMode, fixedFrontPsi, fixedRearPsi } = model
</script>

<style scoped>
.control-box{background:var(--soft);border:1px solid var(--line);border-radius:.8rem;padding:.7rem .85rem;margin-top:.7rem}.control-head{display:flex;justify-content:space-between;gap:.5rem;font-size:.8rem}.control-head span,.control-box>span{color:var(--muted);font-size:.75rem}.control-box input[type=range]{width:100%;accent-color:#09090b}.range-scale{display:flex;justify-content:space-between;margin-top:.25rem;color:#64748b;font-size:.62rem}.presets{display:flex;flex-wrap:wrap;gap:.35rem;margin-top:.5rem}.presets button{border:1px solid #cbd5e1;background:#fff;border-radius:.4rem;padding:.35rem .5rem;font-size:.75rem;cursor:pointer}.presets button.active{background:#09090b;color:#fff;border-color:#09090b}.split{display:grid;grid-template-columns:1fr 1fr;gap:.7rem}@media(max-width:600px){.control-box{padding:.55rem .65rem;margin-top:.5rem}.control-head{font-size:.75rem}.control-head span,.control-box>span{font-size:.68rem}.range-scale{font-size:.57rem}.presets{gap:.25rem;margin-top:.35rem}.presets button{padding:.3rem .4rem;font-size:.68rem}.split{gap:.5rem}}
</style>
