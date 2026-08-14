<template>
  <div class="freehub-groupset-helper mt-5 rounded-2xl bg-slate-900/70 p-4 shadow-[3px_3px_10px_rgba(0,0,0,0.9)]">
    <h3 class="mb-2 text-sm font-semibold text-slate-100">
      {{ title }}
    </h3>
    <p class="mb-3 text-xs tz-text-secondary">
      {{ description }}
    </p>

    <div class="freehub-groupset-helper__fields">
      <div class="freehub-groupset-helper__field">
        <label class="block text-xs font-medium tz-text-secondary" :for="brandSelectId">
          Drivetrain brand
        </label>
        <select
          :id="brandSelectId"
          v-model="selectedBrand"
          class="freehub-groupset-helper__select mt-1 w-full rounded-md border border-slate-600/80 bg-slate-950/80 px-2 py-1.5 text-xs text-slate-100 shadow-[2px_2px_6px_rgba(0,0,0,0.85)] outline-none focus:border-sky-400 focus:ring-0"
        >
          <option disabled value="">
            Select brand
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
          Groupset
        </label>
        <select
          :id="groupsetSelectId"
          v-model="selectedGroupsetId"
          :disabled="!selectedBrand"
          class="freehub-groupset-helper__select mt-1 w-full rounded-md border border-slate-600/80 bg-slate-950/80 px-2 py-1.5 text-xs text-slate-100 shadow-[2px_2px_6px_rgba(0,0,0,0.85)] outline-none focus:border-sky-400 focus:ring-0 disabled:cursor-not-allowed disabled:border-slate-700 disabled:text-slate-500"
        >
          <option disabled value="">
            {{ selectedBrand ? 'Select groupset' : 'Choose brand first' }}
          </option>
          <option
            v-for="option in filteredGroupsets"
            :key="option.id"
            :value="option.id"
          >
            {{ option.label }}
          </option>
        </select>
      </div>
    </div>

    <div class="freehub-groupset-helper__status">
      <p
        v-if="!activeOption"
        class="text-xs tz-text-muted"
      >
        Choose a brand and groupset to see a suggested freehub body type.
      </p>

      <div
        v-else
        class="text-xs tz-text-secondary"
      >
        <p class="font-semibold text-sky-300">
          Recommended freehub body:
          <span class="ml-1">{{ activeOption.freehub }}</span>
        </p>
        <p
          v-if="activeOption.notes"
          class="mt-0.5 tz-caption tz-text-muted"
        >
          {{ activeOption.notes }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, useId } from 'vue'

withDefaults(defineProps<{
  title?: string
  description?: string
}>(), {
  title: 'Freehub body quick finder',
  description: 'Select your drivetrain brand and groupset to see which freehub body type is typically required. Always cross-check with the official compatibility charts from the drivetrain and hub manufacturers.',
})

interface FreehubOption {
  id: string
  brand: string
  label: string
  freehub: string
  notes?: string
}
// NOTE / 说明：
// 如果后续要增加更多套件 → 塔基类型的对应关系，直接在下面的 FREEHUB_OPTIONS 数组中追加一条对象即可。
// 不需要改其他文件；品牌下拉会根据 brand 字段自动生成，套件下拉会根据 brand 自动过滤。
const FREEHUB_OPTIONS: FreehubOption[] = [
  {
    id: 'shimano-deore-m6100',
    brand: 'Shimano',
    label: 'Deore M6100 (12-speed MTB)',
    freehub: 'MS (Micro Spline 12-speed)',
    notes: 'Typical for Shimano 12-speed MTB groupsets like Deore, SLX, XT, XTR.',
  },
  {
    id: 'shimano-slx-m7100',
    brand: 'Shimano',
    label: 'SLX M7100 (12-speed MTB)',
    freehub: 'MS (Micro Spline 12-speed)',
  },
  {
    id: 'shimano-xt-m8100',
    brand: 'Shimano',
    label: 'XT M8100 (12-speed MTB)',
    freehub: 'MS (Micro Spline 12-speed)',
  },
  {
    id: 'shimano-xtr-m9100',
    brand: 'Shimano',
    label: 'XTR M9100 (12-speed MTB)',
    freehub: 'MS (Micro Spline 12-speed)',
  },
  {
    id: 'shimano-105-r7000',
    brand: 'Shimano',
    label: '105 R7000 (11-speed road)',
    freehub: 'HG 9–11 speed road',
  },
  {
    id: 'shimano-ultegra-r8000',
    brand: 'Shimano',
    label: 'Ultegra R8000 (11-speed road)',
    freehub: 'HG 9–11 speed road',
  },
  {
    id: 'shimano-duraace-r9100',
    brand: 'Shimano',
    label: 'Dura-Ace R9100 (11-speed road)',
    freehub: 'HG 9–11 speed road',
  },
  {
    id: 'shimano-105-di2-r7100',
    brand: 'Shimano',
    label: '105 Di2 R7100 (12-speed road)',
    freehub: 'HG L2 12-speed road only',
    notes: 'Shimano 12-speed road specific body; not compatible with older 11-speed-only freehubs.',
  },
  {
    id: 'sram-gx-eagle',
    brand: 'SRAM',
    label: 'GX Eagle (12-speed MTB)',
    freehub: 'XD',
    notes: 'Use XD for Eagle 12-speed; NX Eagle cassette uses HG MTB freehub.',
  },
  {
    id: 'sram-x01-eagle',
    brand: 'SRAM',
    label: 'X01 Eagle (12-speed MTB)',
    freehub: 'XD',
  },
  {
    id: 'sram-nx-eagle',
    brand: 'SRAM',
    label: 'NX Eagle (12-speed MTB)',
    freehub: 'HG 8–11 speed MTB',
  },
  {
    id: 'sram-force-etap-axs',
    brand: 'SRAM',
    label: 'Force eTap AXS (12-speed road)',
    freehub: 'XDR',
  },
  {
    id: 'sram-red-etap-axs',
    brand: 'SRAM',
    label: 'Red eTap AXS (12-speed road)',
    freehub: 'XDR',
  },
  {
    id: 'campagnolo-ekar-13',
    brand: 'Campagnolo',
    label: 'Ekar 13-speed',
    freehub: 'N3W',
  },
  {
    id: 'campagnolo-super-record-11',
    brand: 'Campagnolo',
    label: 'Super Record 11-speed road',
    freehub: 'CP (Campagnolo classic)',
  },
]

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
  color-scheme: dark;
}

.freehub-groupset-helper select option {
  background-color: #151515;
  color: #f8fafc;
}

.freehub-groupset-helper select option:disabled {
  color: #a1a1aa;
}

</style>

