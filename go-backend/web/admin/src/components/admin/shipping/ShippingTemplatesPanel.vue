<template>
  <TabsContent value="templates" class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-sm font-black tracking-tighter uppercase">运费模板</h2>
        <p class="mt-1 text-xs text-muted-foreground">维护基础计费方式、默认费用、免运门槛和区域规则矩阵。</p>
      </div>
      <Button v-if="canCreate" size="sm" @click="emit('create-template')">
        <Plus class="size-3.5" />
        新增运费模板
      </Button>
    </div>

    <AdminTablePanel :loading="loadingTemplates">
      <Table class="min-w-[1120px]">
        <TableHeader>
          <TableRow>
            <TableHead>模板名称</TableHead>
            <TableHead class="w-28">计费类型</TableHead>
            <TableHead class="w-28 text-right">默认运费</TableHead>
            <TableHead class="w-36 text-right">免运门槛</TableHead>
            <TableHead>规则摘要</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-32 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="templates.length === 0" :colspan="7">
            <div class="flex flex-col items-center text-muted-foreground">
              <Calculator class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无运费模板</span>
            </div>
          </TableEmpty>
          <TableRow v-for="template in templates" :key="template.id">
            <TableCell>
 <span class="block font-bold text-xs">{{ template.name || '-'}}</span>
 <span class="block max-w-96 truncate text-[10px] text-muted-foreground/70">{{ template.description || '暂无说明'}}</span>
            </TableCell>
            <TableCell>{{ templateTypeLabel(template.type) }}</TableCell>
            <TableCell class="text-right tabular-nums">{{ formatMoney(template.default_fee) }}</TableCell>
            <TableCell class="text-right tabular-nums">
              {{ template.free_shipping ? formatMoney(template.free_threshold) : '未开启' }}
            </TableCell>
            <TableCell class="max-w-[28rem] truncate text-xs text-muted-foreground">{{ formatRuleSummary(template.rules) }}</TableCell>
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
                  :aria-label="`编辑运费模板 ${template.name}`"
                  @click="emit('edit-template', template)"
                >
                  <Pencil class="size-4" />
                </Button>
                <Button
                  v-if="canDelete"
                  variant="ghost"
                  size="icon-sm"
                  class="text-destructive hover:text-destructive"
                  :aria-label="`删除运费模板 ${template.name}`"
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

    <div class="flex flex-wrap items-end justify-between gap-3 pt-3">
      <div>
        <h3 class="text-sm font-black tracking-tighter uppercase">承运商代码映射</h3>
        <p class="mt-1 text-xs text-muted-foreground">把本地承运商或线路服务映射到 Provider carrier code，后续追踪号注册和轨迹同步统一读取这里。</p>
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
  formatRuleSummary,
  templateTypeLabel,
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
} from './shippingTypes'

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
