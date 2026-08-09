<template>
  <div class="space-y-4">
    <AdminPageHeader title="邮件订阅" description="查看邮件订阅来源、状态和退订记录">
      <template #actions>
        <Button v-if="hasPermission('subscription:export')" variant="outline" @click="exportEmails">
          <Download class="size-4" />
          导出活跃邮箱
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <SubscriptionFilterPanel
      :filters="filters"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <SubscriptionTablePanel
      :loading="loading"
      :subscriptions="subscriptions"
      :selected-subscriptions="selectedSubscriptions"
      :selection-state="selectionState"
      :pagination="pagination"
      :can-edit="hasPermission('subscription:edit')"
      :can-delete="hasPermission('subscription:delete')"
      :is-selected="isSelected"
      :status-name="statusName"
      :status-tone="statusTone"
      :locale-name="localeName"
      :source-name="sourceName"
      :format-date="formatDate"
      @toggle-all="toggleAllSubscriptions"
      @toggle-subscription="toggleSubscription"
      @request-toggle-status="requestToggleStatus"
      @request-delete="requestDelete"
      @request-batch-delete="requestBatchDelete"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />

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

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  CalendarPlus,
  Download,
  Mail,
  MailCheck,
  UserMinus
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import SubscriptionFilterPanel from '@/components/admin/subscription/SubscriptionFilterPanel.vue'
import SubscriptionTablePanel from '@/components/admin/subscription/SubscriptionTablePanel.vue'
import type {
  SubscriptionBatchDeleteResponse,
  SubscriptionConfirmation,
  SubscriptionEmailsResponse,
  SubscriptionFilters,
  SubscriptionListResponse,
  SubscriptionPagination,
  SubscriptionRecord,
  SubscriptionSelectionState,
  SubscriptionStats,
  SubscriptionStatusTone
} from '@/components/admin/subscription/subscriptionTypes'
import { Button } from '@/components/ui/button'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()
const supportedLanguages = useSupportedLanguages()
const loading = ref(false)
const subscriptions = ref<SubscriptionRecord[]>([])
const selectedSubscriptions = ref<SubscriptionRecord[]>([])
const stats = ref<SubscriptionStats>({})
const filters = reactive<SubscriptionFilters>({ search: '', status: 'all' })
const pagination = reactive<SubscriptionPagination>({ page: 1, pageSize: 20, total: 0 })
const confirmation = reactive<SubscriptionConfirmation>({
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
const selectionState = computed<SubscriptionSelectionState>(() => {
  if (subscriptions.value.length === 0 || selectedSubscriptions.value.length === 0) return false
  return selectedSubscriptions.value.length === subscriptions.value.length ? true : 'indeterminate'
})

const statusNames: Record<string, string> = { active: '活跃', unsubscribed: '已退订', cancelled: '已退订' }
const sourceNames: Record<string, string> = { website: '网站', popup: '弹窗', footer: '页脚', checkout: '结账页' }

const hasPermission = (permission: string): boolean => authStore.hasPermission(permission)
const statusName = (status?: string | null): string => statusNames[status || ''] || status || '-'
const statusTone = (status?: string | null): SubscriptionStatusTone => status === 'active' ? 'green' : 'gray'
const localeName = supportedLanguages.localeName
const sourceName = (source?: string | null): string => sourceNames[source || ''] || source || '-'
const formatDate = (dateString?: string | null): string => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'

const isSubscriptionRecord = (target: SubscriptionConfirmation['target']): target is SubscriptionRecord => (
  Boolean(target) && !Array.isArray(target)
)

const fetchSubscriptions = async (): Promise<void> => {
  loading.value = true
  try {
    const response = await axios.get<SubscriptionListResponse>('/api/admin/subscriptions', {
      params: {
        page: pagination.page,
        page_size: pagination.pageSize,
        ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
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
const fetchStats = async (): Promise<void> => {
  try {
    const response = await axios.get<SubscriptionStats>('/api/admin/subscriptions/stats')
    stats.value = response.data || {}
  } catch (error) {
    console.error('Failed to fetch subscription stats:', error)
  }
}
const refreshSubscriptions = async (): Promise<void> => {
  await Promise.all([fetchSubscriptions(), fetchStats()])
}
const applyFilters = (): void => {
  pagination.page = 1
  void fetchSubscriptions()
}
const resetFilters = (): void => {
  filters.search = ''
  filters.status = 'all'
  pagination.page = 1
  void fetchSubscriptions()
}
const updatePage = (page: number): void => {
  pagination.page = page
  void fetchSubscriptions()
}
const updatePageSize = (pageSize: number): void => {
  pagination.pageSize = pageSize
  pagination.page = 1
  void fetchSubscriptions()
}

const isSelected = (email: string): boolean => selectedSubscriptions.value.some((subscription) => subscription.email === email)
const toggleAllSubscriptions = (checked: SubscriptionSelectionState): void => {
  selectedSubscriptions.value = checked === true ? [...subscriptions.value] : []
}
const toggleSubscription = (subscription: SubscriptionRecord, checked: SubscriptionSelectionState): void => {
  if (checked === true && !isSelected(subscription.email)) {
    selectedSubscriptions.value = [...selectedSubscriptions.value, subscription]
  } else if (checked !== true) {
    selectedSubscriptions.value = selectedSubscriptions.value.filter((selected) => selected.email !== subscription.email)
  }
}

const setConfirmation = (values: Partial<SubscriptionConfirmation>): void => {
  Object.assign(confirmation, {
    open: true,
    type: '',
    target: null,
    status: '',
    confirmLabel: '确定',
    destructive: false,
    ...values
  })
}
const requestToggleStatus = (subscription: SubscriptionRecord): void => {
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
const requestDelete = (subscription: SubscriptionRecord): void => setConfirmation({
  type: 'delete',
  target: subscription,
  title: '删除订阅？',
  description: `${subscription.email} 的订阅记录将被永久删除，此操作不可恢复。`,
  confirmLabel: '删除',
  destructive: true
})
const requestBatchDelete = (): void => setConfirmation({
  type: 'batch-delete',
  target: [...selectedSubscriptions.value],
  title: '批量删除订阅？',
  description: `${selectedSubscriptions.value.length} 条订阅记录将被永久删除，此操作不可恢复。`,
  confirmLabel: '批量删除',
  destructive: true
})
const executeConfirmedAction = async (): Promise<void> => {
  const { type, target, status } = confirmation
  confirmation.open = false
  try {
    if (type === 'status' && isSubscriptionRecord(target)) {
      await axios.patch(`/api/admin/subscriptions/${encodeURIComponent(target.email)}/status`, { status })
      toast.success(status === 'active' ? '订阅已恢复' : '订阅已标记为退订')
    } else if (type === 'delete' && isSubscriptionRecord(target)) {
      await axios.delete(`/api/admin/subscriptions/${encodeURIComponent(target.email)}`)
      toast.success('订阅已删除')
    } else if (type === 'batch-delete' && Array.isArray(target)) {
      const response = await axios.post<SubscriptionBatchDeleteResponse>('/api/admin/subscriptions/batch-delete', {
        emails: target.map((subscription) => subscription.email)
      })
      toast.success(`已删除 ${response.data.deleted ?? target.length} 条订阅`)
    }
    await refreshSubscriptions()
  } catch (error) {
    console.error('Failed to update subscriptions:', error)
  }
}

const exportEmails = async (): Promise<void> => {
  try {
    const response = await axios.get<SubscriptionEmailsResponse>('/api/admin/subscriptions/active-emails')
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

onMounted(() => {
  void Promise.all([
    supportedLanguages.fetchLanguages(),
    refreshSubscriptions()
  ])
})
</script>
