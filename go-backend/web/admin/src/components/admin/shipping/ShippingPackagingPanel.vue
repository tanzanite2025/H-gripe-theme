<template>
  <TabsContent value="packaging" class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-sm font-black tracking-tighter uppercase">包装规则</h2>
        <p class="mt-1 text-xs text-muted-foreground">管理包装箱重量、尺寸、最大承重和适用商品；当前先保持产品级事实源，SKU 级包装规则后续统一升级。</p>
      </div>
      <Button v-if="canCreate" size="sm" @click="emit('create')">
        <Plus class="size-3.5" />
        新增包装规则
      </Button>
    </div>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[980px]">
        <TableHeader>
          <TableRow>
            <TableHead>规则名称</TableHead>
            <TableHead class="w-28 text-right">包装重量</TableHead>
            <TableHead class="w-44">尺寸</TableHead>
            <TableHead class="w-28 text-right">最大承重</TableHead>
            <TableHead class="w-28 text-right">绑定商品</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-32 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="packagingRules.length === 0" :colspan="7">
            <div class="flex flex-col items-center text-muted-foreground">
              <Package class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无包装规则</span>
            </div>
          </TableEmpty>
          <TableRow v-for="rule in packagingRules" :key="rule.id">
            <TableCell>
 <span class="block font-bold text-xs">{{ rule.rule_name || '-'}}</span>
 <span class="block max-w-96 truncate text-xs text-muted-foreground">{{ rule.description || '暂无说明'}}</span>
            </TableCell>
            <TableCell class="text-right tabular-nums">{{ formatWeight(rule.box_weight) }}</TableCell>
            <TableCell class="font-mono text-xs text-muted-foreground">{{ formatDimensions(rule) }}</TableCell>
            <TableCell class="text-right tabular-nums">{{ formatWeight(rule.max_weight) }}</TableCell>
            <TableCell class="text-right tabular-nums">{{ appliesCount(rule) }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="rule.is_active ? 'green' : 'gray'">
                {{ rule.is_active ? '启用' : '停用' }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-1">
                <Button
                  v-if="canEdit"
                  variant="ghost"
                  size="icon-sm"
                  :aria-label="`维护包装规则 ${rule.rule_name} 的适用商品`"
                  @click="emit('show-applies', rule)"
                >
                  <Link2 class="size-4" />
                </Button>
                <Button
                  v-if="canEdit"
                  variant="ghost"
                  size="icon-sm"
                  :aria-label="`编辑包装规则 ${rule.rule_name}`"
                  @click="emit('edit', rule)"
                >
                  <Pencil class="size-4" />
                </Button>
                <Button
                  v-if="canDelete"
                  variant="ghost"
                  size="icon-sm"
                  class="text-destructive hover:text-destructive"
                  :aria-label="`删除包装规则 ${rule.rule_name}`"
                  @click="emit('delete', rule)"
                >
                  <Trash2 class="size-4" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>
  </TabsContent>
</template>

<script setup lang="ts">
import { Link2, Package, Pencil, Plus, Trash2 } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { TabsContent } from '@/components/ui/tabs'
import { appliesCount, formatDimensions, formatWeight } from '@/lib/shippingPresentation'
import type { PackagingRule } from './shippingTypes'

withDefaults(defineProps<{
  packagingRules?: PackagingRule[]
  loading?: boolean
  canCreate?: boolean
  canEdit?: boolean
  canDelete?: boolean
}>(), {
  packagingRules: () => [],
  loading: false,
  canCreate: false,
  canEdit: false,
  canDelete: false
})

const emit = defineEmits<{
  (event: 'create'): void
  (event: 'edit', rule: PackagingRule): void
  (event: 'delete', rule: PackagingRule): void
  (event: 'show-applies', rule: PackagingRule): void
}>()
</script>
