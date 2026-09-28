<template>
  <aside class="schwalbe-telemetry" :aria-labelledby="headingId">
    <header class="schwalbe-telemetry__header">
      <div class="schwalbe-telemetry__heading">
        <span class="schwalbe-telemetry__year" aria-hidden="true">25/26</span>
        <div>
          <p class="schwalbe-telemetry__kicker">{{ tx('telemetryGuide.kicker') }}</p>
          <h2 :id="headingId">{{ tx('telemetryGuide.title') }}</h2>
          <p class="schwalbe-telemetry__subtitle">{{ tx('telemetryGuide.subtitle') }}</p>
        </div>
      </div>
      <div class="schwalbe-telemetry__meta">
        <span class="schwalbe-telemetry__badge">{{ tx('telemetryGuide.badge') }}</span>
      </div>
    </header>

    <div class="schwalbe-telemetry__toolbar">
      <div class="schwalbe-telemetry__tabs" role="tablist" :aria-label="tx('telemetryGuide.tabListLabel')">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          role="tab"
          :aria-selected="activeTechTab === tab.id"
          :class="['schwalbe-telemetry__tab', { 'schwalbe-telemetry__tab--active': activeTechTab === tab.id }]"
          @click="activeTechTab = tab.id"
        >
          {{ tx(tab.labelKey) }}
        </button>
      </div>
      <button
        type="button"
        class="schwalbe-telemetry__toggle"
        :aria-expanded="showTechGuide"
        @click="showTechGuide = !showTechGuide"
      >
        {{ showTechGuide ? tx('telemetryGuide.collapse') : tx('telemetryGuide.expand') }}
      </button>
    </div>

    <div v-show="showTechGuide" class="schwalbe-telemetry__content">
      <section
        v-if="activeTechTab === 'radial'"
        class="schwalbe-telemetry__topic schwalbe-telemetry__topic--radial"
      >
        <div class="schwalbe-telemetry__topic-heading">
          <div>
            <p class="schwalbe-telemetry__topic-label">{{ tx('telemetryGuide.radial.label') }}</p>
            <h3>{{ tx('telemetryGuide.radial.title') }}</h3>
          </div>
        </div>

        <p class="schwalbe-telemetry__topic-note">{{ tx('telemetryGuide.radial.note') }}</p>
        <div class="schwalbe-telemetry__radial-grid">
          <article v-for="entry in radialGuides" :key="entry.key" class="schwalbe-telemetry__radial-card">
            <h4>{{ tx(`telemetryGuide.radial.items.${entry.key}.title`) }}</h4>
            <p>{{ tx(`telemetryGuide.radial.items.${entry.key}.body`) }}</p>
          </article>
        </div>
      </section>

      <section
        v-if="activeTechTab === 'green'"
        class="schwalbe-telemetry__topic schwalbe-telemetry__topic--green"
      >
        <div class="schwalbe-telemetry__topic-heading">
          <div>
            <p class="schwalbe-telemetry__topic-label">{{ tx('telemetryGuide.green.label') }}</p>
            <h3>{{ tx('telemetryGuide.green.title') }}</h3>
          </div>
        </div>

        <p class="schwalbe-telemetry__topic-note">{{ tx('telemetryGuide.green.note') }}</p>
        <div class="schwalbe-telemetry__green-grid">
          <article v-for="entry in greenGuides" :key="entry.key" class="schwalbe-telemetry__green-card">
            <h4>{{ tx(`telemetryGuide.green.items.${entry.key}.title`) }}</h4>
            <p>{{ tx(`telemetryGuide.green.items.${entry.key}.body`) }}</p>
          </article>
        </div>
      </section>

      <div class="schwalbe-telemetry__two-column schwalbe-telemetry__two-column--lower">
        <section
          v-if="activeTechTab === 'protection'"
          class="schwalbe-telemetry__topic schwalbe-telemetry__topic--protection"
        >
          <div class="schwalbe-telemetry__topic-heading">
            <div>
              <p class="schwalbe-telemetry__topic-label">{{ tx('telemetryGuide.protection.label') }}</p>
              <h3>{{ tx('telemetryGuide.protection.title') }}</h3>
            </div>
          </div>
          <p class="schwalbe-telemetry__topic-note">{{ tx('telemetryGuide.protection.note') }}</p>
          <div class="schwalbe-telemetry__level-grid">
            <article v-for="entry in protectionLevels" :key="entry.key" class="schwalbe-telemetry__level">
              <span class="schwalbe-telemetry__level-mark">{{ entry.mark }}</span>
              <div>
                <div class="schwalbe-telemetry__list-title">
                  <strong>{{ tx(`telemetryGuide.protection.items.${entry.key}.title`) }}</strong>
                </div>
                <p>{{ tx(`telemetryGuide.protection.items.${entry.key}.body`) }}</p>
              </div>
            </article>
          </div>
          <article class="schwalbe-telemetry__protection-extra">
            <div class="schwalbe-telemetry__list-title">
              <strong>{{ tx('telemetryGuide.protection.extra.title') }}</strong>
            </div>
            <p>{{ tx('telemetryGuide.protection.extra.body') }}</p>
          </article>
        </section>

        <section
          v-if="activeTechTab === 'addix'"
          class="schwalbe-telemetry__topic schwalbe-telemetry__topic--addix"
        >
          <div class="schwalbe-telemetry__topic-heading">
            <div>
              <p class="schwalbe-telemetry__topic-label">{{ tx('telemetryGuide.addix.label') }}</p>
              <h3>{{ tx('telemetryGuide.addix.title') }}</h3>
            </div>
          </div>
          <p class="schwalbe-telemetry__topic-note">{{ tx('telemetryGuide.addix.note') }}</p>
          <div class="schwalbe-telemetry__compound-grid">
            <article v-for="entry in compoundGuides" :key="entry.key" class="schwalbe-telemetry__compound">
              <span class="schwalbe-telemetry__compound-dot" aria-hidden="true"></span>
              <div>
                <div class="schwalbe-telemetry__list-title">
                  <strong>{{ tx(`telemetryGuide.addix.items.${entry.key}.title`) }}</strong>
                </div>
                <p>{{ tx(`telemetryGuide.addix.items.${entry.key}.body`) }}</p>
              </div>
            </article>
          </div>
        </section>
      </div>
    </div>

    <footer class="schwalbe-telemetry__footer">
      {{ tx('telemetryGuide.disclaimer') }}
    </footer>
  </aside>
