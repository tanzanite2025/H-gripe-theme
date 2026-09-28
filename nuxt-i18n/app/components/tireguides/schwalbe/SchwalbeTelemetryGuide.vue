<template>
  <aside class="schwalbe-telemetry" aria-labelledby="schwalbe-telemetry-title">
    <header class="schwalbe-telemetry__header">
      <div class="schwalbe-telemetry__heading">
        <span class="schwalbe-telemetry__year" aria-hidden="true">25/26</span>
        <div>
          <p class="schwalbe-telemetry__kicker">{{ tx('telemetryGuide.kicker') }}</p>
          <h2 id="schwalbe-telemetry-title">{{ tx('telemetryGuide.title') }}</h2>
          <p class="schwalbe-telemetry__subtitle">{{ tx('telemetryGuide.subtitle') }}</p>
        </div>
      </div>
      <div class="schwalbe-telemetry__meta">
        <span class="schwalbe-telemetry__badge">{{ tx('telemetryGuide.badge') }}</span>
        <span class="schwalbe-telemetry__records">
          {{ tx('telemetryGuide.records', { count: catalogItems.length }) }}
        </span>
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
        v-if="activeTechTab === 'all' || activeTechTab === 'radial'"
        class="schwalbe-telemetry__topic schwalbe-telemetry__topic--radial"
      >
        <div class="schwalbe-telemetry__topic-heading">
          <div>
            <p class="schwalbe-telemetry__topic-label">{{ tx('telemetryGuide.radial.label') }}</p>
            <h3>{{ tx('telemetryGuide.radial.title') }}</h3>
          </div>
          <span class="schwalbe-telemetry__topic-count">
            {{ tx('telemetryGuide.radial.count', { count: radialCount }) }}
          </span>
        </div>

        <div class="schwalbe-telemetry__two-column">
          <article class="schwalbe-telemetry__card schwalbe-telemetry__card--violet">
            <h4>{{ tx('telemetryGuide.radial.principleTitle') }}</h4>
            <p>{{ tx('telemetryGuide.radial.principleBody') }}</p>
            <div class="schwalbe-telemetry__comparison">
              <div>
                <strong>{{ tx('telemetryGuide.radial.observedTitle') }}</strong>
                <span>{{ tx('telemetryGuide.radial.observedBody') }}</span>
              </div>
              <div>
                <strong>{{ tx('telemetryGuide.radial.boundaryTitle') }}</strong>
                <span>{{ tx('telemetryGuide.radial.boundaryBody') }}</span>
              </div>
            </div>
          </article>

          <article class="schwalbe-telemetry__card schwalbe-telemetry__card--neutral">
            <h4>{{ tx('telemetryGuide.radial.snapshotTitle') }}</h4>
            <div class="schwalbe-telemetry__big-number">{{ radialCount }}</div>
            <p>{{ tx('telemetryGuide.radial.snapshotBody') }}</p>
            <ul class="schwalbe-telemetry__tag-list">
              <li v-for="entry in radialLabels" :key="entry.label">
                <span>{{ entry.label }}</span>
                <b>{{ entry.count }}</b>
              </li>
            </ul>
          </article>
        </div>
      </section>

      <section
        v-if="activeTechTab === 'all' || activeTechTab === 'green'"
        class="schwalbe-telemetry__topic schwalbe-telemetry__topic--green"
      >
        <div class="schwalbe-telemetry__topic-heading">
          <div>
            <p class="schwalbe-telemetry__topic-label">{{ tx('telemetryGuide.green.label') }}</p>
            <h3>{{ tx('telemetryGuide.green.title') }}</h3>
          </div>
          <span class="schwalbe-telemetry__topic-count">
            {{ tx('telemetryGuide.green.count', { count: greenMarathonCount }) }}
          </span>
        </div>

        <div class="schwalbe-telemetry__two-column">
          <article class="schwalbe-telemetry__card schwalbe-telemetry__card--green">
            <h4>{{ tx('telemetryGuide.green.modelTitle') }}</h4>
            <p>{{ tx('telemetryGuide.green.modelBody', { count: distinctModelCount }) }}</p>
            <div class="schwalbe-telemetry__stat-grid">
              <div>
                <strong>{{ distinctModelCount }}</strong>
                <span>{{ tx('telemetryGuide.green.modelStat') }}</span>
              </div>
              <div>
                <strong>{{ catalogItems.length }}</strong>
                <span>{{ tx('telemetryGuide.green.recordStat') }}</span>
              </div>
            </div>
          </article>

          <article class="schwalbe-telemetry__card schwalbe-telemetry__card--green">
            <h4>{{ tx('telemetryGuide.green.marathonTitle') }}</h4>
            <p>{{ tx('telemetryGuide.green.marathonBody') }}</p>
            <div class="schwalbe-telemetry__field-list">
              <span><b>Version</b>{{ greenMarathonFields.version || tx('telemetryGuide.notObserved') }}</span>
              <span><b>Compound</b>{{ greenMarathonFields.compound || tx('telemetryGuide.notObserved') }}</span>
              <span><b>Seal</b>{{ greenMarathonFields.seal || tx('telemetryGuide.notObserved') }}</span>
              <span><b>Tread</b>{{ greenMarathonFields.tread || tx('telemetryGuide.notObserved') }}</span>
            </div>
          </article>
        </div>
      </section>

      <div class="schwalbe-telemetry__two-column schwalbe-telemetry__two-column--lower">
        <section
          v-if="activeTechTab === 'all' || activeTechTab === 'protection'"
          class="schwalbe-telemetry__topic schwalbe-telemetry__topic--protection"
        >
          <div class="schwalbe-telemetry__topic-heading">
            <div>
              <p class="schwalbe-telemetry__topic-label">{{ tx('telemetryGuide.protection.label') }}</p>
              <h3>{{ tx('telemetryGuide.protection.title') }}</h3>
            </div>
            <span class="schwalbe-telemetry__topic-count">{{ tx('telemetryGuide.protection.badge') }}</span>
          </div>
          <p class="schwalbe-telemetry__topic-note">{{ tx('telemetryGuide.protection.note') }}</p>
          <div class="schwalbe-telemetry__level-grid">
            <article v-for="entry in protectionLevels" :key="entry.level" class="schwalbe-telemetry__level">
              <span class="schwalbe-telemetry__level-mark">{{ entry.level }}</span>
              <div>
                <div class="schwalbe-telemetry__list-title">
                  <strong>{{ entry.label }}</strong>
                  <span>{{ entry.count }}</span>
                </div>
                <p>{{ tx('telemetryGuide.protection.levelBody') }}</p>
              </div>
            </article>
          </div>
          <h4 class="schwalbe-telemetry__subheading">{{ tx('telemetryGuide.protection.labelsTitle') }}</h4>
          <div class="schwalbe-telemetry__list">
            <article v-for="entry in protectionStats" :key="entry.label" class="schwalbe-telemetry__list-item">
              <span class="schwalbe-telemetry__list-mark" aria-hidden="true"></span>
              <div>
                <div class="schwalbe-telemetry__list-title">
                  <strong>{{ entry.label }}</strong>
                  <span>{{ entry.count }}</span>
                </div>
                <p>{{ tx('telemetryGuide.protection.itemBody') }}</p>
              </div>
            </article>
          </div>
        </section>

        <section
          v-if="activeTechTab === 'all' || activeTechTab === 'addix'"
          class="schwalbe-telemetry__topic schwalbe-telemetry__topic--addix"
        >
          <div class="schwalbe-telemetry__topic-heading">
            <div>
              <p class="schwalbe-telemetry__topic-label">{{ tx('telemetryGuide.addix.label') }}</p>
              <h3>{{ tx('telemetryGuide.addix.title') }}</h3>
            </div>
            <span class="schwalbe-telemetry__topic-count">
              {{ tx('telemetryGuide.addix.count', { count: observedCompoundCount }) }}
            </span>
          </div>
          <p class="schwalbe-telemetry__topic-note">{{ tx('telemetryGuide.addix.note') }}</p>
          <div class="schwalbe-telemetry__compound-grid">
            <article v-for="entry in compoundStats" :key="entry.label" class="schwalbe-telemetry__compound">
              <span class="schwalbe-telemetry__compound-dot" aria-hidden="true"></span>
              <div>
                <div class="schwalbe-telemetry__list-title">
                  <strong>{{ entry.label }}</strong>
                  <span>{{ entry.count }}</span>
                </div>
                <p>{{ tx('telemetryGuide.addix.itemBody') }}</p>
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
import { computed, ref } from 'vue'
import { useI18n } from '#imports'
import type { SchwalbeTireCatalogItem } from '~/data/tireguides/schwalbeCatalog'

