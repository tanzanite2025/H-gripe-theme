<template>
  <TabsContent value="routeAggregates" class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-sm font-black tracking-tighter uppercase">聚合线路</h2>
        <p class="mt-1 text-xs text-muted-foreground">读取并整理 4PX、燕文等承运商已发布的线路身份和覆盖国家。这里不配置客户运费。</p>
      </div>
      <Button v-if="canCreate" size="sm" @click="emit('create')">
        <Plus class="size-3.5" />
        新增聚合线路
      </Button>
    </div>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[980px]">
        <TableHeader>
          <TableRow>
            <TableHead>聚合线路</TableHead>
            <TableHead>承运商线路身份</TableHead>
            <TableHead>国家/区域</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-32 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="routeAggregates.length === 0" :colspan="5">
            <div class="flex flex-col items-center text-muted-foreground">
              <Route class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无聚合线路</span>
            </div>
          </TableEmpty>
          <TableRow v-for="aggregate in routeAggregates" :key="aggregate.id">
            <TableCell>
              <span class="block text-xs font-bold">{{ aggregate.name || '-' }}</span>
              <span class="block max-w-80 truncate text-[10px] text-muted-foreground/70">{{ aggregate.description || '暂无说明' }}</span>
            </TableCell>
            <TableCell class="max-w-[30rem] text-xs">
              <div v-if="aggregate.carrier_services?.length" class="flex flex-wrap gap-1">
                <span v-for="service in aggregate.carrier_services" :key="service.id" class="rounded border px-1.5 py-0.5">
                  {{ carrierServiceLabel(service) }}
                </span>
              </div>
              <span v-else class="text-muted-foreground">未选择已发布线路</span>
            </TableCell>
            <TableCell class="max-w-56 truncate font-mono text-xs text-muted-foreground">{{ countryLabel(aggregate.countries) }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="aggregate.enabled ? 'green' : 'gray'">{{ aggregate.enabled ? '启用' : '停用' }}</AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-1">
                <Button v-if="canEdit" variant="ghost" size="icon-sm" :aria-label="`编辑聚合线路 ${aggregate.name}`" @click="emit('edit', aggregate)">
                  <Pencil class="size-4" />
                </Button>
                <Button v-if="canDelete" variant="ghost" size="icon-sm" class="text-destructive hover:text-destructive" :aria-label="`删除聚合线路 ${aggregate.name}`" @click="emit('delete', aggregate)">
                  <Trash2 class="size-4" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>

    <div class="rounded-lg border border-blue-500/25 bg-blue-500/5 p-3 text-xs leading-5 text-muted-foreground">
      线路覆盖范围来自承运商发布数据；需要设置买家支付金额时，请到“线路计费模板”配置。两者的数据和职责完全分开。
    </div>
  </TabsContent>
</template>

<script setup lang="ts">
import { Pencil, Plus, Route, Trash2 } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { TabsContent } from '@/components/ui/tabs'
import type { ShippingCarrierService, ShippingRouteAggregate } from '@/modules/shipping/shippingTypes'

withDefaults(defineProps<{
  routeAggregates?: ShippingRouteAggregate[]
  loading?: boolean
  canCreate?: boolean
  canEdit?: boolean
  canDelete?: boolean
}>(), {
  routeAggregates: () => [],
  loading: false,
  canCreate: false,
  canEdit: false,
  canDelete: false,
})

const emit = defineEmits<{
  (event: 'create'): void
  (event: 'edit', aggregate: ShippingRouteAggregate): void
  (event: 'delete', aggregate: ShippingRouteAggregate): void
}>()

const countryLabel = (value: unknown) => {
  if (Array.isArray(value)) return value.join(', ') || '未限制'
  const raw = String(value || '').trim()
  if (!raw) return '未限制'
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) return parsed.join(', ') || '未限制'
  } catch {
    // Keep the original value visible when an older record is malformed.
  }
  return raw.replace(/[\[\]"']/g, '').replace(/[,，;|]+/g, ', ')
}

const carrierServiceLabel = (service: ShippingCarrierService) => {
  const carrier = service.carrier?.name || service.provider_code || '承运商'
  return `${carrier} · ${service.service_name || service.service_code || '-'}`
}
</script>
