<template>
  <div class="space-y-4">
    <AdminPageHeader title="订阅管理" description="查看邮件订阅来源、状态和退订记录">
      <template #actions>
        <Button v-if="hasPermission('subscription:export')" variant="outline" @click="exportEmails">
          <Download class="size-4" />
          导出活跃邮箱
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <AdminFilterPanel>
      <div class="flex flex-col gap-3 sm:flex-row sm:items-end">
        <label class="w-full space-y-1.5 sm:w-52">
          <span class="text-xs font-medium text-muted-foreground">状态</span>
          <Select v-model="filters.status" @update:model-value="applyFilters">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="active">活跃</SelectItem>
              <SelectItem value="unsubscribed">已退订</SelectItem>
            </SelectContent>
          </Select>
        </label>
        <Button type="button" variant="outline" class="h-9" @click="resetFilters">
          <RotateCcw class="size-4" />
          重置
        </Button>
        <Button type="button" variant="ghost" class="h-9" @click="refreshSubscriptions">
          <RefreshCw class="size-4" />
          刷新
        </Button>
      </div>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading" :batch-visible="selectedSubscriptions.length > 0">
      <template #batch>
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span class="text-xs font-medium">已选择 {{ selectedSubscriptions.length }} 个订阅</span>
          <Button
            v-if="hasPermission('subscription:delete')"
            variant="destructive"
            size="sm"
            @click="requestBatchDelete"
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
                @update:model-value="toggleAllSubscriptions"
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
                @update:model-value="toggleSubscription(subscription, $event)"
              />
            </TableCell>
            <TableCell class="font-mono text-xs text-muted-foreground">{{ subscription.id || '-' }}</TableCell>
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
            <TableCell class="max-w-44 truncate text-xs text-muted-foreground">{{ subscription.tags || '-' }}</TableCell>
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
                    v-if="hasPermission('subscription:edit')"
                    @select="requestToggleStatus(subscription)"
                  >
                    <MailCheck v-if="subscription.status !== 'active'" class="size-4" />
                    <MailX v-else class="size-4" />
                    {{ subscription.status === 'active' ? '标记为退订' : '恢复订阅' }}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="hasPermission('subscription:delete')" />
                  <DropdownMenuItem
                    v-if="hasPermission('subscription:delete')"
                    class="text-destructive focus:text-destructive"
                    @select="requestDelete(subscription)"
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
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </template>
    </AdminTablePanel>

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      :destructive="confirmation.destructive"
      @confirm="executeConfirmedAction"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  CalendarPlus,
  Download,
  Mail,
  MailCheck,
  MailOpen,
  MailX,
  MoreHorizontal,
  RefreshCw,
  RotateCcw,
  Trash2,
  UserMinus
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()
const loading = ref(false)
const subscriptions = ref([])
const selectedSubscriptions = ref([])
const stats = ref({})
const filters = reactive({ status: 'all' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const confirmation = reactive({
  open: false,
  type: '',
  target: null,
  status: '',
  title: '',
  description: '',
  confirmLabel: '确定',
  destructive: false
})

const statItems = computed(() => [
  {
    key: 'total',
    label: '总订阅数',
    value: stats.value.total_count ?? stats.value.total ?? 0,
    icon: Mail,
    tone: 'gray'
  },
  {
    key: 'active',
    label: '活跃订阅',
    value: stats.value.active_count ?? stats.value.active ?? 0,
    icon: MailCheck,
    tone: 'green'
  },
  {
    key: 'unsubscribed',
    label: '已退订',
    value: stats.value.unsubscribed_count ?? stats.value.cancelled ?? 0,
    icon: UserMinus,
    tone: 'coral'
  },
  {
    key: 'monthly',
    label: '本月新增',
    value: stats.value.monthly_count ?? stats.value.today ?? 0,
    icon: CalendarPlus,
    tone: 'blue'
  }
])
const selectionState = computed(() => {
  if (subscriptions.value.length === 0 || selectedSubscriptions.value.length === 0) return false
  return selectedSubscriptions.value.length === subscriptions.value.length ? true : 'indeterminate'
})

const hasPermission = (permission) => authStore.hasPermission(permission)
const statusName = (status) => ({ active: '活跃', unsubscribed: '已退订', cancelled: '已退订' })[status] || status || '-'
const statusTone = (status) => status === 'active' ? 'green' : 'gray'
const localeName = (locale) => ({ zh: '中文', en: 'English' })[locale] || locale || '-'
const sourceName = (source) => ({ website: '网站', popup: '弹窗', footer: '页脚', checkout: '结账页' })[source] || source || '-'
const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'

const fetchSubscriptions = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/subscriptions', {
      params: {
        page: pagination.page,
        page_size: pagination.pageSize,
        ...(filters.status !== 'all' ? { status: filters.status } : {})
      }
    })
    subscriptions.value = response.data.subscriptions || []
    pagination.total = response.data.pagination?.total ?? response.data.total ?? 0
    selectedSubscriptions.value = []
  } catch (error) {
    console.error('Failed to fetch subscriptions:', error)
  } finally {
    loading.value = false
  }
}
const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/subscriptions/stats')
    stats.value = response.data || {}
  } catch (error) {
    console.error('Failed to fetch subscription stats:', error)
  }
}
const refreshSubscriptions = () => Promise.all([fetchSubscriptions(), fetchStats()])
const applyFilters = () => { pagination.page = 1; fetchSubscriptions() }
const resetFilters = () => { filters.status = 'all'; pagination.page = 1; fetchSubscriptions() }
const updatePage = (page) => { pagination.page = page; fetchSubscriptions() }
const updatePageSize = (pageSize) => { pagination.pageSize = pageSize; pagination.page = 1; fetchSubscriptions() }

