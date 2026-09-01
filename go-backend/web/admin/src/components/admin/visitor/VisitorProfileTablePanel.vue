<template>
  <AdminTablePanel class="h-full min-h-0" :loading="loading" scroll-body>
    <Table class="min-w-[1360px] table-fixed">
      <colgroup>
        <col class="w-16" />
        <col class="w-20" />
        <col class="w-32" />
        <col class="w-40" />
        <col class="w-36" />
        <col class="w-60" />
        <col class="w-40" />
        <col class="w-40" />
        <col class="w-40" />
        <col class="w-16" />
      </colgroup>
      <TableHeader>
        <TableRow>
          <TableHead>ID</TableHead>
          <TableHead>身份</TableHead>
          <TableHead>质量/状态</TableHead>
          <TableHead>联系信息</TableHead>
          <TableHead>地区/语言</TableHead>
          <TableHead>绑定事实</TableHead>
          <TableHead>指纹状态</TableHead>
          <TableHead>最后活跃</TableHead>
          <TableHead>创建时间</TableHead>
          <TableHead class="sticky right-0 z-20 border-l border-dashed border-border/60 bg-card text-right">详情</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="profiles.length === 0" :colspan="10">
          <div class="flex flex-col items-center text-muted-foreground">
            <Fingerprint class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无访客画像</span>
          </div>
        </TableEmpty>

        <TableRow
          v-for="profile in profiles"
          :key="profile.id"
        >
          <TableCell class="font-mono text-xs text-muted-foreground">#{{ profile.id }}</TableCell>
          <TableCell>
            <AdminStatusBadge :tone="profile.identity === 'account' ? 'green' : 'amber'">
              {{ profile.identity === 'account' ? '会员' : '匿名' }}
            </AdminStatusBadge>
            <span v-if="profile.user_id" class="mt-1 block font-mono text-[11px] text-muted-foreground">UID {{ profile.user_id }}</span>
          </TableCell>
          <TableCell>
            <div class="flex flex-wrap items-center gap-1.5">
              <AdminStatusBadge :tone="statusTone(profile.profile_status)">
                {{ statusLabel(profile.profile_status) }}
              </AdminStatusBadge>
              <span class="font-mono text-[11px] font-black text-foreground">Q{{ profile.profile_quality_score || 0 }}</span>
            </div>
            <p class="mt-1 truncate font-mono text-[11px] text-muted-foreground">
              {{ actionLabel(profile.last_meaningful_action) }}
            </p>
          </TableCell>
          <TableCell>
            <div class="min-w-0">
 <p class="truncate font-medium">{{ profile.email || '未采集邮箱'}}</p>
 <p class="mt-1 text-[11px] text-muted-foreground">来源：{{ profile.email_source || 'not_captured'}}</p>
            </div>
          </TableCell>
          <TableCell>
 <p class="truncate text-xs font-bold">{{ profile.region_label || '未采集地区'}}</p>
 <p class="mt-1 font-mono text-[11px] text-muted-foreground">{{ profile.locale || 'no-locale'}}</p>
          </TableCell>
          <TableCell>
            <div class="flex flex-wrap gap-1.5">
              <AdminStatusBadge :tone="profile.has_customer_service_visitor ? 'green' : 'gray'">Public Chat</AdminStatusBadge>
              <AdminStatusBadge :tone="profile.has_cart_session ? 'blue' : 'gray'">Cart</AdminStatusBadge>
              <AdminStatusBadge :tone="profile.has_email ? 'green' : 'gray'">Email</AdminStatusBadge>
            </div>
            <p class="mt-2 break-all font-mono text-[11px] text-muted-foreground">
              {{ profile.customer_service_visitor_hash_preview || 'no-chat-visitor' }}
            </p>
          </TableCell>
          <TableCell>
            <div class="flex flex-wrap gap-1.5">
              <AdminStatusBadge :tone="profile.has_ip_fingerprint ? 'green' : 'gray'">IP hash</AdminStatusBadge>
              <AdminStatusBadge :tone="profile.has_user_agent_fingerprint ? 'green' : 'gray'">UA hash</AdminStatusBadge>
              <AdminStatusBadge :tone="profile.ip_block_match ? 'coral' : 'gray'">
                {{ profile.ip_block_match ? 'IP blocked' : 'IP clear' }}
              </AdminStatusBadge>
            </div>
          </TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(profile.last_seen_at) }}</TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(profile.created_at) }}</TableCell>
          <TableCell class="sticky right-0 z-10 border-l border-dashed border-border/60 bg-card text-right">
            <Tooltip>
              <TooltipTrigger as-child>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  :aria-label="`查看画像 #${profile.id} 详情`"
                  @click="emit('preview-profile', profile)"
                >
                  <Eye class="size-3.5" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>查看画像详情</TooltipContent>
            </Tooltip>
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
import { Eye, Fingerprint } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import type { VisitorPagination, VisitorProfile } from '@/modules/visitor/visitorTypes'

withDefaults(defineProps<{
  loading?: boolean
  profiles?: VisitorProfile[]
  pagination: VisitorPagination
  formatDate: (value: unknown) => string
}>(), {
  loading: false,
  profiles: () => [],
})

const emit = defineEmits<{
  (event: 'preview-profile', profile: VisitorProfile): void
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()

const statusLabel = (status?: string): string => ({
  active: '有效',
  candidate: '候选',
  archived: '归档',
  suppressed: '抑制',
} as Record<string, string>)[status || ''] || '有效'

const statusTone = (status?: string): AdminStatusTone => ({
  active: 'green',
  candidate: 'amber',
  archived: 'gray',
  suppressed: 'coral',
} as Record<string, AdminStatusTone>)[status || ''] || 'green'

const actionLabel = (action?: string): string => ({
  cart_action: '购物车动作',
  customer_service: '客服会话',
  email_capture: '邮箱捕获',
  account: '账号绑定',
  identity_bind: '身份绑定',
} as Record<string, string>)[action || ''] || '无有效动作'
</script>