</template>

<script setup lang="ts">
import { ref, useId, watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

type TelemetryTab = 'radial' | 'green' | 'protection' | 'addix'

const { locale, t: translate } = useI18n()
const { loadPageMessages } = usePageMessages('guidesSchwalbeTireSelector')
await loadPageMessages(locale.value)
watch(locale, (nextLocale) => void loadPageMessages(nextLocale))
const tx = (key: string, params?: Record<string, unknown>) => translate(`guidesSchwalbeTireSelector.${key}`, params || {})
const headingId = `schwalbe-telemetry-title-${useId()}`

const activeTechTab = ref<TelemetryTab>('radial')
const showTechGuide = ref(true)
const tabs: Array<{ id: TelemetryTab; labelKey: string }> = [
  { id: 'radial', labelKey: 'telemetryGuide.tabs.radial' },
  { id: 'green', labelKey: 'telemetryGuide.tabs.green' },
  { id: 'protection', labelKey: 'telemetryGuide.tabs.protection' },
  { id: 'addix', labelKey: 'telemetryGuide.tabs.addix' },
]

const radialGuides = [
  { key: 'casing' },
  { key: 'contact' },
  { key: 'tradeoff' },
] as const

const greenGuides = [
  { key: 'compound' },
  { key: 'naturalRubber' },
  { key: 'greenGuard' },
  { key: 'circularMaterials' },
] as const

const protectionLevels = [
  { key: 'level7', mark: 'L7' },
  { key: 'level6Plus', mark: 'L6+' },
  { key: 'level6', mark: 'L6' },
  { key: 'level5', mark: 'L5' },
  { key: 'level4', mark: 'L4' },
  { key: 'level3', mark: 'L3' },
  { key: 'level2', mark: 'L2' },
  { key: 'level1', mark: 'L1' },
] as const

const compoundGuides = [
  { key: 'speed' },
  { key: 'mid' },
  { key: 'soft' },
  { key: 'ultraSoft' },
] as const
</script>

<style scoped>
.schwalbe-telemetry {
  display: grid;
  gap: 1rem;
  width: 100%;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1.25rem;
  background: var(--tz-card-surface);
  padding: clamp(1rem, 2vw, 1.5rem);
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.06);
}