type TelemetryTab = 'all' | 'radial' | 'green' | 'protection' | 'addix'

const props = withDefaults(defineProps<{
  catalogItems?: SchwalbeTireCatalogItem[]
}>(), {
  catalogItems: () => [],
})

const { t: translate } = useI18n()
const tx = (key: string, params?: Record<string, unknown>) => translate(`guidesSchwalbeTireSelector.${key}`, params || {})

const activeTechTab = ref<TelemetryTab>('all')
const showTechGuide = ref(true)
const tabs: Array<{ id: TelemetryTab; labelKey: string }> = [
  { id: 'all', labelKey: 'telemetryGuide.tabs.all' },
  { id: 'radial', labelKey: 'telemetryGuide.tabs.radial' },
  { id: 'green', labelKey: 'telemetryGuide.tabs.green' },
  { id: 'protection', labelKey: 'telemetryGuide.tabs.protection' },
  { id: 'addix', labelKey: 'telemetryGuide.tabs.addix' },
]

const textOf = (item: SchwalbeTireCatalogItem) => [
  item.model_name,
  item.version_label,
  item.compound,
  item.seal,
  item.tread,
].filter(Boolean).join(' ').toLowerCase()

const countText = (term: string) => props.catalogItems.filter((item) => textOf(item).includes(term.toLowerCase())).length
const countVersion = (term: string) => props.catalogItems.filter((item) => item.version_label?.toLowerCase().includes(term.toLowerCase())).length
const countVersions = (terms: string[]) => props.catalogItems.filter((item) => {
  const version = item.version_label?.toLowerCase() || ''
  return terms.some((term) => version.includes(term.toLowerCase()))
}).length

