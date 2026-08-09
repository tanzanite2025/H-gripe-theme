<template>
  <AdminTablePanel :loading="loading">
    <Table class="min-w-[780px]">
      <TableHeader>
        <TableRow>
          <TableHead>等级名称</TableHead>
          <TableHead class="w-52">积分范围</TableHead>
          <TableHead class="w-28 text-right">折扣率</TableHead>
          <TableHead>权益说明</TableHead>
          <TableHead class="w-20 text-right">排序</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="levels.length === 0" :colspan="6">
          <div class="flex flex-col items-center text-muted-foreground">
            <Crown class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无会员等级</span>
          </div>
        </TableEmpty>
        <TableRow v-for="level in levels" :key="level.id || level.name">
          <TableCell>
            <div class="flex items-center gap-2">
              <span class="size-3 rounded-full border" :style="{ backgroundColor: level.color || '#94a3b8' }" />
              <span class="font-bold text-xs">{{ memberLevelLabel(level.name) }}</span>
            </div>
          </TableCell>
          <TableCell class="tabular-nums">{{ level.min_points }} - {{ level.max_points }}</TableCell>
          <TableCell class="text-right tabular-nums">{{ formatRate(level.discount_rate) }}</TableCell>
          <TableCell class="max-w-72 truncate text-muted-foreground">{{ level.benefits || '-' }}</TableCell>
          <TableCell class="text-right tabular-nums">{{ level.sort_order || 0 }}</TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" :aria-label="`管理会员等级 ${memberLevelLabel(level.name)}`">
                  <MoreHorizontal class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-36">
                <DropdownMenuItem
                  v-if="canEdit"
                  :disabled="!level.id"
                  @select="level.id && emit('edit', level)"
                >
                  <Pencil class="size-4" />
                  {{ level.id ? '编辑' : '待初始化' }}
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="canDelete" />
                <DropdownMenuItem
                  v-if="canDelete"
                  class="text-destructive focus:text-destructive"
                  @select="emit('delete', level)"
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
  </AdminTablePanel>
</template>

<script setup lang="ts">
import { Crown, MoreHorizontal, Pencil, Trash2 } from '@lucide/vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'

interface MemberLevel {
  id?: string | number | null
  name?: string
  color?: string
  min_points?: number | string
  max_points?: number | string
  discount_rate?: number | string
  benefits?: string
  sort_order?: number | string
}

const props = withDefaults(defineProps<{
  loading?: boolean
  levels?: MemberLevel[]
  canEdit?: boolean
  canDelete?: boolean
  formatRate: (value: unknown) => string
}>(), {
  loading: false,
  levels: () => [],
  canEdit: false,
  canDelete: false,
})

const emit = defineEmits<{
  (event: 'edit', level: MemberLevel): void
  (event: 'delete', level: MemberLevel): void
}>()

const memberLevelLabel = (name: unknown): string => {
  const key = String(name || '').trim().toLowerCase()
  const labels: Record<string, string> = {
    ordinary: '普通',
    bronze: '铜牌',
    silver: '银牌',
    gold: '金牌',
    platinum: '铂金',
    diamond: '钻石',
  }
  return labels[key] || String(name || '-')
}
</script>
