<template>
  <AdminTablePanel :loading="loading" :batch-visible="selectedUsers.length > 0">
    <template #batch>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <span class="text-xs font-bold text-muted-foreground/80">已选择 {{ selectedUsers.length }} 个用户</span>
        <Button v-if="canDelete" variant="destructive" size="sm" @click="emit('batch-delete')">
          <Trash2 class="size-3.5" />
          批量删除
        </Button>
      </div>
    </template>

    <Table class="min-w-[980px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-11">
            <Checkbox
              :model-value="selectionState"
              aria-label="选择当前页用户"
              @update:model-value="emit('toggle-all-users', $event)"
            />
          </TableHead>
          <TableHead class="w-20">ID</TableHead>
          <TableHead>用户名</TableHead>
          <TableHead>邮箱</TableHead>
          <TableHead>姓名</TableHead>
          <TableHead class="w-28">角色</TableHead>
          <TableHead class="w-24">状态</TableHead>
          <TableHead class="w-44">创建时间</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="users.length === 0" :colspan="9">
          <div class="flex flex-col items-center text-muted-foreground">
            <UsersRound class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无用户</span>
          </div>
        </TableEmpty>

        <TableRow v-for="user in users" :key="user.id">
          <TableCell>
            <Checkbox
              :model-value="isUserSelected(user.id)"
              :disabled="user.id === currentUserId"
              :aria-label="`选择用户 ${user.username}`"
              @update:model-value="emit('toggle-user', user, $event)"
            />
          </TableCell>
          <TableCell class="font-mono text-xs text-muted-foreground">{{ user.id }}</TableCell>
          <TableCell class="font-bold text-xs">{{ user.username }}</TableCell>
          <TableCell class="max-w-64 truncate text-muted-foreground">{{ user.email }}</TableCell>
          <TableCell>{{ formatFullName(user) }}</TableCell>
          <TableCell>
            <AdminStatusBadge :tone="roleTone(user.role)">{{ getRoleName(user.role) }}</AdminStatusBadge>
          </TableCell>
          <TableCell>
            <AdminStatusBadge :tone="statusTone(user.status)">{{ getStatusName(user.status) }}</AdminStatusBadge>
          </TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(user.created_at) }}</TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" :aria-label="`管理用户 ${user.username}`">
                  <MoreHorizontal class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-40">
                <DropdownMenuItem v-if="canEdit" @select="emit('edit', user)">
                  <Pencil class="size-4" />
                  编辑
                </DropdownMenuItem>
                <DropdownMenuItem
                  v-if="canEdit && user.id !== currentUserId"
                  @select="emit('toggle-status', user)"
                >
                  <UserRoundCheck v-if="user.status !== 'active'" class="size-4" />
                  <UserRoundX v-else class="size-4" />
                  {{ user.status === 'active' ? '停用' : '启用' }}
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="canDelete && user.id !== currentUserId" />
                <DropdownMenuItem
                  v-if="canDelete && user.id !== currentUserId"
                  class="text-destructive focus:text-destructive"
                  @select="emit('delete', user)"
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
import {
  MoreHorizontal,
  Pencil,
  Trash2,
  UserRoundCheck,
  UserRoundX,
  UsersRound,
} from '@lucide/vue'
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
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type {
  UserDateFormatter,
  UserId,
  UserLabelResolver,
  UserNameFormatter,
  UserPagination,
  UserRecord,
  UserSelectionState,
  UserToneResolver
} from '@/modules/user/userTypes'

const props = withDefaults(defineProps<{
  loading?: boolean
  users?: UserRecord[]
  selectedUsers?: UserRecord[]
  pagination: UserPagination
  selectionState?: UserSelectionState
  currentUserId?: UserId | null
  canEdit?: boolean
  canDelete?: boolean
  getRoleName: UserLabelResolver
  roleTone: UserToneResolver
  getStatusName: UserLabelResolver
  statusTone: UserToneResolver
  formatDate: UserDateFormatter
  formatFullName: UserNameFormatter
}>(), {
  loading: false,
  users: () => [],
  selectedUsers: () => [],
  selectionState: false,
  currentUserId: null,
  canEdit: false,
  canDelete: false
})

const emit = defineEmits<{
  (event: 'batch-delete'): void
  (event: 'toggle-all-users', checked: UserSelectionState): void
  (event: 'toggle-user', user: UserRecord, checked: UserSelectionState): void
  (event: 'edit', user: UserRecord): void
  (event: 'toggle-status', user: UserRecord): void
  (event: 'delete', user: UserRecord): void
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()

const isUserSelected = (userId: UserId): boolean => props.selectedUsers.some((user) => user.id === userId)
</script>

