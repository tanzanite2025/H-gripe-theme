<template>
  <div class="freehub-groupset-helper mt-5 rounded-2xl tz-surface-panel p-4 shadow-md">
    <h3 class="mb-2 text-sm font-semibold tz-text-secondary">
      {{ title || t('wheelsetFreehubHelper.title') }}
    </h3>
    <p class="mb-3 text-xs tz-text-secondary">
      {{ description || t('wheelsetFreehubHelper.description') }}
    </p>

    <div class="freehub-groupset-helper__fields">
      <div class="freehub-groupset-helper__field">
        <label class="block text-xs font-medium tz-text-secondary" :for="brandSelectId">
          {{ t('wheelsetFreehubHelper.drivetrainBrand') }}
        </label>
        <select
          :id="brandSelectId"
          v-model="selectedBrand"
          class="freehub-groupset-helper__select mt-1 w-full rounded-md border tz-border-strong tz-surface-panel px-2 py-1.5 text-xs tz-text-secondary shadow-md outline-none focus:border-emerald-600 focus:ring-0"
        >
          <option disabled value="">
            {{ t('wheelsetFreehubHelper.selectBrand') }}
          </option>
          <option
            v-for="brand in brands"
            :key="brand"
            :value="brand"
          >
            {{ brand }}
          </option>
        </select>
      </div>

      <div class="freehub-groupset-helper__field">
        <label class="block text-xs font-medium tz-text-secondary" :for="groupsetSelectId">
          {{ t('wheelsetFreehubHelper.groupset') }}
        </label>
        <select
          :id="groupsetSelectId"
          v-model="selectedGroupsetId"
          :disabled="!selectedBrand"
          class="freehub-groupset-helper__select mt-1 w-full rounded-md border tz-border-strong tz-surface-panel px-2 py-1.5 text-xs tz-text-secondary shadow-md outline-none focus:border-emerald-600 focus:ring-0 disabled:cursor-not-allowed disabled:border-[var(--tz-border-subtle)] disabled:tz-text-muted"
        >
          <option disabled value="">
            {{ selectedBrand
              ? t('wheelsetFreehubHelper.selectGroupset')
              : t('wheelsetFreehubHelper.chooseBrandFirst') }}
          </option>
          <option
            v-for="option in filteredGroupsets"
            :key="option.id"
            :value="option.id"
          >
            {{ optionLabel(option) }}
          </option>
        </select>
      </div>
    </div>

    <div class="freehub-groupset-helper__status">
      <p
        v-if="!activeOption"
        class="text-xs tz-text-muted"
      >
        {{ t('wheelsetFreehubHelper.empty') }}
      </p>

      <div
        v-else
        class="text-xs tz-text-secondary"
      >
        <p class="font-semibold text-emerald-700">
          {{ t('wheelsetFreehubHelper.recommended') }}
          <span class="ml-1">{{ optionFreehub(activeOption) }}</span>
        </p>
        <p
          v-if="activeOption.notesKey && !activeOption.factId"
          class="mt-0.5 tz-caption tz-text-muted"
        >
          {{ optionNotes(activeOption) }}
        </p>
      </div>
    </div>

    <section
      v-if="activeOption"
      class="freehub-groupset-helper__facts"
      :aria-label="t('wheelsetFreehubHelper.facts.heading')"
    >
      <div
        ref="factRail"
        class="freehub-groupset-helper__fact-rail"
        role="region"
        :aria-label="t('wheelsetFreehubHelper.facts.heading')"
        @scroll.passive="updateActiveFact"
      >
        <article
          v-for="(fact, index) in compatibilityFacts"
          :key="fact.id"
          :id="`${helperId}-fact-panel-${index}`"
          :data-freehub-fact-index="index"
          :aria-labelledby="`${helperId}-fact-tab-${index}`"
          class="freehub-groupset-helper__fact"
          :class="[
            `freehub-groupset-helper__fact--${fact.id}`,
            { 'is-relevant': isFactRelevant(fact.id) },
          ]"
          role="tabpanel"
        >
          <div class="freehub-groupset-helper__fact-heading">
            <span class="freehub-groupset-helper__fact-label">{{ t(fact.labelKey) }}</span>
            <h4>{{ t(fact.titleKey) }}</h4>
          </div>
          <p class="freehub-groupset-helper__fact-callout">
            {{ t(fact.calloutKey) }}
          </p>
          <p class="freehub-groupset-helper__fact-copy">
            {{ t(fact.bodyKey) }}
          </p>
        </article>
      </div>

      <div
        class="tz-carousel-pagination freehub-groupset-helper__pagination"
        role="tablist"
        :aria-label="t('wheelsetFreehubHelper.facts.pagination')"
      >
        <button
          v-for="(fact, index) in compatibilityFacts"
          :key="`${fact.id}-dot`"
          :id="`${helperId}-fact-tab-${index}`"
          type="button"
          class="tz-carousel-pagination__dot"
          :class="{ 'is-active': activeFactIndex === index }"
          :aria-label="t('wheelsetFreehubHelper.facts.showFact', { fact: index + 1 })"
          :aria-controls="`${helperId}-fact-panel-${index}`"
          :aria-selected="activeFactIndex === index"
          role="tab"
          @click="scrollToFact(index)"
        />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, useId, watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

