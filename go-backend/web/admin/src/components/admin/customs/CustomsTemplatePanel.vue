<template>
  <Card class="overflow-hidden">
    <CardHeader class="border-b">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <CardTitle>清关资料模板</CardTitle>
          <CardDescription>按商品规格模板、清关部件/产品族和材质维护可复用资料。</CardDescription>
        </div>
        <Button v-if="canCreate" variant="outline" size="sm" @click="emit('create')">
          <Plus class="size-3.5" />
          添加
        </Button>
      </div>
    </CardHeader>
    <CardContent class="p-0">
      <div v-if="loading" class="p-6 text-center text-sm text-muted-foreground">加载中...</div>
      <div v-else-if="templates.length" class="divide-y">
        <div v-for="template in templates" :key="template.id" class="flex items-start justify-between gap-3 p-4">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <p class="font-semibold">{{ template.name }}</p>
              <span class="rounded-full bg-muted px-2 py-0.5 text-[11px] text-muted-foreground">
                {{ template.status === 'paused' ? '暂停' : template.status === 'draft' ? '草稿' : '启用' }}
              </span>
            </div>
            <p class="mt-1 font-mono text-xs text-muted-foreground">
              {{ template.component_kind || '未分组件' }} · {{ template.material || '未分材质' }}
            </p>
            <p class="mt-2 font-mono text-xs text-foreground">
              HS {{ template.hs_code }}<span v-if="template.cn_code"> · CN {{ template.cn_code }}</span>
            </p>
            <p v-if="template.customs_description" class="mt-1 line-clamp-2 text-xs text-muted-foreground">
              {{ template.customs_description }}
            </p>
          </div>
          <div class="flex shrink-0 items-center gap-1">
            <Button
              v-if="canEdit"
              variant="ghost"
              size="icon"
              aria-label="编辑清关模板"
              @click="emit('edit', template)"
            >
              <Pencil class="size-4" />
            </Button>
            <Button
              v-if="canDelete"
              variant="ghost"
              size="icon"
              aria-label="删除清关模板"
              @click="emit('delete', template)"
            >
              <Trash2 class="size-4 text-destructive" />
            </Button>
          </div>
        </div>
      </div>
      <p v-else class="p-6 text-center text-sm text-muted-foreground">暂无清关模板。</p>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Pencil, Plus, Trash2 } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import type { CustomsClassificationRecord } from './customsClassificationTypes'

withDefaults(defineProps<{
  templates?: CustomsClassificationRecord[]
  loading?: boolean
  canCreate?: boolean
  canEdit?: boolean
  canDelete?: boolean
}>(), {
  templates: () => [],
  loading: false,
  canCreate: false,
  canEdit: false,
  canDelete: false,
})

const emit = defineEmits<{
  (event: 'create'): void
  (event: 'edit', template: CustomsClassificationRecord): void
  (event: 'delete', template: CustomsClassificationRecord): void
}>()
</script>
