<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-h-[88dvh] max-w-4xl overflow-hidden p-0">
      <DialogHeader class="border-b px-5 py-4 pr-12">
        <DialogTitle>同步商品规格模板</DialogTitle>
        <DialogDescription>
          当前商品不会自动跟随模板变化。请核对差异后确认，确认时会物化模板当前版本。
        </DialogDescription>
      </DialogHeader>

      <div v-if="diff" class="min-h-0 space-y-4 overflow-y-auto px-5 py-4">
        <div class="grid gap-2 sm:grid-cols-3">
          <div class="rounded-lg border bg-muted/20 px-3 py-2 text-xs">
            <div class="text-muted-foreground">模板版本</div>
            <strong class="font-mono text-sm">{{ diff.materialized_revision }} → {{ diff.template_revision }}</strong>
          </div>
          <div class="rounded-lg border bg-muted/20 px-3 py-2 text-xs">
            <div class="text-muted-foreground">候选值变更</div>
            <strong class="font-mono text-sm">{{ diff.items.length }}</strong>
          </div>
          <div class="rounded-lg border bg-muted/20 px-3 py-2 text-xs">
            <div class="text-muted-foreground">价格 / 标签变更</div>
            <strong class="font-mono text-sm">{{ diff.summary.price_changed }} / {{ diff.summary.label_changed }}</strong>
          </div>
        </div>

        <div v-if="!diff.has_changes" class="rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-4 py-5 text-center text-sm text-emerald-700 dark:text-emerald-200">
          当前商品已经是模板最新版本。
        </div>
        <div v-else class="overflow-x-auto rounded-lg border">
          <table class="w-full min-w-[680px] text-left text-xs">
            <thead class="bg-muted/40 text-muted-foreground">
              <tr>
                <th class="px-3 py-2 font-semibold">选项组</th>
                <th class="px-3 py-2 font-semibold">候选值</th>
                <th class="px-3 py-2 font-semibold">变化</th>
                <th class="px-3 py-2 font-semibold">标签</th>
                <th class="px-3 py-2 font-semibold">加价（最小单位）</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in diff.items" :key="`${item.spec_definition_id}:${item.value_key}`" class="border-t">
                <td class="px-3 py-2">
                  <div class="font-medium">{{ item.group_name || item.group_slug }}</div>
                  <div class="font-mono text-[10px] text-muted-foreground">{{ item.group_slug }}</div>
                </td>
                <td class="px-3 py-2 font-mono">{{ item.value_key }}</td>
                <td class="px-3 py-2">
                  <div class="flex flex-wrap gap-1">
                    <span v-for="change in item.changes" :key="change" class="rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-medium text-primary">
                      {{ changeLabel(change) }}
                    </span>
                  </div>
                </td>
                <td class="px-3 py-2">
                  <span v-if="item.current_label && item.template_label && item.current_label !== item.template_label">{{ item.current_label }} → {{ item.template_label }}</span>
                  <span v-else>{{ item.template_label || item.current_label || '未设置' }}</span>
                </td>
                <td class="px-3 py-2 font-mono">
                  <span v-if="item.current_price_delta_minor !== undefined || item.template_price_delta_minor !== undefined">
                    {{ item.current_price_delta_minor ?? 0 }} → {{ item.template_price_delta_minor ?? 0 }}
                  </span>
                  <span v-else>—</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div v-else class="px-5 py-8 text-center text-sm text-muted-foreground">正在读取模板差异…</div>

      <DialogFooter class="border-t px-5 py-3">
        <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
        <Button type="button" :disabled="!diff || !diff.has_changes || applying" @click="emit('confirm')">
          <LoaderCircle v-if="applying" class="size-4 animate-spin" />
          <RefreshCw v-else class="size-4" />
          {{ applying ? '同步中' : '确认同步' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { LoaderCircle, RefreshCw } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import type { ProductTemplateSyncDiff } from '@/modules/product/productEditorTypes'

withDefaults(defineProps<{
  open?: boolean
  diff?: ProductTemplateSyncDiff | null
  applying?: boolean
}>(), {
  open: false,
  diff: null,
  applying: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'confirm'): void
}>()

const changeLabels: Record<string, string> = {
  added: '新增',
  disabled: '禁用',
  enabled: '启用',
  label_changed: '标签变化',
  price_changed: '价格变化',
  removed: '已移除',
}

const changeLabel = (value: string): string => changeLabels[value] || value
</script>
