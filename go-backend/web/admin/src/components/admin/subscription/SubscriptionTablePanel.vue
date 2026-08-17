<template>
  <AdminTablePanel :loading="loading" :batch-visible="selectedSubscriptions.length > 0">
    <template #batch>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <span class="text-xs font-medium">已选择 {{ selectedSubscriptions.length }} 个订阅</span>
        <Button
          v-if="canDelete"
          variant="destructive"
          size="sm"
          @click="emit('request-batch-delete')"
        >
          <Trash2 class="size-3.5" />
          批量删除
        </Button>
      </div>
    </template>

    <Table class="min-w-[1120px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-11">
            <Checkbox
              :model-value="selectionState"
              aria-label="选择当前页订阅"
              @update:model-value="emit('toggle-all', $event)"
            />
          </TableHead>
          <TableHead class="w-16">ID</TableHead>
          <TableHead>邮箱</TableHead>
          <TableHead class="w-24">状态</TableHead>
          <TableHead class="w-24">语言</TableHead>
          <TableHead class="w-28">来源</TableHead>
          <TableHead class="w-44">标签</TableHead>
          <TableHead class="w-44">订阅时间</TableHead>
          <TableHead class="w-44">退订时间</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="subscriptions.length === 0" :colspan="10">
          <div class="flex flex-col items-center text-muted-foreground">
            <MailOpen class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无订阅</span>
          </div>
        </TableEmpty>

        <TableRow v-for="subscription in subscriptions" :key="subscription.id || subscription.email">
          <TableCell>
            <Checkbox
              :model-value="isSelected(subscription.email)"
              :aria-label="`选择订阅 ${subscription.email}`"
              @update:model-value="emit('toggle-subscription', subscription, $event)"
            />
          </TableCell>
 <TableCell class="font-mono text-xs text-muted-foreground">{{ subscription.id || '-'}}</TableCell>
          <TableCell>
            <a :href="`mailto:${subscription.email}`" class="font-medium hover:text-primary hover:underline">
              {{ subscription.email }}
            </a>
          </TableCell>
          <TableCell>
            <AdminStatusBadge :tone="statusTone(subscription.status)">{{ statusName(subscription.status) }}</AdminStatusBadge>
          </TableCell>
          <TableCell>{{ localeName(subscription.locale) }}</TableCell>
          <TableCell>{{ sourceName(subscription.source) }}</TableCell>
 <TableCell class="max-w-44 truncate text-xs text-muted-foreground">{{ subscription.tags || '-'}}</TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(subscription.subscribed_at) }}</TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(subscription.unsubscribed_at) }}</TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" :aria-label="`管理订阅 ${subscription.email}`">
                  <MoreHorizontal class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-40">
                <DropdownMenuItem
                  v-if="canEdit"
                  @select="emit('request-toggle-status', subscription)"
                >
                  <MailCheck v-if="subscription.status !== 'active'" class="size-4" />
                  <MailX v-else class="size-4" />
                  {{ subscription.status === 'active' ? '标记为退订' : '恢复订阅' }}
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="canDelete" />
                <DropdownMenuItem
                  v-if="canDelete"
                  class="text-destructive focus:text-destructive"
                  @select="emit('request-delete', subscription)"
                >
                  <Trash2 class="size-4" />
                  删除
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <template #footer>
      <AdminPagination
        :page="pagination.page"
        :page-size="pagination.pageSize"
        :total="pagination.total"
        @update:page="emit('update-page', $event)"
        @update:page-size="emit('update-page-size', $event)"
      />
    </template>
  </AdminTablePanel>
</template>

<script setup lang="ts">
import { MailCheck, MailOpen, MailX, MoreHorizontal, Trash2 } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type {
  SubscriptionBooleanResolver,
  SubscriptionDateFormatter,
  SubscriptionLabelResolver,
  SubscriptionPagination,
  SubscriptionRecord,
  SubscriptionSelectionState,
  SubscriptionToneResolver
} from './subscriptionTypes'

withDefaults(defineProps<{
  loading?: boolean
  subscriptions?: SubscriptionRecord[]
  selectedSubscriptions?: SubscriptionRecord[]
  selectionState?: SubscriptionSelectionState
  pagination: SubscriptionPagination
  canEdit?: boolean
  canDelete?: boolean
  isSelected: SubscriptionBooleanResolver
  statusName: SubscriptionLabelResolver
  statusTone: SubscriptionToneResolver
  localeName: SubscriptionLabelResolver
  sourceName: SubscriptionLabelResolver
  formatDate: SubscriptionDateFormatter
}>(), {
  loading: false,
  subscriptions: () => [],
  selectedSubscriptions: () => [],
  selectionState: false,
  canEdit: false,
  canDelete: false
})

const emit = defineEmits<{
  (event: 'toggle-all', checked: SubscriptionSelectionState): void
  (event: 'toggle-subscription', subscription: SubscriptionRecord, checked: SubscriptionSelectionState): void
  (event: 'request-toggle-status', subscription: SubscriptionRecord): void
  (event: 'request-delete', subscription: SubscriptionRecord): void
  (event: 'request-batch-delete'): void
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()
</script>
