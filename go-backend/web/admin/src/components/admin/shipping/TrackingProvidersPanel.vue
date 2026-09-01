<template>
  <TabsContent value="tracking" class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-sm font-black tracking-tighter uppercase">追踪配置</h2>
        <p class="mt-1 text-xs text-muted-foreground">统一维护 17TRACK / AfterShip 等追踪 Provider 的 API 凭证、Webhook 和同步策略。</p>
      </div>
      <Button v-if="canCreate" size="sm" @click="emit('create')">
        <Plus class="size-3.5" />
        新增追踪配置
      </Button>
    </div>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[1120px]">
        <TableHeader>
          <TableRow>
            <TableHead>Provider</TableHead>
            <TableHead>接口地址</TableHead>
            <TableHead class="w-44">同步策略</TableHead>
            <TableHead class="w-44">凭证状态</TableHead>
            <TableHead class="w-24 text-right">排序</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-32 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="trackingProviders.length === 0" :colspan="7">
            <div class="flex flex-col items-center text-muted-foreground">
              <Radar class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无追踪 Provider 配置</span>
            </div>
          </TableEmpty>
          <TableRow v-for="provider in trackingProviders" :key="provider.id">
            <TableCell>
 <span class="block font-bold text-xs">{{ provider.provider_name || '-'}}</span>
              <span class="block font-mono text-[10px] text-muted-foreground/70">
                {{ provider.provider_code || '-' }} · {{ trackingEnvironmentLabel(provider.environment) }}
              </span>
            </TableCell>
            <TableCell>
 <span class="block max-w-96 truncate font-mono text-xs">{{ provider.base_url || '未配置 Base URL'}}</span>
              <div class="mt-1 flex max-w-96 items-center gap-1 rounded-md bg-muted/35 px-2 py-1">
                <span
                  class="min-w-0 flex-1 truncate font-mono text-[10px] text-muted-foreground"
                  :title="trackingWebhookUrl(provider)"
                >
                  {{ trackingWebhookUrl(provider) || '填写 Provider 代码后生成 Webhook URL' }}
                </span>
                <Button
                  v-if="trackingWebhookUrl(provider)"
                  variant="ghost"
                  size="icon-sm"
                  class="size-6 shrink-0"
                  :aria-label="`复制 ${provider.provider_name || provider.provider_code} Webhook 地址`"
                  @click="emit('copy-webhook', provider)"
                >
                  <Copy class="size-3.5" />
                </Button>
              </div>
 <span class="mt-1 block max-w-96 truncate text-[10px] text-muted-foreground/70">{{ provider.description || '暂无说明'}}</span>
            </TableCell>
            <TableCell class="text-xs text-muted-foreground">{{ formatTrackingSyncPolicy(provider) }}</TableCell>
            <TableCell>
              <div class="flex flex-wrap gap-1.5">
                <AdminStatusBadge :tone="trackingProviderHasApiKey(provider) ? 'green' : 'gray'">
                  API {{ trackingProviderHasApiKey(provider) ? '已配置' : '未配置' }}
                </AdminStatusBadge>
                <AdminStatusBadge :tone="trackingProviderHasWebhookSecret(provider) ? 'green' : 'gray'">
                  WEBHOOK {{ trackingProviderHasWebhookSecret(provider) ? '已配置' : '未配置' }}
                </AdminStatusBadge>
              </div>
            </TableCell>
            <TableCell class="text-right tabular-nums">{{ provider.sort_order || 0 }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="provider.enabled ? 'green' : 'gray'">
                {{ provider.enabled ? '启用' : '停用' }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-1">
                <Button
                  v-if="canEdit"
                  variant="ghost"
                  size="icon-sm"
                  :aria-label="`编辑追踪配置 ${provider.provider_name}`"
                  @click="emit('edit', provider)"
                >
                  <Pencil class="size-4" />
                </Button>
                <Button
                  v-if="canDelete"
                  variant="ghost"
                  size="icon-sm"
                  class="text-destructive hover:text-destructive"
                  :aria-label="`删除追踪配置 ${provider.provider_name}`"
                  @click="emit('delete', provider)"
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
import { Copy, Pencil, Plus, Radar, Trash2 } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { TabsContent } from '@/components/ui/tabs'
import {
  formatTrackingSyncPolicy,
  trackingEnvironmentLabel,
  trackingProviderHasApiKey,
  trackingProviderHasWebhookSecret,
  trackingWebhookUrl,
} from '@/lib/shippingPresentation'
import type { TrackingProvider } from '@/modules/shipping/shippingTypes'

withDefaults(defineProps<{
  trackingProviders?: TrackingProvider[]
  loading?: boolean
  canCreate?: boolean
  canEdit?: boolean
  canDelete?: boolean
}>(), {
  trackingProviders: () => [],
  loading: false,
  canCreate: false,
  canEdit: false,
  canDelete: false
})

const emit = defineEmits<{
  (event: 'create'): void
  (event: 'edit', provider: TrackingProvider): void
  (event: 'delete', provider: TrackingProvider): void
  (event: 'copy-webhook', provider: TrackingProvider): void
}>()
</script>

