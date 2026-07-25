<template>
  <AdminTablePanel :loading="loading">
    <Table class="min-w-[1280px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-20">ID</TableHead>
          <TableHead class="w-24">身份</TableHead>
          <TableHead>联系信息</TableHead>
          <TableHead class="w-56">地区/语言</TableHead>
          <TableHead class="w-72">绑定事实</TableHead>
          <TableHead class="w-44">指纹状态</TableHead>
          <TableHead class="w-44">最后活跃</TableHead>
          <TableHead class="w-44">创建时间</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="profiles.length === 0" :colspan="8">
          <div class="flex flex-col items-center text-muted-foreground">
            <Fingerprint class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无访客画像</span>
          </div>
        </TableEmpty>

        <TableRow
          v-for="profile in profiles"
          :key="profile.id"
          class="cursor-pointer"
          :class="selectedProfile?.id === profile.id ? 'bg-primary/5' : ''"
          @click="emit('select-profile', profile)"
        >
          <TableCell class="font-mono text-xs text-muted-foreground">#{{ profile.id }}</TableCell>
          <TableCell>
            <AdminStatusBadge :tone="profile.identity === 'account' ? 'green' : 'amber'">
              {{ profile.identity === 'account' ? '会员' : '匿名' }}
            </AdminStatusBadge>
            <span v-if="profile.user_id" class="mt-1 block font-mono text-[11px] text-muted-foreground">UID {{ profile.user_id }}</span>
          </TableCell>
          <TableCell>
            <div class="min-w-0">
              <p class="truncate font-medium">{{ profile.email || '未采集邮箱' }}</p>
              <p class="mt-1 text-[11px] text-muted-foreground">来源：{{ profile.email_source || 'not_captured' }}</p>
            </div>
          </TableCell>
          <TableCell>
            <p class="truncate text-xs font-bold">{{ profile.region_label || '未采集地区' }}</p>
            <p class="mt-1 font-mono text-[11px] text-muted-foreground">{{ profile.locale || 'no-locale' }}</p>
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
            </div>
          </TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(profile.last_seen_at) }}</TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(profile.created_at) }}</TableCell>
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

<script setup>
import { Fingerprint } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'

defineProps({
  loading: { type: Boolean, default: false },
  profiles: { type: Array, default: () => [] },
  selectedProfile: { type: Object, default: null },
  pagination: { type: Object, required: true },
  formatDate: { type: Function, required: true },
})

const emit = defineEmits(['select-profile', 'update-page', 'update-page-size'])
</script>