withDefaults(defineProps<{
  title?: string
  description?: string
}>(), {
  title: '',
  description: '',
})

interface FreehubOption {
  id: string
  brand: string
  labelKey: string
  freehubKey: string
  notesKey?: string
  factId?: CompatibilityFactId
}

type CompatibilityFactId = 'xdr-xd' | 'shimano-road-12-speed'
// NOTE / 说明：
// 如果后续要增加更多套件 → 塔基类型的对应关系，直接在下面的 FREEHUB_OPTIONS 数组中追加一条对象即可。
// 不需要改其他文件；品牌下拉会根据 brand 字段自动生成，套件下拉会根据 brand 自动过滤。
const FREEHUB_OPTIONS: FreehubOption[] = [
  {
    id: 'shimano-deore-m6100',
    brand: 'Shimano',
    labelKey: 'options.shimanoDeoreM6100.label',
    freehubKey: 'options.shimanoDeoreM6100.freehub',
    notesKey: 'options.shimanoDeoreM6100.notes',
  },
  {
    id: 'shimano-slx-m7100',
    brand: 'Shimano',
    labelKey: 'options.shimanoSlxM7100.label',
    freehubKey: 'options.shimanoSlxM7100.freehub',
  },
  {
    id: 'shimano-xt-m8100',
    brand: 'Shimano',
    labelKey: 'options.shimanoXtM8100.label',
    freehubKey: 'options.shimanoXtM8100.freehub',
  },
  {
    id: 'shimano-xtr-m9100',
    brand: 'Shimano',
    labelKey: 'options.shimanoXtrM9100.label',
    freehubKey: 'options.shimanoXtrM9100.freehub',
  },
  {
    id: 'shimano-105-r7000',
    brand: 'Shimano',
    labelKey: 'options.shimano105R7000.label',
    freehubKey: 'options.shimano105R7000.freehub',
  },
  {
    id: 'shimano-ultegra-r8000',
    brand: 'Shimano',
    labelKey: 'options.shimanoUltegraR8000.label',
    freehubKey: 'options.shimanoUltegraR8000.freehub',
  },
  {
    id: 'shimano-duraace-r9100',
    brand: 'Shimano',
    labelKey: 'options.shimanoDuraAceR9100.label',
    freehubKey: 'options.shimanoDuraAceR9100.freehub',
  },
  {
    id: 'shimano-105-di2-r7100',
    brand: 'Shimano',
    labelKey: 'options.shimano105Di2R7100.label',
    freehubKey: 'options.shimano105Di2R7100.freehub',
    notesKey: 'options.shimano105Di2R7100.notes',
    factId: 'shimano-road-12-speed',
  },
  {
    id: 'shimano-ultegra-di2-r8100',
    brand: 'Shimano',
    labelKey: 'options.shimanoUltegraDi2R8100.label',
    freehubKey: 'options.shimanoUltegraDi2R8100.freehub',
    notesKey: 'options.shimanoUltegraDi2R8100.notes',
    factId: 'shimano-road-12-speed',
  },
  {
    id: 'sram-gx-eagle',
    brand: 'SRAM',
    labelKey: 'options.sramGxEagle.label',
    freehubKey: 'options.sramGxEagle.freehub',
    notesKey: 'options.sramGxEagle.notes',
    factId: 'xdr-xd',
  },
  {
    id: 'sram-x01-eagle',
    brand: 'SRAM',
    labelKey: 'options.sramX01Eagle.label',
    freehubKey: 'options.sramX01Eagle.freehub',
    factId: 'xdr-xd',
  },
  {
    id: 'sram-nx-eagle',
    brand: 'SRAM',
    labelKey: 'options.sramNxEagle.label',
    freehubKey: 'options.sramNxEagle.freehub',
  },
  {
    id: 'sram-force-etap-axs',
    brand: 'SRAM',
    labelKey: 'options.sramForceEtapAxs.label',
    freehubKey: 'options.sramForceEtapAxs.freehub',
    factId: 'xdr-xd',
  },
  {
    id: 'sram-red-etap-axs',
    brand: 'SRAM',
    labelKey: 'options.sramRedEtapAxs.label',
    freehubKey: 'options.sramRedEtapAxs.freehub',
    factId: 'xdr-xd',
  },
  {
    id: 'campagnolo-ekar-13',
    brand: 'Campagnolo',
    labelKey: 'options.campagnoloEkar13.label',
    freehubKey: 'options.campagnoloEkar13.freehub',
  },
  {
    id: 'campagnolo-super-record-11',
    brand: 'Campagnolo',
    labelKey: 'options.campagnoloSuperRecord11.label',
    freehubKey: 'options.campagnoloSuperRecord11.freehub',
  },
]

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('wheelsetFreehubHelper')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const selectedBrand = ref<string>('')
const selectedGroupsetId = ref<string>('')
const helperId = useId()
const factRail = ref<HTMLElement | null>(null)
const activeFactIndex = ref(0)
const brandSelectId = computed(() => `${helperId}-freehub-brand`)
const groupsetSelectId = computed(() => `${helperId}-freehub-groupset`)

