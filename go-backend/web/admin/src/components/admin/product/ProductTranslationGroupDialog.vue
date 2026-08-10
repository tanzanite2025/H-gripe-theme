<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[88dvh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>商品翻译组</DialogTitle>
        <DialogDescription v-if="currentProduct">
          {{ currentProduct.name }} · {{ localeName(currentProduct.locale) }}
        </DialogDescription>
      </DialogHeader>

      <div v-if="translationGroup" class="space-y-5">
        <div class="flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-muted/20 px-4 py-3">
          <div>
            <p class="text-xs font-bold">语言覆盖</p>
            <p class="mt-1 text-[11px] text-muted-foreground">
              已覆盖 {{ translationGroup.translations.length }} / {{ totalLanguages }} 个启用语言
            </p>
          </div>
          <AdminStatusBadge
            :tone="translationGroup.missing_locales.length > 0 ? 'amber' : 'green'"
          >
            {{ translationGroup.missing_locales.length > 0 ? `缺少 ${translationGroup.missing_locales.length} 个语言` : '已覆盖全部语言' }}
          </AdminStatusBadge>
        </div>

        <section class="space-y-2">
          <div class="flex items-center justify-between gap-3">
            <div>
              <h3 class="text-sm font-bold">已加入翻译组</h3>
              <p class="text-[11px] text-muted-foreground">每个语言版本都是独立商品记录，共享同一个根商品。</p>
            </div>
            <span class="font-mono text-[10px] text-muted-foreground">ROOT #{{ translationGroup.root_id }}</span>
          </div>

          <div class="relative min-h-28 overflow-x-auto rounded-lg border">
            <div v-if="translationsLoading" class="absolute inset-0 z-10 flex items-center justify-center bg-background/80">
              <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载翻译组" />
            </div>
            <Table class="min-w-[680px]">
              <TableHeader>
                <TableRow>
                  <TableHead class="w-28">语言</TableHead>
                  <TableHead>商品</TableHead>
                  <TableHead class="w-48">Slug</TableHead>
                  <TableHead class="w-40">SKU</TableHead>
                  <TableHead class="w-24">状态</TableHead>
                  <TableHead class="w-20 text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableEmpty v-if="!translationsLoading && translationGroup.translations.length === 0" :colspan="6">
                  暂无翻译版本
                </TableEmpty>
                <TableRow v-for="translation in translationGroup.translations" :key="translation.id">
                  <TableCell>
                    <div class="flex items-center gap-1.5">
                      <span v-if="translation.is_root" class="font-mono text-[10px] text-admin-selected">ROOT</span>
                      <span class="font-bold text-xs">{{ localeName(translation.locale) }}</span>
                    </div>
                  </TableCell>
                  <TableCell class="max-w-64 truncate text-xs font-bold">{{ translation.name || '-' }}</TableCell>
                  <TableCell class="max-w-48 truncate font-mono text-[10px] text-muted-foreground">{{ translation.slug || '-' }}</TableCell>
                  <TableCell class="max-w-40 truncate font-mono text-[10px] text-muted-foreground">{{ translation.sku || '-' }}</TableCell>
                  <TableCell>
                    <AdminStatusBadge :tone="statusTone(translation.status)">
                      {{ statusName(translation.status) }}
                    </AdminStatusBadge>
                  </TableCell>
                  <TableCell class="text-right">
                    <Button
                      v-if="canEdit"
                      variant="ghost"
                      size="icon"
                      :aria-label="`编辑 ${translation.name || '商品翻译'}`"
                      @click="emit('edit', translation)"
                    >
                      <Pencil class="size-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
        </section>

        <section class="space-y-2 border-t border-dashed pt-5">
          <div>
            <h3 class="text-sm font-bold">缺失语言</h3>
            <p class="text-[11px] text-muted-foreground">复制会创建独立商品，并自动生成不冲突的 slug 与 SKU。</p>
          </div>
          <div v-if="translationGroup.missing_locales.length > 0" class="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
            <div
              v-for="locale in translationGroup.missing_locales"
              :key="locale"
              class="flex items-center justify-between gap-2 rounded-lg border px-3 py-2"
            >
              <span class="text-xs font-bold">{{ localeName(locale) }}</span>
              <Button
                v-if="canCreateTranslation"
                size="sm"
                :disabled="Boolean(copyingLocale)"
                @click="emit('copy', locale)"
              >
                <LoaderCircle v-if="copyingLocale === locale" class="size-3.5 animate-spin" />
                <Copy v-else class="size-3.5" />
                {{ copyingLocale === locale ? '复制中' : '复制到' }}
              </Button>
              <span v-else class="text-[10px] text-muted-foreground">无创建权限</span>
            </div>
          </div>
          <div v-else class="flex items-center gap-2 rounded-lg border border-emerald-500/20 bg-emerald-500/5 px-3 py-2 text-xs text-emerald-700 dark:text-emerald-300">
            <CircleCheck class="size-4" />
            当前翻译组已覆盖所有启用语言。
          </div>
        </section>
      </div>

      <div v-else-if="!translationsLoading" class="rounded-lg border border-dashed px-4 py-8 text-center text-xs text-muted-foreground">
        暂无翻译组数据
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { CircleCheck, Copy, LoaderCircle, Pencil } from '@lucide/vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type { ProductRecord, ProductTranslation, ProductTranslationGroup } from './productEditorTypes'

withDefaults(defineProps<{
  open?: boolean
  currentProduct?: ProductRecord | null
  translationGroup?: ProductTranslationGroup | null
  translationsLoading?: boolean
  copyingLocale?: string
  canEdit?: boolean
  canCreateTranslation?: boolean
  localeName: (locale?: string | null) => string
  totalLanguages?: number
}>(), {
  open: false,
  currentProduct: null,
  translationGroup: null,
  translationsLoading: false,
  copyingLocale: '',
  canEdit: false,
  canCreateTranslation: false,
  totalLanguages: 20
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'copy', locale: string): void
  (event: 'edit', translation: ProductTranslation): void
}>()

const statusNames: Record<string, string> = {
  active: '在售',
  inactive: '下架',
  out_of_stock: '缺货'
}
const statusTones: Record<string, AdminStatusTone> = {
  active: 'green',
  inactive: 'gray',
  out_of_stock: 'coral'
}
const statusName = (status?: string | null): string => statusNames[status || ''] || status || '-'
const statusTone = (status?: string | null): AdminStatusTone => statusTones[status || ''] || 'gray'
</script>
