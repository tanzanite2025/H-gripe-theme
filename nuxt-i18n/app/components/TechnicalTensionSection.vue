<template>
  <div class="tension-section space-y-8">
    <!-- Intro Card -->
    <div class="tension-panel tension-panel--intro p-5 md:p-6 text-center">
      <h3 class="tension-title mb-2">Wheel Spoke Tension Standard</h3>
      <p class="tension-copy max-w-2xl mx-auto mb-6">
        Tension is the soul of a wheel. We customize tension based on
        <strong>Rider Weight</strong>, <strong>Usage</strong>, and <strong>Rim Material</strong>
        to balance stiffness, durability, and comfort.
       </p>
      <div class="flex justify-center flex-col items-center gap-4">
        <GuideImage
          src="/public/technical/tension/wheel-spoke-tension.webp"
          alt="Wheel Spoke Tension Distribution"
          :zoomOnClick="true"
          caption="Workshop: Tension Distribution Reference"
          class="tension-image rounded-xl max-w-lg w-full"
        />
        <NuxtLink
          :to="localePath('/support/test-report/wheelset-assembly')"
          class="premium-button tension-link"
        >
          Check the wheelset assembly guide
        </NuxtLink>
      </div>
    </div>

    <!-- 1. Key Factors Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- Rider Weight -->
      <div class="tension-panel tension-panel--compact p-5 md:p-6">
        <h3 class="tension-title tension-title--section">
          Effect of Rider Weight
        </h3>
        <ul class="tension-list space-y-3 text-sm">
          <li class="tension-list__item flex items-start gap-3">
            <span class="tension-list__marker mt-1 shrink-0" />
            <span><strong>Static Load:</strong> A 75kg rider increases upper spoke tension by ~15%, while lower spokes lose tension. (De-tensioning cycle).</span>
          </li>
          <li class="tension-list__item flex items-start gap-3">
            <span class="tension-list__marker mt-1 shrink-0" />
            <span><strong>Heavy Riders:</strong> Need higher base tension and thicker spokes (14G) to prevent spokes from going slack, which causes fatigue failure.</span>
          </li>
        </ul>
      </div>

      <!-- Usage Scenarios -->
      <div class="tension-panel tension-panel--compact p-5 md:p-6">
        <h3 class="tension-title tension-title--section">
          Standard by Usage
        </h3>
        <div class="tension-options space-y-3 text-sm">
          <div class="tension-option p-2 rounded-lg">
            <strong class="tension-option__label block">Road Racing</strong>
            <span class="tension-option__copy text-xs">High stiffness focus. 120-130kgf (Drive Side).</span>
          </div>
          <div class="tension-option p-2 rounded-lg">
            <strong class="tension-option__label block">MTB / Gravel</strong>
            <span class="tension-option__copy text-xs">Durability focus. Slightly lower tension for compliance and impact absorption.</span>
          </div>
          <div class="tension-option p-2 rounded-lg">
            <strong class="tension-option__label block">Touring / Commuting</strong>
            <span class="tension-option__copy text-xs">Reliability focus. Moderate tension, thicker spokes.</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 2. Tension Chart -->
    <div class="tension-panel tension-table-panel overflow-hidden">
      <div class="tension-table-header p-4 flex items-center justify-between">
        <h3 class="font-bold tz-text-primary">Recommended Drive-Side Tension (kgf)</h3>
        <span class="tension-unit text-xs font-mono">Unit: kgf</span>
      </div>
      <div class="overflow-x-auto">
        <table class="tension-table min-w-full text-left text-sm">
          <thead>
            <tr>
              <th>Rider Weight</th>
              <th>Road Racing</th>
              <th>MTB</th>
              <th>Touring</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td class="tension-table__row-label">≤ 60 kg</td>
              <td>115–120</td>
              <td>95–105</td>
              <td>90–100</td>
            </tr>
            <tr>
              <td class="tension-table__row-label">60–80 kg</td>
              <td>120–125</td>
              <td>100–115</td>
              <td>95–105</td>
            </tr>
            <tr>
              <td class="tension-table__row-label">80–100 kg</td>
              <td>125–130</td>
              <td>110–120</td>
              <td>100–110</td>
            </tr>
            <tr>
              <td class="tension-table__row-label">≥ 100 kg</td>
              <td>130–135</td>
              <td>115–125</td>
              <td>105–115</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="tension-note text-xs">
        <strong>Note:</strong> These are baseline values for Drive-Side spokes. Non-Drive Side tension is typically 60-70% of Drive Side, determined by hub geometry. Max tension for Carbon rims is usually ~135kgf; Alloy rims ~125kgf.
      </div>
    </div>

    <!-- 3. Workshop Process (Timeline) -->
    <div class="tension-panel tension-panel--process p-5 md:p-6">
      <h3 class="tension-title tension-title--section mb-6">
        Our Workshop Building Process
      </h3>

      <div class="tension-timeline relative pl-6 space-y-8">
        <!-- Step 1 -->
        <div class="tension-step tension-step--one relative">
          <span class="tension-step__marker absolute">
            <span class="tension-step__dot block" />
          </span>
          <h4 class="text-sm font-bold tz-text-primary">1. Baseline Calculation</h4>
          <p class="text-xs tz-text-secondary mt-1">Determine base tension range based on Rider Weight classification.</p>
        </div>
        <!-- Step 2 -->
        <div class="tension-step tension-step--two relative">
          <span class="tension-step__marker absolute">
            <span class="tension-step__dot block" />
          </span>
          <h4 class="text-sm font-bold tz-text-primary">2. Usage Adjustment</h4>
          <p class="text-xs tz-text-secondary mt-1">Adjust for stiffness (Road) or compliance (MTB).</p>
        </div>
        <!-- Step 3 -->
        <div class="tension-step tension-step--three relative">
          <span class="tension-step__marker absolute">
            <span class="tension-step__dot block" />
          </span>
          <h4 class="text-sm font-bold tz-text-primary">3. Safety Cap Check</h4>
          <p class="text-xs tz-text-secondary mt-1">Ensure target does not exceed Rim Limits (Carbon 135kgf / Alloy 125kgf).</p>
        </div>
        <!-- Step 4 -->
        <div class="tension-step tension-step--four relative">
          <span class="tension-step__marker absolute">
            <span class="tension-step__dot block" />
          </span>
          <h4 class="text-sm font-bold tz-text-primary">4. Stress Relief & Balancing</h4>
          <p class="text-xs tz-text-secondary mt-1">Repeated de-stressing cycles to ensure tension does not drop after first ride.</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLocalePath } from '#imports'
