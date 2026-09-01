<template>
  <TabsContent value="carriers" class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-sm font-black tracking-tighter uppercase">承运商</h2>
        <p class="mt-1 text-xs text-muted-foreground">维护 DHL、FedEx、UPS、邮政小包、专线等物流公司基础资料。</p>
      </div>
      <Button v-if="canCreate" size="sm" @click="emit('create')">
        <Plus class="size-3.5" />
        新增承运商
      </Button>
    </div>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[1080px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-32">代码</TableHead>
            <TableHead>名称</TableHead>
            <TableHead class="w-44">联系人</TableHead>
            <TableHead>服务区域</TableHead>
            <TableHead class="w-24 text-right">排序</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-32 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="carriers.length === 0" :colspan="7">
            <div class="flex flex-col items-center text-muted-foreground">
              <Truck class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无承运商</span>
            </div>
          </TableEmpty>
          <TableRow v-for="carrier in carriers" :key="carrier.id">
 <TableCell class="font-mono text-xs font-bold">{{ carrier.code || '-'}}</TableCell>
            <TableCell>
 <span class="block font-bold text-xs">{{ carrier.name || '-'}}</span>
 <span class="block truncate text-xs text-muted-foreground">{{ carrier.tracking_url || '未配置查询链接'}}</span>
            </TableCell>
            <TableCell>
 <span class="block truncate text-sm">{{ carrier.contact || '-'}}</span>
 <span class="block truncate text-xs text-muted-foreground">{{ carrier.email || carrier.phone || '-'}}</span>
            </TableCell>
            <TableCell class="max-w-80 truncate text-xs text-muted-foreground">{{ serviceAreaLabel(carrier.service_area) }}</TableCell>
            <TableCell class="text-right tabular-nums">{{ carrier.sort_order || 0 }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="carrier.enabled ? 'green' : 'gray'">
                {{ carrier.enabled ? '启用' : '停用' }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-1">
                <Button
                  v-if="canEdit"
                  variant="ghost"
                  size="icon-sm"
                  :aria-label="`编辑承运商 ${carrier.name}`"
                  @click="emit('edit', carrier)"
                >
                  <Pencil class="size-4" />
                </Button>
                <Button
                  v-if="canDelete"
                  variant="ghost"
                  size="icon-sm"
                  class="text-destructive hover:text-destructive"
                  :aria-label="`删除承运商 ${carrier.name}`"
                  @click="emit('delete', carrier)"
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
import { Pencil, Plus, Trash2, Truck } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { TabsContent } from '@/components/ui/tabs'
import { serviceAreaLabel } from '@/lib/shippingPresentation'
import type { ShippingCarrier } from '@/modules/shipping/shippingTypes'

withDefaults(defineProps<{
  carriers?: ShippingCarrier[]
  loading?: boolean
  canCreate?: boolean
  canEdit?: boolean
  canDelete?: boolean
}>(), {
  carriers: () => [],
  loading: false,
  canCreate: false,
  canEdit: false,
  canDelete: false,
})

const emit = defineEmits<{
  (event: 'create'): void
  (event: 'edit', carrier: ShippingCarrier): void
  (event: 'delete', carrier: ShippingCarrier): void
}>()
</script>