const isSelected = (email) => selectedSubscriptions.value.some((subscription) => subscription.email === email)
const toggleAllSubscriptions = (checked) => {
  selectedSubscriptions.value = checked === true ? [...subscriptions.value] : []
}
const toggleSubscription = (subscription, checked) => {
  if (checked === true && !isSelected(subscription.email)) {
    selectedSubscriptions.value = [...selectedSubscriptions.value, subscription]
  } else if (checked !== true) {
    selectedSubscriptions.value = selectedSubscriptions.value.filter((selected) => selected.email !== subscription.email)
  }
}

const setConfirmation = (values) => Object.assign(confirmation, {
  open: true,
  type: '',
  target: null,
  status: '',
  confirmLabel: '确定',
  destructive: false,
  ...values
})
const requestToggleStatus = (subscription) => {
  const status = subscription.status === 'active' ? 'unsubscribed' : 'active'
  const restoring = status === 'active'
  setConfirmation({
    type: 'status',
    target: subscription,
    status,
    title: restoring ? '恢复订阅？' : '标记为退订？',
    description: `${subscription.email} 将被${restoring ? '恢复为活跃订阅' : '标记为已退订'}。`,
    confirmLabel: restoring ? '恢复订阅' : '确认退订'
  })
}
const requestDelete = (subscription) => setConfirmation({
  type: 'delete',
  target: subscription,
  title: '删除订阅？',
  description: `${subscription.email} 的订阅记录将被永久删除，此操作不可恢复。`,
  confirmLabel: '删除',
  destructive: true
})
const requestBatchDelete = () => setConfirmation({
  type: 'batch-delete',
  target: [...selectedSubscriptions.value],
  title: '批量删除订阅？',
  description: `${selectedSubscriptions.value.length} 条订阅记录将被永久删除，此操作不可恢复。`,
  confirmLabel: '批量删除',
  destructive: true
})
const executeConfirmedAction = async () => {
  const { type, target, status } = confirmation
  confirmation.open = false
  try {
    if (type === 'status') {
      await axios.patch(`/api/admin/subscriptions/${encodeURIComponent(target.email)}/status`, { status })
      toast.success(status === 'active' ? '订阅已恢复' : '订阅已标记为退订')
    } else if (type === 'delete') {
      await axios.delete(`/api/admin/subscriptions/${encodeURIComponent(target.email)}`)
      toast.success('订阅已删除')
    } else if (type === 'batch-delete') {
      const response = await axios.post('/api/admin/subscriptions/batch-delete', {
        emails: target.map((subscription) => subscription.email)
      })
      toast.success(`已删除 ${response.data.deleted ?? target.length} 条订阅`)
    }
    await refreshSubscriptions()
  } catch (error) {
    console.error('Failed to update subscriptions:', error)
  }
}

const exportEmails = async () => {
  try {
    const response = await axios.get('/api/admin/subscriptions/active-emails')
    const emails = Array.isArray(response.data.emails) ? response.data.emails : []
    const blob = new Blob([emails.join('\n')], { type: 'text/plain;charset=utf-8' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `subscriptions_${new Date().toISOString().slice(0, 10)}.txt`
    link.click()
    window.URL.revokeObjectURL(url)
    toast.success(`已导出 ${emails.length} 个活跃邮箱`)
  } catch (error) {
    console.error('Failed to export subscription emails:', error)
  }
}

onMounted(refreshSubscriptions)
</script>