import GuideImage from '~/components/GuideImage.vue'

const localePath = useLocalePath()
</script>

<style scoped>
.tension-section {
  --tension-accent: var(--tz-site-accent);
  --tension-accent-soft: rgba(5, 150, 105, 0.12);
  --tension-border: var(--tz-border-subtle);
  --tension-border-strong: var(--tz-border-strong);
  --tension-soft-surface: var(--tz-surface-subtle);
  color: var(--tz-text-secondary);
}

.tension-panel {
  border: 1px solid var(--tension-border);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.2);
}

.tension-panel--intro {
  border-top: 2px solid var(--tension-accent);
}

.tension-title {
  color: var(--tz-text-primary);
  font-size: var(--tz-type-card-title);
  font-weight: 700;
  line-height: 1.35;
}

.tension-title--section {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.tension-panel--compact .tension-title--section {
  margin-bottom: 1rem;
}

.tension-panel--process .tension-title--section {
  margin-bottom: 1.5rem;
}

.tension-title--section::before {
  width: 0.4rem;
  height: 0.4rem;
  flex: 0 0 auto;
  border-radius: 9999px;
  background: var(--tension-accent);
  box-shadow: 0 0 0 4px var(--tension-accent-soft);
  content: '';
}

.tension-copy {
  color: var(--tz-text-secondary);
  font-size: var(--tz-type-description);
  line-height: 1.65;
}

.tension-copy strong,
.tension-list strong {
  color: var(--tz-text-primary);
  font-weight: 600;
}

.tension-image {
  border: 1px solid var(--tension-border-strong);
  box-shadow: 0 8px 22px rgba(0, 0, 0, 0.26);
}

.tension-link {
  border: 1px solid var(--tension-border) !important;
  background: var(--tension-soft-surface);
}

.tension-link:hover {
  border-color: var(--tension-accent-soft) !important;
  background: var(--tension-accent-soft);
  color: var(--tz-text-primary);
}

.tension-list__item {
  color: var(--tz-text-secondary);
}

.tension-list__marker {
  width: 0.4rem;
  height: 0.4rem;
  border-radius: 9999px;
  background: var(--tension-accent);
  box-shadow: 0 0 0 3px var(--tension-accent-soft);
}

.tension-option {
  border: 1px solid var(--tension-border);
  border-left: 2px solid var(--tension-accent);
  background: var(--tension-soft-surface);
}

.tension-option__label {
  color: var(--tz-text-primary);
  line-height: 1.35;
}

.tension-option__copy {
  color: var(--tz-text-muted);
  line-height: 1.55;
}

.tension-table-panel {
  overflow: hidden;
}

.tension-table-header {
  border-bottom: 1px solid var(--tension-border);
  background: var(--tension-soft-surface);
}

.tension-unit {
  border: 1px solid var(--tension-border);
  border-radius: 0.45rem;
  padding: 0.3rem 0.5rem;
  background: var(--tz-input-surface);
  color: var(--tz-text-muted);
}

.tension-table {
  border-collapse: collapse;
  color: var(--tz-text-secondary);
}

.tension-table th,
.tension-table td {
  padding: 0.8rem 1.25rem;
  border-bottom: 1px solid var(--tension-border);
}

.tension-table th {
  background: var(--tz-surface-subtle);
  color: var(--tz-text-secondary);
  font-weight: 600;
}

.tension-table tbody tr {
  transition: background-color 0.18s ease;
}

.tension-table tbody tr:hover {
  background: var(--tension-soft-surface);
}

.tension-table tbody tr:last-child td {
  border-bottom: 0;
}

.tension-table__row-label {
  color: var(--tz-text-primary);
  font-weight: 600;
}

.tension-note {
  border-top: 1px solid rgba(5, 150, 105, 0.16);
  background: rgba(5, 150, 105, 0.045);
  color: var(--tz-text-secondary);
  line-height: 1.6;
  padding: 1rem 1.25rem;
}

.tension-note strong {
  color: var(--tension-accent);
  font-weight: 600;
}

.tension-timeline {
  border-left: 1px solid rgba(5, 150, 105, 0.22);
}

.tension-step__marker {
  top: 0.05rem;
  left: -1.9rem;
  padding: 0.35rem;
  background: var(--tz-card-surface);
}

.tension-step__dot {
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 9999px;
  background: var(--tension-accent);
  box-shadow: 0 0 0 3px var(--tension-accent-soft);
}

.tension-step--two .tension-step__dot {
  opacity: 0.82;
}

.tension-step--three .tension-step__dot {
  opacity: 0.64;
}

.tension-step--four .tension-step__dot {
  opacity: 0.46;
}

@media (max-width: 640px) {
  .tension-table th,
  .tension-table td {
    padding: 0.72rem 0.9rem;
  }

  .tension-table-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 0.65rem;
  }
}
</style>
