<template>
  <TabsContent value="services" class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-sm font-black tracking-tighter uppercase">线路服务</h2>
        <p class="mt-1 text-xs text-muted-foreground">维护承运商下的国际线路、计费口径、首续重、体积重和预计时效。</p>
      </div>
      <Button
        v-if="canCreate"
        size="sm"
        :disabled="carriers.length === 0"
        @click="emit('create')"
      >
        <Plus class="size-3.5" />
        新增线路服务
      </Button>
    </div>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[1280px]">
        <TableHeader>
          <TableRow>
            <TableHead>承运商 / 线路</TableHead>
            <TableHead>关联模板</TableHead>
            <TableHead>国家/区域</TableHead>
            <TableHead class="w-40">计费模式</TableHead>
            <TableHead class="w-48">首续重</TableHead>
            <TableHead class="w-40">体积重</TableHead>
            <TableHead class="w-28">时效</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-32 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="carrierServices.length === 0" :colspan="9">
            <div class="flex flex-col items-center text-muted-foreground">
              <Route class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无线路服务</span>
            </div>
          </TableEmpty>
          <TableRow v-for="service in carrierServices" :key="service.id">
            <TableCell>
 <span class="block font-bold text-xs">{{ service.service_name || '-'}}</span>
              <span class="block font-mono text-[10px] text-muted-foreground/70">
                {{ carrierServiceCarrierName(service, carriers) }} · {{ service.service_code || '-' }}
              </span>
            </TableCell>
            <TableCell>
              <span class="block font-bold text-xs">{{ carrierServiceTemplateName(service, templates) }}</span>
 <span class="block text-[10px] text-muted-foreground/70">{{ service.route_name || '未填写线路渠道'}}</span>
            </TableCell>
            <TableCell class="max-w-72 truncate text-xs text-muted-foreground">{{ compactListLabel(service.countries) }}</TableCell>
            <TableCell>{{ billingModeLabel(service.billing_mode) }}</TableCell>
            <TableCell class="font-mono text-xs text-muted-foreground">
              {{ formatServiceWeightStep(service) }}
            </TableCell>
            <TableCell class="font-mono text-xs text-muted-foreground">
              {{ formatVolumetricDivisor(service) }}
            </TableCell>
            <TableCell class="text-xs text-muted-foreground">{{ formatEta(service) }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="service.enabled ? 'green' : 'gray'">
                {{ service.enabled ? '启用' : '停用' }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-1">
                <Button
                  v-if="canEdit"
                  variant="ghost"
                  size="icon-sm"
                  :aria-label="`编辑线路服务 ${service.service_name}`"
                  @click="emit('edit', service)"
                >
                  <Pencil class="size-4" />
                </Button>
                <Button
                  v-if="canDelete"
                  variant="ghost"
                  size="icon-sm"
                  class="text-destructive hover:text-destructive"
                  :aria-label="`删除线路服务 ${service.service_name}`"
                  @click="emit('delete', service)"
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
import { Pencil, Plus, Route, Trash2 } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { TabsContent } from '@/components/ui/tabs'
import {
  billingModeLabel,
  carrierServiceCarrierName,
  carrierServiceTemplateName,
  compactListLabel,
  formatEta,
  formatServiceWeightStep,
  formatVolumetricDivisor,
} from '@/lib/shippingPresentation'
import type {
  ShippingCarrier,
  ShippingCarrierService,
  ShippingTemplate
} from './shippingTypes'

withDefaults(defineProps<{
  carrierServices?: ShippingCarrierService[]
  carriers?: ShippingCarrier[]
  templates?: ShippingTemplate[]
  loading?: boolean
  canCreate?: boolean
  canEdit?: boolean
  canDelete?: boolean
}>(), {
  carrierServices: () => [],
  carriers: () => [],
  templates: () => [],
  loading: false,
  canCreate: false,
  canEdit: false,
  canDelete: false,
})

const emit = defineEmits<{
  (event: 'create'): void
  (event: 'edit', service: ShippingCarrierService): void
  (event: 'delete', service: ShippingCarrierService): void
}>()
</script>
