<template>
  <aside class="schwalbe-telemetry" aria-labelledby="schwalbe-telemetry-title">
    <div>
      <p class="schwalbe-telemetry__kicker">{{ tx('telemetry.kicker') }}</p>
      <h2 id="schwalbe-telemetry-title">{{ tx('telemetry.title') }}</h2>
    </div>
    <p>{{ tx('telemetry.body') }}</p>
    <ul>
      <li>{{ tx('telemetry.points.catalog') }}</li>
      <li>{{ tx('telemetry.points.sales') }}</li>
      <li>{{ tx('telemetry.points.compatibility') }}</li>
    </ul>
  </aside>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'

const { t: translate } = useI18n()
const tx = (key: string, params?: Record<string, unknown>) => translate(`guidesSchwalbeTireSelector.${key}`, params || {})
</script>

<style scoped>
.schwalbe-telemetry {
  display: grid;
  grid-template-columns: minmax(12rem, 0.7fr) minmax(16rem, 1fr) minmax(16rem, 1fr);
  gap: 1rem;
  align-items: start;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  background: var(--tz-surface-subtle);
  padding: 1rem;
}

.schwalbe-telemetry__kicker {
  margin: 0 0 0.25rem;
  color: var(--tz-text-accent);
  font-size: 0.7rem;
  font-weight: 750;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.schwalbe-telemetry h2 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1rem;
  line-height: 1.35;
}

.schwalbe-telemetry p,
.schwalbe-telemetry li {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
  line-height: 1.65;
}

.schwalbe-telemetry ul {
  display: grid;
  gap: 0.35rem;
  margin: 0;
  padding-left: 1rem;
}

@media (max-width: 800px) {
  .schwalbe-telemetry {
    grid-template-columns: 1fr;
  }
}
</style>
