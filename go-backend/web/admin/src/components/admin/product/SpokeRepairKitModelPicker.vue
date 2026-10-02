<template>
  <div class="space-y-3">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h3 class="text-sm font-semibold text-foreground">适配轮组型号</h3>
        <p class="mt-1 text-xs leading-5 text-muted-foreground">从已核验的轮组型号目录中多选。选中的型号会直接成为前台买家下单时的型号选项。</p>
      </div>
      <span class="rounded-full bg-primary/10 px-2.5 py-1 text-xs font-semibold text-primary">已选 {{ selectedKeys.length }}</span>
    </div>

    <div v-if="loading" class="rounded-lg border border-dashed px-3 py-5 text-center text-xs text-muted-foreground">正在读取轮组型号目录…</div>
    <div v-else-if="error" class="flex items-center justify-between gap-3 rounded-lg border border-destructive/30 bg-destructive/5 px-3 py-3 text-xs text-destructive">
      <span>{{ error }}</span>
      <Button type="button" variant="outline" size="sm" @click="emit('retry')">重试</Button>
    </div>
    <template v-else>
      <Input v-model="search" placeholder="搜索品牌或轮组型号" />
      <div class="grid max-h-80 gap-2 overflow-y-auto rounded-lg border bg-muted/10 p-2 sm:grid-cols-2 xl:grid-cols-3">
        <label
          v-for="model in filteredModels"
          :key="buildSpokeRepairKitModelSelectionKey(model)"
          class="flex cursor-pointer items-start gap-2 rounded-md border bg-background px-3 py-2 transition hover:border-primary/50"
          :class="isSpokeRepairKitWheelsetModelSelected(model) ? 'border-primary bg-primary/5' : ''"
        >
          <input
            type="checkbox"
            class="mt-0.5 size-4 shrink-0 accent-primary"
            :checked="isSpokeRepairKitWheelsetModelSelected(model)"
            @change="toggleSpokeRepairKitWheelsetModelSelection(model)"
          >
          <span class="min-w-0">
            <span class="block truncate text-xs font-semibold text-foreground">{{ model.brandName }} / {{ model.model }}</span>
            <span class="mt-0.5 block truncate font-mono text-[10px] text-muted-foreground">{{ model.slug }}{{ model.lifecycleStatus === 'legacy' ? ' · 旧型号' : '' }}</span>
          </span>
        </label>
        <p v-if="filteredModels.length === 0" class="col-span-full py-5 text-center text-xs text-muted-foreground">没有匹配的轮组型号。</p>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { SpokeRepairKitCatalogModel } from '@/api/spokeRepairKitCatalog'

const props = withDefaults(defineProps<{
  models?: SpokeRepairKitCatalogModel[]
  selectedKeys?: string[]
  loading?: boolean
  error?: string
}>(), {
  models: () => [],
  selectedKeys: () => [],
  loading: false,
  error: ''
})

const emit = defineEmits<{
  (event: 'update:selectedKeys', value: string[]): void
  (event: 'retry'): void
}>()

const search = ref('')
const buildSpokeRepairKitModelSelectionKey = (model: SpokeRepairKitCatalogModel): string => `${model.brandSlug}:${model.slug}`.toLowerCase()
const selectedKeys = computed(() => props.selectedKeys)
const filteredModels = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  if (!keyword) return props.models
  return props.models.filter((model) => `${model.brandName} ${model.model} ${model.slug}`.toLowerCase().includes(keyword))
})
const isSpokeRepairKitWheelsetModelSelected = (model: SpokeRepairKitCatalogModel): boolean => selectedKeys.value.includes(buildSpokeRepairKitModelSelectionKey(model))
const toggleSpokeRepairKitWheelsetModelSelection = (model: SpokeRepairKitCatalogModel): void => {
  const key = buildSpokeRepairKitModelSelectionKey(model)
  const next = isSpokeRepairKitWheelsetModelSelected(model)
    ? selectedKeys.value.filter((item) => item !== key)
    : [...selectedKeys.value, key]
  emit('update:selectedKeys', next)
}
</script>
