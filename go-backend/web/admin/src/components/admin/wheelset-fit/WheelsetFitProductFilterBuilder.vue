<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <p class="text-[11px] font-bold text-muted-foreground">
        只从轮组商品里真实出现过的规格字段和值里选择，保存时会自动生成 `spec_filters`。
      </p>
      <Button type="button" size="sm" variant="outline" :disabled="!canAddRule" @click="addRow">
        <Plus class="size-3.5" />
        新增条件
      </Button>
    </div>

    <AdminFormField label="关键词" description="可选：同时保留 keyword 过滤。">
      <Input
        :model-value="keyword"
        :disabled="disabled || loading"
        placeholder="可选"
        @update:model-value="handleKeywordInput"
      />
    </AdminFormField>

    <p v-if="loading" class="text-[11px] font-bold text-muted-foreground">
      正在读取轮组商品动态值...
    </p>
    <p v-else-if="!filterOptions.length" class="text-[11px] font-bold text-muted-foreground">
      当前没有可用的轮组商品动态值，请先补充轮组商品参数。
    </p>

    <div v-if="rows.length" class="space-y-3">
      <section
        v-for="(row, index) in rows"
        :key="row.id"
        class="rounded-lg border border-border/60 bg-muted/10 p-3"
      >
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="flex items-center gap-2">
            <span class="grid size-6 shrink-0 place-items-center rounded-full bg-muted text-[11px] font-black">
              {{ index + 1 }}
            </span>
            <span class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/70">
              条件 {{ index + 1 }}
            </span>
          </div>
          <Button
            type="button"
            size="icon-sm"
            variant="ghost"
            class="text-destructive hover:text-destructive"
            :disabled="disabled || loading"
            :aria-label="`删除条件 ${index + 1}`"
            @click="removeRow(index)"
          >
            <Trash2 class="size-4" />
          </Button>
        </div>

        <div class="mt-3 grid gap-3 lg:grid-cols-[240px_minmax(0,1fr)]">
          <label class="space-y-1.5">
            <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
              规格字段
            </span>
            <select
              :value="row.slug"
              :disabled="disabled || loading"
              class="h-9 w-full rounded-md border border-input bg-background px-3 text-xs font-bold outline-none transition focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 disabled:cursor-not-allowed disabled:opacity-50"
              @change="handleFieldChange(row, $event)"
            >
              <option value="">请选择字段</option>
              <option v-if="row.slug && !hasKnownField(row.slug)" :value="row.slug">
                当前字段：{{ row.slug }}（轮组商品中未找到）
              </option>
              <option v-for="field in fieldOptionsForRow()" :key="field.slug" :value="field.slug">
                {{ fieldLabel(field) }}
              </option>
            </select>
          </label>

          <div class="space-y-2">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">可选值</p>
              <p v-if="fieldForRow(row)" class="text-[10px] font-bold text-muted-foreground">
                {{ fieldForRow(row)?.field_type }}
                <span v-if="fieldForRow(row)?.unit"> · {{ fieldForRow(row)?.unit }}</span>
              </p>
            </div>

            <div v-if="candidateValues(row).length" class="flex max-h-36 flex-wrap gap-2 overflow-y-auto pr-1">
              <button
                v-for="value in candidateValues(row)"
                :key="value.value"
                type="button"
                :disabled="disabled || loading"
                class="rounded-full border px-3 py-1 text-xs font-bold transition"
                :class="valueButtonClass(row, value)"
                @click="toggleValue(row, value.value)"
              >
                {{ value.value }}
              </button>
            </div>
            <p v-else class="rounded-md border border-dashed border-border/70 px-3 py-2 text-xs font-bold text-muted-foreground">
              先选择字段，再从真实轮组商品里选值。
            </p>
          </div>
        </div>
      </section>
    </div>
    <div v-else class="rounded-lg border border-dashed border-border/70 px-4 py-5 text-sm font-bold text-muted-foreground">
      当前还没有条件。
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Plus, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { WheelsetFitProductFilterOption } from '@/api/wheelsetFitQuestionnaire'

interface FilterRuleRow {
  id: number
  slug: string
  values: string[]
}

interface CandidateValue {
  value: string
  available: boolean
}

const props = withDefaults(defineProps<{
  modelValue?: string
  filterOptions: WheelsetFitProductFilterOption[]
  loading?: boolean
  disabled?: boolean
}>(), {
  modelValue: '{}',
  loading: false,
  disabled: false,
})

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void
}>()

const keyword = ref('')
const rows = ref<FilterRuleRow[]>([])
const extraFields = ref<Record<string, unknown>>({})
let nextRowID = 1

const canAddRule = computed(() => !props.disabled && !props.loading && props.filterOptions.length > 0)

const normalizeStringList = (value: unknown): string[] => {
  const items = Array.isArray(value)
    ? value
    : typeof value === 'string'
      ? [value]
      : []
  const seen = new Set<string>()
  const result: string[] = []
  for (const item of items) {
    const text = String(item).trim()
    if (!text || seen.has(text)) continue
    seen.add(text)
    result.push(text)
  }
  return result
}