.schwalbe-telemetry__header,
.schwalbe-telemetry__toolbar,
.schwalbe-telemetry__topic-heading,
.schwalbe-telemetry__list-title {
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
  justify-content: space-between;
}

.schwalbe-telemetry__heading {
  display: flex;
  gap: 0.75rem;
  min-width: 0;
  align-items: flex-start;
}

.schwalbe-telemetry__year {
  display: grid;
  flex: 0 0 auto;
  width: 2.5rem;
  height: 2.5rem;
  place-items: center;
  border: 1px solid #bae6fd;
  border-radius: 0.75rem;
  background: #f0f9ff;
  color: #0369a1;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.65rem;
  font-weight: 800;
}

.schwalbe-telemetry__kicker,
.schwalbe-telemetry__topic-label {
  margin: 0 0 0.25rem;
  color: var(--tz-text-accent);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  line-height: 1.35;
  text-transform: uppercase;
}

.schwalbe-telemetry h2,
.schwalbe-telemetry h3,
.schwalbe-telemetry h4 {
  margin: 0;
  color: var(--tz-text-primary);
}

.schwalbe-telemetry h2 {
  font-size: clamp(1rem, 1.4vw, 1.25rem);
  line-height: 1.3;
}

.schwalbe-telemetry h3 {
  font-size: 0.9rem;
  line-height: 1.4;
}

.schwalbe-telemetry h4 {
  font-size: 0.84rem;
  line-height: 1.4;
}

.schwalbe-telemetry__subtitle,
.schwalbe-telemetry__topic-note,
.schwalbe-telemetry__green-card p,
.schwalbe-telemetry__compound p,
.schwalbe-telemetry__footer {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.76rem;
  line-height: 1.65;
}

.schwalbe-telemetry__subtitle {
  margin-top: 0.35rem;
  max-width: 62rem;
}

.schwalbe-telemetry__meta {
  display: flex;
  flex: 0 0 auto;
  flex-wrap: wrap;
  gap: 0.4rem;
  align-items: center;
  justify-content: flex-end;
}

.schwalbe-telemetry__badge {
  display: inline-flex;
  border-radius: 999px;
  padding: 0.25rem 0.55rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.65rem;
  font-weight: 700;
  line-height: 1.3;
  white-space: nowrap;
}

.schwalbe-telemetry__badge {
  border: 1px solid #bae6fd;
  background: #f0f9ff;
  color: #0369a1;
}

.schwalbe-telemetry__toolbar {
  flex-wrap: wrap;
  align-items: center;
  border-top: 1px solid var(--tz-border-subtle);
  border-bottom: 1px solid var(--tz-border-subtle);
  padding: 0.75rem 0;
}

.schwalbe-telemetry__tabs {
  display: flex;
  flex: 1 1 36rem;
  flex-wrap: wrap;
  gap: 0.25rem;
  border-radius: 0.75rem;
  background: var(--tz-surface-subtle);
  padding: 0.25rem;
}

.schwalbe-telemetry__tab,
.schwalbe-telemetry__toggle {
  min-height: 2rem;
  border: 1px solid transparent;
  border-radius: 0.55rem;
  background: transparent;
  color: var(--tz-text-secondary);
  padding: 0.35rem 0.55rem;
  font: inherit;
  font-size: 0.72rem;
  font-weight: 650;
  cursor: pointer;
}

.schwalbe-telemetry__tab:hover,
.schwalbe-telemetry__toggle:hover {
  color: var(--tz-text-primary);
}

.schwalbe-telemetry__tab--active {
  border-color: var(--tz-border-subtle);
  background: var(--tz-card-surface);
  color: var(--tz-text-primary);
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.08);
}

.schwalbe-telemetry__toggle {
  border-color: var(--tz-border-strong);
  white-space: nowrap;
}

.schwalbe-telemetry__content {
  display: grid;
  gap: 1.25rem;
}

.schwalbe-telemetry__topic {
  display: grid;
  gap: 0.75rem;
  min-width: 0;
}

.schwalbe-telemetry__topic-heading {
  align-items: center;
}

