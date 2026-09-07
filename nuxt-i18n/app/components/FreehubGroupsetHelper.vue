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
          v-if="activeOption.notesKey"
          class="mt-0.5 tz-caption tz-text-muted"
        >
          {{ optionNotes(activeOption) }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, useId, watch } from 'vue'
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
}
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
  },
  {
    id: 'sram-gx-eagle',
    brand: 'SRAM',
    labelKey: 'options.sramGxEagle.label',
    freehubKey: 'options.sramGxEagle.freehub',
    notesKey: 'options.sramGxEagle.notes',
  },
  {
    id: 'sram-x01-eagle',
    brand: 'SRAM',
    labelKey: 'options.sramX01Eagle.label',
    freehubKey: 'options.sramX01Eagle.freehub',
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
  },
  {
    id: 'sram-red-etap-axs',
    brand: 'SRAM',
    labelKey: 'options.sramRedEtapAxs.label',
    freehubKey: 'options.sramRedEtapAxs.freehub',
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
const brandSelectId = computed(() => `${helperId}-freehub-brand`)
const groupsetSelectId = computed(() => `${helperId}-freehub-groupset`)

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

const optionLabel = (option: FreehubOption) => (
  t(`wheelsetFreehubHelper.${option.labelKey}`)
)
const optionFreehub = (option: FreehubOption) => (
  t(`wheelsetFreehubHelper.${option.freehubKey}`)
)
const optionNotes = (option: FreehubOption) => (
  option.notesKey ? t(`wheelsetFreehubHelper.${option.notesKey}`) : ''
)
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

