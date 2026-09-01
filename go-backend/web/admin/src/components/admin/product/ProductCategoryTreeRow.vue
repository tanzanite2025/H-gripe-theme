<template>
  <div
    class="product-category-tree-row grid gap-3 px-4 py-3"
 :class="row.is_new || row.dirty ? 'bg-amber-50/60': ''"
  >
    <div class="flex min-w-0 items-center gap-2" :style="{ paddingLeft: `${Math.max(row.depth - 1, 0) * 24}px` }">
      <CornerDownRight v-if="row.depth > 1" class="size-4 shrink-0 text-muted-foreground" />
      <Folder class="size-4 shrink-0 text-muted-foreground" />
      <span class="shrink-0 rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">L{{ row.depth }}</span>
      <Input :model-value="row.name" placeholder="分类名称" @update:model-value="updateField('name', String($event))" />
    </div>

    <div class="flex min-w-0 items-center gap-2">
      <button
        type="button"
        class="product-category-image-preview"
        :aria-label="row.image_url ? '更换分类图片' : '选择分类图片'"
        :title="row.image_url ? '更换分类图片' : '选择分类图片'"
        :disabled="saving"
        @click="emit('pick-image', row)"
      >
        <img v-if="row.image_url" :src="row.image_url" alt="" loading="lazy" />
        <ImagePlus v-else class="size-4 text-muted-foreground" />
      </button>
      <Button
        v-if="row.image_url"
        variant="ghost"
        size="icon"
        class="size-8 shrink-0"
        aria-label="移除分类图片"
        title="移除分类图片"
        :disabled="saving"
        @click="emit('clear-image', row)"
      >
        <X class="size-3.5 text-muted-foreground" />
      </Button>
    </div>

    <Input :model-value="row.slug" class="min-w-0 font-mono" placeholder="slug" @update:model-value="updateField('slug', String($event))" @blur="emit('normalize-slug', row)" />

    <div class="min-w-0">
      <Select :model-value="row.parent_key || rootParentValue" :disabled="saving" @update:model-value="emit('change-parent', row, String($event))">
        <SelectTrigger class="w-full"><SelectValue placeholder="顶级分类" /></SelectTrigger>
        <SelectContent>
          <SelectItem :value="rootParentValue">顶级分类</SelectItem>
          <SelectItem
            v-for="option in parentOptions"
            :key="option.key"
            :value="option.key"
            :disabled="option.disabled"
          >
            {{ `${'　'.repeat(Math.max(option.depth - 1, 0))}${option.name || '未命名分类'}` }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="min-w-0">
      <Input :model-value="row.description" placeholder="可选描述" @update:model-value="updateField('description', String($event))" />
    </div>

    <div
      class="flex flex-col items-center justify-center gap-0.5 text-center"
      :title="translationStatusTitle(row)"
    >
      <span class="font-mono text-xs font-semibold">
        {{ row.translation_completed }}/{{ row.translation_total || '-' }}
      </span>
      <span
        class="text-[11px] font-medium"
 :class="translationMissingCount(row) ? 'text-amber-600': 'text-emerald-600'"
      >
        {{ translationStatusLabel(row) }}
      </span>
    </div>

    <div class="flex items-center justify-center">
      <Switch
        :model-value="row.is_enabled"
        size="sm"
        :aria-label="row.is_enabled ? '停用分类' : '启用分类'"
        :disabled="saving"
        @update:model-value="updateField('is_enabled', Boolean($event))"
      />
    </div>

    <div class="flex items-center justify-end gap-1">
      <Button
        v-if="canTranslate && row.id"
        variant="ghost"
        size="icon"
        aria-label="编辑分类翻译"
        title="编辑分类翻译"
        :disabled="saving"
        @click="emit('edit-translations', row)"
      >
        <Languages class="size-4" />
      </Button>
      <Button v-if="canCreate" variant="ghost" size="icon" aria-label="添加同级分类" :disabled="saving" @click="emit('add-sibling', row)">
        <PlusSquare class="size-4" />
      </Button>
      <Button v-if="canCreate" variant="ghost" size="icon" aria-label="添加下级分类" :disabled="saving || row.depth >= maxDepth" @click="emit('add-child', row)">
        <ListPlus class="size-4" />
      </Button>
      <Button v-if="canDelete" variant="ghost" size="icon" aria-label="删除分类" :disabled="saving" @click="emit('delete', row)">
        <Trash2 class="size-4 text-destructive" />
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { CornerDownRight, Folder, ImagePlus, Languages, ListPlus, PlusSquare, Trash2, X } from '@lucide/vue'
import type { DraftCategoryRow, ProductCategoryParentOption } from '@/modules/product/productCategoryTypes'
import { rootProductCategoryParentValue } from '@/composables/product/useProductCategoryTreeEditor'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

const rootParentValue = rootProductCategoryParentValue

const props = defineProps<{
  row: DraftCategoryRow
  parentOptions: ProductCategoryParentOption[]
  maxDepth: number
  saving: boolean
  canCreate: boolean
  canDelete: boolean
  canTranslate: boolean
}>()

const emit = defineEmits<{
  'add-sibling': [row: DraftCategoryRow]
  'add-child': [row: DraftCategoryRow]
  'change-parent': [row: DraftCategoryRow, value: string]
  'pick-image': [row: DraftCategoryRow]
  'clear-image': [row: DraftCategoryRow]
  'delete': [row: DraftCategoryRow]
  'edit-translations': [row: DraftCategoryRow]
  'normalize-slug': [row: DraftCategoryRow]
  'mark-dirty': [row: DraftCategoryRow]
}>()

const updateField = <K extends keyof DraftCategoryRow>(field: K, value: DraftCategoryRow[K]) => {
  props.row[field] = value
  emit('mark-dirty', props.row)
}

const translationMissingCount = (row: DraftCategoryRow): number => {
  if (!row.translation_total) return 0
  return Math.max(row.translation_total - row.translation_completed, 0)
}

const translationStatusLabel = (row: DraftCategoryRow): string => {
  if (!row.translation_total) return row.is_new ? '保存后填写' : '暂无语言'
  const missing = translationMissingCount(row)
  return missing ? `缺 ${missing} 种` : '已完成'
}

const translationStatusTitle = (row: DraftCategoryRow): string => {
  if (!row.translation_total) return '保存分类后可填写翻译'
  const missing = row.translation_missing_locales || []
  return missing.length
    ? `缺少语言：${missing.join(', ')}`
    : '已完成全部系统语言翻译'
}
</script>

<style scoped>
.product-category-tree-row {
  width: 100%;
  min-width: 0;
  grid-template-columns: minmax(220px, 1.3fr) minmax(128px, 0.5fr) minmax(140px, 0.8fr) minmax(160px, 0.9fr) minmax(160px, 1fr) minmax(92px, 0.5fr) 56px minmax(140px, max-content);
  align-items: center;
}

.product-category-image-preview {
  display: inline-flex;
  width: 88px;
  height: 50px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border: 1px solid hsl(var(--border));
  border-radius: 8px;
  background: hsl(var(--muted) / 0.55);
  transition: border-color 0.18s ease, background-color 0.18s ease;
}

.product-category-image-preview:hover,
.product-category-image-preview:focus-visible {
  border-color: hsl(var(--ring));
  background: hsl(var(--muted));
  outline: none;
}

.product-category-image-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

@media (max-width: 1279px) {
  .product-category-tree-row {
    grid-template-columns: 1fr;
    align-items: stretch;
  }
}

</style>

