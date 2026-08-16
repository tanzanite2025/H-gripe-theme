<template>
  <Card size="sm">
    <CardHeader class="flex flex-col gap-3 border-b border-dashed border-border/70 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <CardTitle>现有后台账号</CardTitle>
        <CardDescription>只显示 admin、manager、editor 和 support 账号。</CardDescription>
      </div>
      <div class="flex w-full gap-2 sm:w-auto">
        <Input
          :model-value="search"
          class="sm:w-64"
          placeholder="搜索邮箱或用户名"
          @update:model-value="emit('update:search', String($event))"
          @keyup.enter="emit('refresh')"
        />
        <Button variant="outline" size="icon" aria-label="搜索账号" title="搜索账号" @click="emit('refresh')">
          <Search class="size-4" />
        </Button>
      </div>
    </CardHeader>
    <CardContent class="pt-1">
      <div v-if="loading" class="flex items-center justify-center gap-2 py-8 text-xs text-muted-foreground">
        <LoaderCircle class="size-4 animate-spin" />
        正在读取账号
      </div>
      <div v-else-if="accounts.length === 0" class="py-8 text-center text-xs text-muted-foreground">
        当前环境没有后台账号，请先使用发布前 CLI 创建首个账号。
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full min-w-[680px] text-left text-xs">
          <thead class="border-b border-dashed border-border/70 text-[10px] uppercase tracking-widest text-muted-foreground/70">
            <tr>
              <th class="px-3 py-3 font-black">账号</th>
              <th class="px-3 py-3 font-black">角色</th>
              <th class="px-3 py-3 font-black">状态</th>
              <th class="px-3 py-3 font-black">更新时间</th>
              <th class="px-3 py-3 text-right font-black">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="account in accounts" :key="account.id" class="border-b border-dashed border-border/60 last:border-0">
              <td class="px-3 py-3">
                <p class="font-black">{{ account.email }}</p>
                <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ account.username }}</p>
              </td>
              <td class="px-3 py-3">{{ roleLabel(account.role) }}</td>
              <td class="px-3 py-3">
                <span class="rounded-full px-2 py-1 text-[10px] font-black" :class="statusClass(account.status)">
                  {{ statusLabel(account.status) }}
                </span>
              </td>
              <td class="px-3 py-3 text-muted-foreground">{{ formatDate(account.updated_at) }}</td>
              <td class="px-3 py-3 text-right">
                <Button variant="ghost" size="sm" @click="emit('select-account', account)">
                  <Settings2 class="size-4" />
                  载入
                </Button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { LoaderCircle, Search, Settings2 } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import type { OpsAdminAccount } from '@/api/ops'

defineProps<{
  accounts: OpsAdminAccount[]
  loading: boolean
  search: string
}>()

const emit = defineEmits<{
  (event: 'update:search', value: string): void
  (event: 'refresh'): void
  (event: 'select-account', account: OpsAdminAccount): void
}>()

const roleLabel = (role: string): string => ({
  admin: '超级管理员',
  manager: '经理',
  editor: '编辑',
  support: '客服',
}[role] || role)

const statusLabel = (status: string): string => ({
  active: '活跃',
  inactive: '未激活',
  suspended: '已停用',
}[status] || status)

const statusClass = (status: string): string => ({
  active: 'bg-emerald-500/10 text-emerald-700',
  inactive: 'bg-muted text-muted-foreground',
  suspended: 'bg-destructive/10 text-destructive',
}[status] || 'bg-muted text-muted-foreground')

const formatDate = (value?: string): string => value ? new Date(value).toLocaleString('zh-CN') : '-'
</script>
