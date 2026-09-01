<template>
  <TabsContent value="zones" class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-sm font-black tracking-tighter uppercase">配送区域</h2>
        <p class="mt-1 text-xs text-muted-foreground">按国家/地区组织区域，供运费模板匹配使用。</p>
      </div>
      <Button v-if="canCreate" size="sm" @click="emit('create')">
        <Plus class="size-3.5" />
        新增配送区域
      </Button>
    </div>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[640px]">
        <TableHeader>
          <TableRow>
            <TableHead>区域名称</TableHead>
            <TableHead>国家/地区</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-32 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="zones.length === 0" :colspan="4">
            <div class="flex flex-col items-center text-muted-foreground">
              <MapPin class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无配送区域</span>
            </div>
          </TableEmpty>
          <TableRow v-for="zone in zones" :key="zone.id">
 <TableCell class="font-bold text-xs">{{ zone.name || '-'}}</TableCell>
            <TableCell class="max-w-96 truncate text-[10px] text-muted-foreground/70">{{ addressRegionSummary(zone.countries) }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="zone.enabled ? 'green' : 'gray'">
                {{ zone.enabled ? '启用' : '停用' }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-1">
                <Button
                  v-if="canEdit"
                  variant="ghost"
                  size="icon-sm"
                  :aria-label="`编辑配送区域 ${zone.name}`"
                  @click="emit('edit', zone)"
                >
                  <Pencil class="size-4" />
                </Button>
                <Button
                  v-if="canDelete"
                  variant="ghost"
                  size="icon-sm"
                  class="text-destructive hover:text-destructive"
                  :aria-label="`删除配送区域 ${zone.name}`"
                  @click="emit('delete', zone)"
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
import { MapPin, Pencil, Plus, Trash2 } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { TabsContent } from '@/components/ui/tabs'
import { addressRegionSummary } from '@/lib/addressRegions'
import type { ShippingZone } from '@/modules/shipping/shippingTypes'

withDefaults(defineProps<{
  zones?: ShippingZone[]
  loading?: boolean
  canCreate?: boolean
  canEdit?: boolean
  canDelete?: boolean
}>(), {
  zones: () => [],
  loading: false,
  canCreate: false,
  canEdit: false,
  canDelete: false
})

const emit = defineEmits<{
  (event: 'create'): void
  (event: 'edit', zone: ShippingZone): void
  (event: 'delete', zone: ShippingZone): void
}>()
</script>

