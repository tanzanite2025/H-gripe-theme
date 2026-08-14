<template>
  <div class="wheelset-guide-subsection">
    <p class="wheelset-guide-step__body">
            Example riding scenarios where "Mixed Rim Profiles" (different front/rear specs) provide a tactical advantage.
            This approach balances aerodynamics, handling, and durability.
    </p>

    <div class="wheelset-guide-actions">
      <button
        type="button"
        class="wheelset-guide-button"
        @click="openQuickBuy"
      >
        Open Quick Custom Wheelset Builder
      </button>
      <button
        type="button"
        class="wheelset-guide-button wheelset-guide-button--solid"
        @click="openWhatsAppChat"
      >
        Consult a Wheelset Expert
      </button>
    </div>

    <div class="wheelset-guide-panel-grid">
      <div class="wheelset-guide-panel">
        <strong>Different models</strong>
        <p>
             Using completely different rim models (e.g., Lightweight Front + Heavy Duty Rear). Common in custom builds to balance specific performance needs.
        </p>
      </div>

      <div class="wheelset-guide-panel">
        <strong>Different heights</strong>
        <p>
             E.g., 50mm Front (handling) + 60mm Rear (aero/stiffness). The most common "Mixed" setup to optimize crosswind stability and speed.
        </p>
      </div>
    </div>

    <div>
      <h4 class="wheelset-guide-kicker">Configuration Cheatsheet</h4>

      <div class="wheelset-guide-panel-grid">
        <div class="wheelset-guide-panel">
          <strong>Climbing</strong>
          <p><span>Front:</span> 30-40mm. <span>Rear:</span> 40-50mm.</p>
          <p>Front for handling/weight, rear for stiffness.</p>
        </div>
        <div class="wheelset-guide-panel">
          <strong>Aerodynamics</strong>
          <p><span>Front:</span> 50-60mm. <span>Rear:</span> 60-80mm.</p>
          <p>Front balances crosswinds, rear maximizes speed.</p>
        </div>
        <div class="wheelset-guide-panel">
          <strong>Training / Durability</strong>
          <p><span>Front:</span> mid alloy/carbon. <span>Rear:</span> high/wide profile.</p>
          <p>Rear bears load, front preserves comfort.</p>
        </div>
        <div class="wheelset-guide-panel">
          <strong>Mixed / Versatile</strong>
          <p><span>Front:</span> 40-50mm. <span>Rear:</span> 55-65mm.</p>
          <p>A balanced choice for racing and training.</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { openQuickBuy, openWhatsAppChat } = defineProps<{
  openQuickBuy: () => void | Promise<void>
  openWhatsAppChat: () => void | Promise<void>
}>()
</script>

<style scoped>
.guide-section__cta-wrapper {
  color: var(--tz-text-primary);
}

.wheelset-link {
  color: #B5FF6D;
  text-decoration: underline;
}

.wheelset-link:hover {
  color: #60a5fa;
}

.wheelset-inline-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem 0.65rem;
  margin: 0 0.25rem;
  border-radius: 9999px;
  border: none;
  background: rgba(31, 41, 55, 0.65);
  color: var(--tz-text-primary);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  box-shadow:
    0 10px 22px rgba(0, 0, 0, 0.85);
  transition: background-color 0.15s ease, transform 0.08s ease;
}

.wheelset-inline-button:hover {
  background: rgba(51, 65, 85, 0.75);
}

.wheelset-inline-button:active {
  transform: scale(0.98);
}

.wheelset-guide-kicker {
  margin: 0 0 0.8rem;
  color: var(--tz-text-muted);
  font-size: var(--tz-type-micro-label);
  font-weight: 700;
  letter-spacing: 0.12em;
  line-height: 1.25;
  text-transform: uppercase;
}

.wheelset-guide-panel span {
  color: var(--tz-text-muted);
  font-weight: 650;
}
</style>
