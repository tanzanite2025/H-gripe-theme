<template>
  <TabsContent value="templates" class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-sm font-black tracking-tighter uppercase">线路计费模板</h2>
        <p class="mt-1 text-xs text-muted-foreground">独立定义向客户收取的运费、首续重、附加费和免邮条件。产品及 SKU 可以绑定多个计费模板，Checkout 返回可选运输方案。</p>
      </div>
      <Button v-if="canCreate" size="sm" @click="emit('create-template')">
        <Plus class="size-3.5" />
        新增线路计费模板
      </Button>
    </div>

    <AdminTablePanel :loading="loadingTemplates">
      <Table class="min-w-[1120px]">
        <TableHeader>
          <TableRow>
            <TableHead>模板名称</TableHead>
            <TableHead class="w-28">模板类型</TableHead>
            <TableHead class="w-24">源币种</TableHead>
            <TableHead class="w-28 text-right">默认运费</TableHead>
            <TableHead class="w-36 text-right">免运门槛</TableHead>
            <TableHead>区域/规则摘要</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-32 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="templates.length === 0" :colspan="8">
            <div class="flex flex-col items-center text-muted-foreground">
              <Calculator class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无线路计费模板</span>
            </div>
          </TableEmpty>
          <TableRow v-for="template in templates" :key="template.id">
            <TableCell>
 <span class="block font-bold text-xs">{{ template.name || '-'}}</span>
 <span class="block max-w-96 truncate text-[10px] text-muted-foreground/70">{{ template.description || '暂无说明'}}</span>
            </TableCell>
            <TableCell>
              <span class="block">{{ templateKindLabel(template.template_kind, template.type) }}</span>
              <span v-if="template.template_kind !== 'system_free_shipping' && template.type !== 'free_shipping'" class="text-[10px] text-muted-foreground">{{ templateTypeLabel(template.type) }}</span>
            </TableCell>
            <TableCell class="font-mono font-bold">{{ template.currency || '-' }}</TableCell>
            <TableCell class="text-right tabular-nums">{{ formatMoney(template.default_fee_minor, template.currency) }}</TableCell>
            <TableCell class="text-right tabular-nums">
              {{ template.free_shipping ? formatMoney(template.free_threshold_minor, template.currency) : '未开启' }}
            </TableCell>
            <TableCell class="max-w-[28rem] truncate text-xs text-muted-foreground">
              <template v-if="template.template_kind === 'system_free_shipping' || template.type === 'free_shipping'">
                {{ freeShippingCountriesLabel(template.free_shipping_countries) }}
              </template>
              <template v-else>{{ formatRuleSummary(template.rules, template.type, template.currency) }}</template>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="template.enabled ? 'green' : 'gray'">
                {{ template.enabled ? '启用' : '停用' }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-1">
                <Button
                  v-if="canEdit"
                  variant="ghost"
                  size="icon-sm"
                  :aria-label="`编辑线路计费模板 ${template.name}`"
                  @click="emit('edit-template', template)"
                >
                  <Pencil class="size-4" />
                </Button>
                <Button
                  v-if="canDelete && !template.is_system_managed && template.template_kind !== 'system_free_shipping' && template.type !== 'free_shipping'"
                  variant="ghost"
                  size="icon-sm"
                  class="text-destructive hover:text-destructive"
                  :aria-label="`删除线路计费模板 ${template.name}`"
                  @click="emit('delete-template', template)"
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
import { Calculator, Pencil, Plus, Radar, Trash2 } from '@lucide/vue'
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
