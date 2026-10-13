<template>
 <div class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3 pt-3">
      <div>
        <h3 class="text-sm font-black tracking-tighter uppercase">承运商代码映射</h3>
        <p class="mt-1 text-xs text-muted-foreground">把本地承运商或承运商线路映射到 Provider carrier code，后续追踪号注册和轨迹同步统一读取这里。</p>
      </div>
      <Button
        v-if="canCreate"
        size="sm"
        :disabled="trackingProviders.length === 0 || (carriers.length === 0 && carrierServices.length === 0)"
        @click="emit('create-mapping')"
      >
        <Plus class="size-3.5" />
        新增承运商映射
      </Button>
    </div>

    <AdminTablePanel :loading="loadingTrackingMappings">
      <Table class="min-w-[1120px]">
        <TableHeader>
          <TableRow>
            <TableHead>追踪 Provider</TableHead>
            <TableHead>本地对象</TableHead>
            <TableHead>Provider Carrier</TableHead>
            <TableHead class="w-24 text-right">优先级</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-32 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="trackingCarrierMappings.length === 0" :colspan="6">
            <div class="flex flex-col items-center text-muted-foreground">
              <Radar class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无承运商代码映射</span>
            </div>
          </TableEmpty>
          <TableRow v-for="mapping in trackingCarrierMappings" :key="mapping.id">
            <TableCell>
              <span class="block font-bold text-xs">{{ trackingProviderName(mapping, trackingProviders) }}</span>
              <span class="block font-mono text-[10px] text-muted-foreground/70">
                provider_id={{ mapping.provider_id || '-' }}
              </span>
            </TableCell>
            <TableCell>
              <span class="block font-bold text-xs">{{ trackingMappingLocalTargetLabel(mapping, carriers, carrierServices) }}</span>
              <span class="block text-[10px] text-muted-foreground/70">{{ trackingMappingScopeLabel(mapping.scope) }}</span>
            </TableCell>
            <TableCell>
 <span class="block font-mono text-xs font-bold">{{ mapping.provider_carrier_code || '-'}}</span>
              <span class="block max-w-80 truncate text-[10px] text-muted-foreground/70">
                {{ mapping.provider_carrier_name || mapping.description || '暂无 Provider 名称' }}
              </span>
            </TableCell>
            <TableCell class="text-right tabular-nums">{{ mapping.priority || 0 }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="mapping.enabled ? 'green' : 'gray'">
                {{ mapping.enabled ? '启用' : '停用' }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-1">
                <Button
                  v-if="canEdit"
                  variant="ghost"
                  size="icon-sm"
                  :aria-label="`编辑承运商映射 ${mapping.provider_carrier_code}`"
                  @click="emit('edit-mapping', mapping)"
                >
                  <Pencil class="size-4" />
                </Button>
                <Button
                  v-if="canDelete"
                  variant="ghost"
                  size="icon-sm"
                  class="text-destructive hover:text-destructive"
                  :aria-label="`删除承运商映射 ${mapping.provider_carrier_code}`"
                  @click="emit('delete-mapping', mapping)"
                >
                  <Trash2 class="size-4" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>
 </div>
</template>
<script setup lang="ts">
import { Pencil, Plus, Radar, Trash2 } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { TabsContent } from '@/components/ui/tabs'
import {
  formatMoney,
  freeShippingCountriesLabel,
  formatRuleSummary,
  templateTypeLabel,
  templateKindLabel,
  trackingMappingLocalTargetLabel,
  trackingMappingScopeLabel,
  trackingProviderName,
} from '@/lib/shippingPresentation'
import type {
  ShippingCarrier,
  ShippingCarrierService,
  ShippingTemplate,
  TrackingCarrierMapping,
  TrackingProvider
} from '@/modules/shipping/shippingTypes'

withDefaults(defineProps<{
  templates?: ShippingTemplate[]
  trackingCarrierMappings?: TrackingCarrierMapping[]
  trackingProviders?: TrackingProvider[]
  carriers?: ShippingCarrier[]
  carrierServices?: ShippingCarrierService[]
  loadingTemplates?: boolean
  loadingTrackingMappings?: boolean
  canCreate?: boolean
  canEdit?: boolean
  canDelete?: boolean
}>(), {
  templates: () => [],
  trackingCarrierMappings: () => [],
  trackingProviders: () => [],
  carriers: () => [],
  carrierServices: () => [],
  loadingTemplates: false,
  loadingTrackingMappings: false,
  canCreate: false,
  canEdit: false,
  canDelete: false
})

const emit = defineEmits<{
  (event: 'create-template'): void
  (event: 'edit-template', template: ShippingTemplate): void
  (event: 'delete-template', template: ShippingTemplate): void
  (event: 'create-mapping'): void
  (event: 'edit-mapping', mapping: TrackingCarrierMapping): void
  (event: 'delete-mapping', mapping: TrackingCarrierMapping): void
}>()
</script>