.schwalbe-telemetry__topic--radial .schwalbe-telemetry__topic-label {
  color: #7c3aed;
}

.schwalbe-telemetry__topic--green .schwalbe-telemetry__topic-label {
  color: #047857;
}

.schwalbe-telemetry__topic--protection .schwalbe-telemetry__topic-label {
  color: #475569;
}

.schwalbe-telemetry__topic--addix .schwalbe-telemetry__topic-label {
  color: #0369a1;
}

.schwalbe-telemetry__two-column {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.schwalbe-telemetry__two-column--lower {
  align-items: start;
}

.schwalbe-telemetry__compound p {
  font-size: 0.68rem;
  line-height: 1.5;
}

.schwalbe-telemetry__topic-note {
  border-left: 3px solid var(--tz-border-strong);
  padding-left: 0.65rem;
}

.schwalbe-telemetry__green-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.5rem;
}

.schwalbe-telemetry__radial-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.5rem;
}

.schwalbe-telemetry__radial-card {
  display: grid;
  gap: 0.4rem;
  min-width: 0;
  border: 1px solid #ddd6fe;
  border-radius: 0.7rem;
  background: #faf5ff;
  padding: 0.7rem;
}

.schwalbe-telemetry__radial-card p {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.76rem;
  line-height: 1.65;
}

.schwalbe-telemetry__green-card {
  display: grid;
  gap: 0.4rem;
  min-width: 0;
  border: 1px solid #bbf7d0;
  border-radius: 0.7rem;
  background: #f0fdf4;
  padding: 0.7rem;
}

.schwalbe-telemetry__compound-grid {
  display: grid;
  gap: 0.5rem;
}

.schwalbe-telemetry__level-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.5rem;
}

.schwalbe-telemetry__level {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 0.6rem;
  align-items: start;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.7rem;
  background: var(--tz-surface-subtle);
  padding: 0.65rem;
}

.schwalbe-telemetry__level-mark {
  display: grid;
  width: 2rem;
  height: 2rem;
  place-items: center;
  border-radius: 0.55rem;
  background: var(--tz-text-primary);
  color: var(--tz-card-surface);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.66rem;
  font-weight: 800;
}

.schwalbe-telemetry__level p {
  margin: 0.25rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.66rem;
  line-height: 1.5;
}

.schwalbe-telemetry__protection-extra {
  display: grid;
  gap: 0.25rem;
  border: 1px dashed var(--tz-border-strong);
  border-radius: 0.7rem;
  background: var(--tz-surface-subtle);
  padding: 0.65rem;
}

.schwalbe-telemetry__protection-extra p {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.66rem;
  line-height: 1.5;
}

.schwalbe-telemetry__compound {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 0.6rem;
  align-items: start;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.7rem;
  background: var(--tz-surface-subtle);
  padding: 0.65rem;
}

.schwalbe-telemetry__compound-dot {
  display: block;
  width: 0.55rem;
  height: 0.55rem;
  margin-top: 0.25rem;
  border-radius: 999px;
  background: var(--tz-text-secondary);
}

.schwalbe-telemetry__list-title {
  align-items: baseline;
}

.schwalbe-telemetry__list-title strong {
  color: var(--tz-text-primary);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.72rem;
}

.schwalbe-telemetry__footer {
  border-top: 1px solid var(--tz-border-subtle);
  padding-top: 0.85rem;
  color: var(--tz-text-muted);
  font-size: 0.7rem;
}

@media (max-width: 760px) {
  .schwalbe-telemetry__header,
  .schwalbe-telemetry__topic-heading {
    display: grid;
  }

  .schwalbe-telemetry__meta {
    justify-content: flex-start;
  }

  .schwalbe-telemetry__two-column {
    grid-template-columns: 1fr;
  }

  .schwalbe-telemetry__green-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .schwalbe-telemetry__radial-grid {
    grid-template-columns: 1fr;
  }

  .schwalbe-telemetry__level-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 520px) {
  .schwalbe-telemetry__heading {
    display: grid;
  }

  .schwalbe-telemetry__tabs {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .schwalbe-telemetry__tab {
    width: 100%;
  }

  .schwalbe-telemetry__green-grid {
    grid-template-columns: 1fr;
  }
}
</style>