const serializeModelValue = (): string => {
  const payload: Record<string, unknown> = { ...extraFields.value }
  const keywordValue = keyword.value.trim()
  if (keywordValue) {
    payload.keyword = keywordValue
  }

  const specFilters: Record<string, string[]> = {}
  for (const row of rows.value) {
    const slug = row.slug.trim()
    if (!slug) continue
    const values = normalizeStringList(row.values)
    if (values.length > 0) {
      specFilters[slug] = values
    }
  }
  if (Object.keys(specFilters).length > 0) {
    payload.spec_filters = specFilters
  }

  return JSON.stringify(payload, null, 2)
}

const emitModelValue = () => {
  const nextValue = serializeModelValue()
  if (nextValue !== props.modelValue) {
    emit('update:modelValue', nextValue)
  }
}

const parseModelValue = (raw: string) => {
  let parsed: unknown = {}
  try {
    parsed = JSON.parse(raw.trim() || '{}')
  } catch {
    parsed = {}
  }
  if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
    parsed = {}
  }

  const payload = parsed as Record<string, unknown>
  keyword.value = typeof payload.keyword === 'string' ? payload.keyword.trim() : ''
  extraFields.value = {}
  for (const [key, value] of Object.entries(payload)) {
    if (key === 'keyword' || key === 'spec_filters') continue
    extraFields.value[key] = value
  }

  const nextRows: FilterRuleRow[] = []
  const specFilters = payload.spec_filters
  if (specFilters && typeof specFilters === 'object' && !Array.isArray(specFilters)) {
    for (const [slug, value] of Object.entries(specFilters as Record<string, unknown>)) {
      const normalizedSlug = slug.trim()
      if (!normalizedSlug) continue
      nextRows.push({
        id: nextRowID++,
        slug: normalizedSlug,
        values: normalizeStringList(value),
      })
    }
  }
  rows.value = nextRows
}

watch(
  () => props.modelValue,
  (value) => {
    const raw = (value || '{}').trim() || '{}'
    if (raw === serializeModelValue()) {
      return
    }
    nextRowID = 1
    parseModelValue(raw)
  },
  { immediate: true },
)

const handleKeywordInput = (value: string) => {
  keyword.value = value
  emitModelValue()
}

const addRow = () => {
  rows.value.push({
    id: nextRowID++,
    slug: '',
    values: [],
  })
  emitModelValue()
}

const removeRow = (index: number) => {
  rows.value.splice(index, 1)
  emitModelValue()
}

const hasKnownField = (slug: string) => props.filterOptions.some((field) => field.slug === slug)

const fieldForRow = (row: FilterRuleRow): WheelsetFitProductFilterOption | undefined => (
  props.filterOptions.find((field) => field.slug === row.slug)
)

const fieldOptionsForRow = (): WheelsetFitProductFilterOption[] => {
  return props.filterOptions
}

const fieldLabel = (field: WheelsetFitProductFilterOption): string => {
  const parts = [field.name.trim() || field.slug]
  if (field.slug.trim() && field.slug !== field.name.trim()) {
    parts.push(field.slug)
  }
  if (field.field_type.trim()) {
    parts.push(field.field_type)
  }
  if (field.unit.trim()) {
    parts.push(field.unit)
  }
  return parts.filter(Boolean).join(' · ')
}

const candidateValues = (row: FilterRuleRow): CandidateValue[] => {
  const field = fieldForRow(row)
  const available = field?.values || []
  const selected = normalizeStringList(row.values)
  const result: CandidateValue[] = available.map((value) => ({
    value,
    available: true,
  }))
  for (const value of selected) {
    if (!available.includes(value)) {
      result.push({
        value,
        available: false,
      })
    }
  }
  return result.filter((item, index, self) => self.findIndex((entry) => entry.value === item.value) === index)
}

const valueButtonClass = (row: FilterRuleRow, candidate: CandidateValue): string => {
  const selected = row.values.includes(candidate.value)
  if (selected && candidate.available) {
    return 'border-primary bg-primary/10 text-primary'
  }
  if (selected) {
    return 'border-amber-500/40 bg-amber-500/10 text-amber-700'
  }
  if (candidate.available) {
    return 'border-border/70 bg-background text-foreground hover:border-primary/40'
  }
  return 'border-dashed border-amber-500/40 bg-background text-muted-foreground'
}

const toggleValue = (row: FilterRuleRow, value: string) => {
  const index = row.values.indexOf(value)
  if (index >= 0) {
    row.values.splice(index, 1)
  } else {
    row.values.push(value)
  }
  emitModelValue()
}

const handleFieldChange = (row: FilterRuleRow, event: Event) => {
  const target = event.target as HTMLSelectElement | null
  row.slug = target?.value || ''
  row.values = []
  emitModelValue()
}
</script>
