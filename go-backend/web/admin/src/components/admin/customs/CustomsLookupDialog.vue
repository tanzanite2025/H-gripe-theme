<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="max-w-4xl">
      <DialogHeader>
        <DialogTitle>免费编码查询</DialogTitle>
        <DialogDescription>查询结果只作为候选资料，生成模板前请结合实际材质和目的地规则确认。</DialogDescription>
      </DialogHeader>

      <div class="space-y-4">
        <div class="grid gap-3 md:grid-cols-[12rem_minmax(0,1fr)_auto]">
          <Select :model-value="provider" @update:model-value="emit('update:provider', String($event))">
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="us_hts">US HTS</SelectItem>
              <SelectItem value="uk_trade_tariff">UK Trade Tariff</SelectItem>
            </SelectContent>
          </Select>
          <Input
            :model-value="query"
            :placeholder="provider === 'uk_trade_tariff' ? '输入 8 或 10 位 commodity code' : '输入英文商品描述或关键词'"
            @update:model-value="emit('update:query', String($event))"
            @keyup.enter="emit('run')"
          />
          <Button :disabled="loading || !query.trim()" @click="emit('run')">
            <LoaderCircle v-if="loading" class="size-4 animate-spin" />
            <FileSearch v-else class="size-4" />
            查询
          </Button>
        </div>

        <div v-if="candidates.length" class="max-h-[52dvh] overflow-y-auto pr-1">
          <div class="grid gap-3 lg:grid-cols-2">
            <div v-for="candidate in candidates" :key="`${candidate.provider}-${candidate.source_code}`" class="rounded-lg border p-3">
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <p class="font-mono text-sm font-semibold">{{ candidate.source_code }}</p>
                  <p class="mt-1 text-sm">{{ candidate.description }}</p>
                  <p class="mt-2 text-xs text-muted-foreground">
                    HS {{ candidate.hs_code }}<span v-if="candidate.cn_code"> · CN {{ candidate.cn_code }}</span>
                    <span v-if="candidate.duty"> · {{ candidate.duty }}</span>
                  </p>
                </div>
                <Button v-if="canCreate" variant="outline" size="sm" @click="emit('create-template', candidate)">
                  <Plus class="size-3.5" />
                  生成模板
                </Button>
              </div>
            </div>
          </div>
        </div>
        <p v-else-if="completed" class="text-sm text-muted-foreground">没有返回候选结果。</p>
        <div v-else class="rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground">
          输入商品描述或编码后开始查询。
        </div>
      </div>

      <DialogFooter>
        <Button type="button" variant="outline" @click="emit('update:open', false)">关闭</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { FileSearch, LoaderCircle, Plus } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import type { LookupCandidate } from '@/modules/customs/customsClassificationTypes'

withDefaults(defineProps<{
  open: boolean
  provider: string
  query: string
  loading?: boolean
  completed?: boolean
  candidates?: LookupCandidate[]
  canCreate?: boolean
}>(), {
  loading: false,
  completed: false,
  candidates: () => [],
  canCreate: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'update:provider', value: string): void
  (event: 'update:query', value: string): void
  (event: 'run'): void
  (event: 'create-template', candidate: LookupCandidate): void
}>()
</script>