const compatibilityFacts = [
  {
    id: 'xdr-xd',
    labelKey: 'wheelsetFreehubHelper.facts.xdrXd.label',
    titleKey: 'wheelsetFreehubHelper.facts.xdrXd.title',
    calloutKey: 'wheelsetFreehubHelper.facts.xdrXd.callout',
    bodyKey: 'wheelsetFreehubHelper.facts.xdrXd.body',
  },
  {
    id: 'shimano-road-12-speed',
    labelKey: 'wheelsetFreehubHelper.facts.shimano12.label',
    titleKey: 'wheelsetFreehubHelper.facts.shimano12.title',
    calloutKey: 'wheelsetFreehubHelper.facts.shimano12.callout',
    bodyKey: 'wheelsetFreehubHelper.facts.shimano12.body',
  },
] as const

const brands = computed(() => {
  const unique = new Set<string>()
  for (const option of FREEHUB_OPTIONS) {
    unique.add(option.brand)
  }
  return Array.from(unique)
})

const filteredGroupsets = computed(() => {
  if (!selectedBrand.value) return []
  return FREEHUB_OPTIONS.filter((option) => option.brand === selectedBrand.value)
})

const activeOption = computed(() => {
  if (!selectedGroupsetId.value) return null
  return FREEHUB_OPTIONS.find((option) => option.id === selectedGroupsetId.value) ?? null
})

watch(selectedBrand, () => {
  selectedGroupsetId.value = ''
})

const isFactRelevant = (factId: CompatibilityFactId) => (
  activeOption.value?.factId === factId
)

watch(activeOption, async () => {
  const relevantIndex = compatibilityFacts.findIndex((fact) => isFactRelevant(fact.id))
  const index = relevantIndex >= 0 ? relevantIndex : 0
  activeFactIndex.value = index
  await nextTick()
  scrollToFact(index)
})

const optionLabel = (option: FreehubOption) => (
  t(`wheelsetFreehubHelper.${option.labelKey}`)
)
const optionFreehub = (option: FreehubOption) => (
  t(`wheelsetFreehubHelper.${option.freehubKey}`)
)
const optionNotes = (option: FreehubOption) => (
  option.notesKey ? t(`wheelsetFreehubHelper.${option.notesKey}`) : ''
)