const radialCount = computed(() => props.catalogItems.filter((item) => textOf(item).includes('radial')).length)
const distinctModelCount = computed(() => new Set(props.catalogItems.map((item) => item.model_name.trim()).filter(Boolean)).size)
const greenMarathonItems = computed(() => props.catalogItems.filter((item) => item.model_name.trim().toLowerCase() === 'green marathon'))
const greenMarathonCount = computed(() => greenMarathonItems.value.length)

const greenMarathonFields = computed(() => {
  const item = greenMarathonItems.value[0]
  return {
    version: item?.version_label,
    compound: item?.compound,
    seal: item?.seal,
    tread: item?.tread,
  }
})

const radialLabels = computed(() => [
  'GRAVITY PRO, Radial',
  'TRAIL PRO, Radial',
  'DD, RaceGuard, Radial',
  'DD, GreenGuard, Radial',
].map((label) => ({ label, count: countText(label.toLowerCase()) })))

const protectionLevels = computed(() => [
  { level: 'L7', label: 'SmartGuard / Smart DualGuard', terms: ['SmartGuard', 'Smart DualGuard'] },
  { level: 'L6', label: 'Super Defense / Double Defense', terms: ['Super Defense', 'Double Defense'] },
  { level: 'L5', label: 'V-Guard / GreenGuard / RaceGuard', terms: ['V-Guard', 'GreenGuard', 'RaceGuard'] },
  { level: 'L4', label: 'RaceGuard', terms: ['RaceGuard'] },
  { level: 'L3', label: 'K-Guard', terms: ['K-Guard'] },
  { level: 'L2', label: '67 EPI carcass', terms: [] },
  { level: 'L1', label: '50 EPI carcass', terms: [] },
].map((entry) => ({
  ...entry,
  count: entry.terms.length
    ? countVersions(entry.terms)
    : props.catalogItems.filter((item) => item.epi === (entry.level === 'L2' ? 67 : 50)).length,
})))

const protectionStats = computed(() => [
  'SmartGuard',
  'V-Guard',
  'GreenGuard',
  'RaceGuard',
  'K-Guard',
  'PunctureGuard',
  'DD',
].map((label) => ({ label, count: countVersion(label) })).filter((entry) => entry.count > 0))

const compoundLabels = [
  'ADDIX',
  'ADDIX Green',
  'ADDIX Race',
  'ADDIX Eco',
  'ADDIX 365',
  'ADDIX Speed',
  'ADDIX SpeedGrip',
  'ADDIX E',
  'ADDIX 4-Season',
  'ADDIX Soft',
  'ADDIX Ultra Soft',
]