const scrollToFact = (index: number) => {
  const currentRail = factRail.value
  const card = currentRail?.querySelector<HTMLElement>(
    `[data-freehub-fact-index="${index}"]`,
  )
  if (!currentRail || !card) return

  currentRail.scrollTo({ left: card.offsetLeft, behavior: 'smooth' })
  activeFactIndex.value = index
}

const updateActiveFact = () => {
  const currentRail = factRail.value
  if (!currentRail) return

  const center = currentRail.scrollLeft + currentRail.clientWidth / 2
  const cards = Array.from(
    currentRail.querySelectorAll<HTMLElement>('[data-freehub-fact-index]'),
  )
  if (!cards.length) return
  const nearestCard = cards.reduce((nearest, card) => (
    Math.abs(card.offsetLeft + card.offsetWidth / 2 - center)
      < Math.abs(nearest.offsetLeft + nearest.offsetWidth / 2 - center)
      ? card
      : nearest
  ))

  activeFactIndex.value = Number(nearestCard.dataset.freehubFactIndex || 0)
}
</script>

<style scoped>
.freehub-groupset-helper__fields {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.75rem;
  align-items: end;
}

.freehub-groupset-helper__field,
.freehub-groupset-helper__select {
  min-width: 0;
}

.freehub-groupset-helper__status {
  margin-top: 0.75rem;
}

.freehub-groupset-helper__facts {
  margin-top: 0.875rem;
}

.freehub-groupset-helper__fact-rail {
  display: grid;
  grid-auto-columns: 100%;
  grid-auto-flow: column;
  gap: 0.625rem;
  overflow-x: auto;
  overscroll-behavior-inline: contain;
  scroll-snap-type: x mandatory;
  scrollbar-width: none;
}

.freehub-groupset-helper__fact-rail::-webkit-scrollbar {
  display: none;
}

.freehub-groupset-helper__fact {
  min-width: 0;
  padding: 0.75rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.5rem;
  background: var(--tz-surface-panel);
  scroll-snap-align: start;
}

.freehub-groupset-helper__fact--xdr-xd {
  border-color: #d6a23d;
  background: #fffaf0;
}

.freehub-groupset-helper__fact--shimano-road-12-speed {
  border-color: #5aa88e;
  background: #f2fbf7;
}

.freehub-groupset-helper__fact-heading h4 {
  margin-top: 0.25rem;
  color: var(--tz-text-primary);
  font-size: 0.75rem;
  font-weight: 650;
  line-height: 1.35;
}

.freehub-groupset-helper__fact-label {
  color: var(--tz-text-muted);
  font-size: 0.625rem;
  font-weight: 600;
}

.freehub-groupset-helper__fact-callout {
  margin-top: 0.5rem;
  color: var(--tz-text-primary);
  font-size: 0.7rem;
  font-weight: 650;
  line-height: 1.45;
}

.freehub-groupset-helper__fact--xdr-xd .freehub-groupset-helper__fact-callout {
  color: #a16207;
}

.freehub-groupset-helper__fact--shimano-road-12-speed .freehub-groupset-helper__fact-callout {
  color: #047857;
}

.freehub-groupset-helper__fact-copy {
  margin-top: 0.25rem;
  color: var(--tz-text-secondary);
  font-size: 0.675rem;
  line-height: 1.5;
}

.freehub-groupset-helper__pagination {
  margin-top: 0.5rem;
}

@media (min-width: 768px) {
  .freehub-groupset-helper__fact-rail {
    grid-auto-columns: minmax(0, 1fr);
    grid-auto-flow: initial;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 22rem), 1fr));
    overflow: visible;
  }

  .freehub-groupset-helper__pagination {
    display: none;
  }
}

.freehub-groupset-helper select {
  color-scheme: light;
}

.freehub-groupset-helper select option {
  background-color: var(--tz-form-control-surface);
  color: var(--tz-text-primary);
}

.freehub-groupset-helper select option:disabled {
  color: var(--tz-text-muted);
}

</style>