const compoundStats = computed(() => compoundLabels.map((label) => ({
  label,
  count: props.catalogItems.filter((item) => item.compound?.trim().toLowerCase() === label.toLowerCase()).length,
})).filter((entry) => entry.count > 0))

const observedCompoundCount = computed(() => new Set(props.catalogItems.map((item) => item.compound?.trim()).filter(Boolean)).size)
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
.schwalbe-telemetry__card p,
.schwalbe-telemetry__topic-note,
.schwalbe-telemetry__list-item p,
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

.schwalbe-telemetry__badge,
.schwalbe-telemetry__records,
.schwalbe-telemetry__topic-count {
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

.schwalbe-telemetry__records {
  background: var(--tz-surface-subtle);
  color: var(--tz-text-secondary);
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

.schwalbe-telemetry__topic-count {
  background: var(--tz-surface-subtle);
  color: var(--tz-text-secondary);
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

.schwalbe-telemetry__card {
  display: grid;
  gap: 0.6rem;
  min-width: 0;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.9rem;
  padding: 0.9rem;
}

.schwalbe-telemetry__card--violet {
  border-color: #ddd6fe;
  background: #faf5ff;
}

.schwalbe-telemetry__card--green {
  border-color: #bbf7d0;
  background: #f0fdf4;
}

.schwalbe-telemetry__card--neutral {
  background: var(--tz-surface-subtle);
}

.schwalbe-telemetry__comparison {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.5rem;
}

.schwalbe-telemetry__comparison > div,
.schwalbe-telemetry__stat-grid > div {
  display: grid;
  gap: 0.2rem;
  border-radius: 0.65rem;
  background: color-mix(in srgb, var(--tz-card-surface) 80%, transparent);
  padding: 0.55rem;
}

.schwalbe-telemetry__comparison strong,
.schwalbe-telemetry__stat-grid span,
.schwalbe-telemetry__field-list,
.schwalbe-telemetry__tag-list,
.schwalbe-telemetry__list-item p,
.schwalbe-telemetry__compound p {
  font-size: 0.68rem;
  line-height: 1.5;
}

.schwalbe-telemetry__comparison span {
  color: var(--tz-text-secondary);
  font-size: 0.68rem;
  line-height: 1.5;
}

.schwalbe-telemetry__big-number {
  color: #7c3aed;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 2rem;
  font-weight: 850;
  line-height: 1;
}

.schwalbe-telemetry__tag-list,
.schwalbe-telemetry__field-list {
  display: grid;
  gap: 0.35rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.schwalbe-telemetry__tag-list li,
.schwalbe-telemetry__field-list span {
  display: flex;
  gap: 0.5rem;
  align-items: baseline;
  justify-content: space-between;
  border-bottom: 1px solid color-mix(in srgb, var(--tz-border-subtle) 70%, transparent);
  color: var(--tz-text-secondary);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

.schwalbe-telemetry__tag-list b {
  color: var(--tz-text-primary);
}

.schwalbe-telemetry__field-list b {
  color: var(--tz-text-primary);
  font-weight: 700;
}

.schwalbe-telemetry__stat-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.5rem;
}

.schwalbe-telemetry__stat-grid strong {
  color: #047857;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 1.25rem;
}

.schwalbe-telemetry__topic-note {
  border-left: 3px solid var(--tz-border-strong);
  padding-left: 0.65rem;
}

.schwalbe-telemetry__list,
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

.schwalbe-telemetry__subheading {
  color: var(--tz-text-secondary) !important;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.72rem !important;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.schwalbe-telemetry__list-item,
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

.schwalbe-telemetry__list-mark,
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

.schwalbe-telemetry__list-title span {
  flex: 0 0 auto;
  color: var(--tz-text-secondary);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.65rem;
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

  .schwalbe-telemetry__level-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 520px) {
  .schwalbe-telemetry__heading {
    display: grid;
  }

  .schwalbe-telemetry__comparison,
  .schwalbe-telemetry__stat-grid {
    grid-template-columns: 1fr;
  }

  .schwalbe-telemetry__tabs {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .schwalbe-telemetry__tab {
    width: 100%;
  }
}
</style>
