<template>
  <div class="space-y-4">
    <AdminPageHeader title="审计日志" description="追踪后台操作、资源变更和请求结果">
      <template #actions>
        <Button variant="outline" @click="refreshLogs">
          <RefreshCw class="size-4" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <AuditLogFilterPanel
      :filters="filters"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <AuditLogTablePanel
      :loading="loading"
      :logs="logs"
      :pagination="pagination"
      :action-name="actionName"
      :action-tone="actionTone"
      :resource-name="resourceName"
      :method-tone="methodTone"
      :duration-class="durationClass"
      :format-date="formatDate"
      @view-detail="viewDetail"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />

    <AuditLogDetailDialog
      v-model:open="detailDialogVisible"
      :detail-loading="detailLoading"
      :current-log="currentLog"
      :action-name="actionName"
      :action-tone="actionTone"
      :resource-name="resourceName"
      :method-tone="methodTone"
      :format-date="formatDate"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RefreshCw, ScrollText, ShieldCheck, ShieldX } from '@lucide/vue'
import AuditLogDetailDialog from '@/components/admin/audit-log/AuditLogDetailDialog.vue'
import AuditLogFilterPanel from '@/components/admin/audit-log/AuditLogFilterPanel.vue'
import AuditLogTablePanel from '@/components/admin/audit-log/AuditLogTablePanel.vue'
import type {
  AuditLogDetailResponse,
  AuditLogFilters,
  AuditLogPagination,
  AuditLogRecord,
  AuditLogsResponse,
  AuditLogStats,
  AuditLogTone
} from '@/modules/audit-log/auditLogTypes'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import { Button } from '@/components/ui/button'
import axios from '@/utils/axios'

const loading = ref(false)
const logs = ref<AuditLogRecord[]>([])
const stats = ref<AuditLogStats>({})
const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const currentLog = ref<AuditLogRecord | null>(null)
const filters = reactive<AuditLogFilters>({ keyword: '', action: 'all', resource: 'all', user_id: '', ip_address: '', start_date: '', end_date: '' })
const pagination = reactive<AuditLogPagination>({ page: 1, pageSize: 20, total: 0 })

const statItems = computed(() => [
  { key: 'total', label: '总日志数', value: stats.value.total_count || 0, icon: ScrollText, tone: 'gray' },
  { key: 'today', label: '今日操作', value: stats.value.today_count || 0, icon: RefreshCw, tone: 'blue' },
  { key: 'success', label: '成功操作', value: stats.value.success_count || 0, icon: ShieldCheck, tone: 'green' },
  { key: 'failed', label: '失败操作', value: stats.value.failed_count || 0, icon: ShieldX, tone: 'coral' }
])

const actionNames: Record<string, string> = { create: '创建', update: '更新', delete: '删除', view: '查看' }
const actionTones: Record<string, AuditLogTone> = { create: 'green', update: 'amber', delete: 'coral', view: 'gray' }
const resourceNames: Record<string, string> = {
  user: '用户', product: '商品', order: '订单', post: '文章', ticket: '工单', faq: 'FAQ', gallery: '图库',
  subscription: '订阅', marketing: '营销', setting: '设置'
}
const methodTones: Record<string, AuditLogTone> = { GET: 'gray', POST: 'green', PUT: 'amber', PATCH: 'blue', DELETE: 'coral' }

const actionName = (action?: string | null): string => actionNames[action || ''] || action || '-'
const actionTone = (action?: string | null): AuditLogTone => actionTones[action || ''] || 'gray'
const resourceName = (resource?: string | null): string => resourceNames[resource || ''] || resource || '-'
const methodTone = (method?: string | null): AuditLogTone => methodTones[method || ''] || 'gray'
const durationClass = (duration?: number | string | null): string => Number(duration || 0) >= 1000 ? 'font-medium text-destructive' : Number(duration || 0) >= 300 ? 'text-amber-700' : ''
const formatDate = (dateString?: string | null): string => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
const dateParams = (): Record<string, string> => ({
  ...(filters.start_date ? { start_date: filters.start_date } : {}),
  ...(filters.end_date ? { end_date: filters.end_date } : {})
})

const fetchLogs = async (): Promise<void> => {
  loading.value = true
  try {
    const keyword = filters.keyword.trim()
    const endpoint = keyword ? '/api/admin/logs/search' : '/api/admin/logs'
    const params: Record<string, string | number> = keyword
      ? { keyword, page: pagination.page, page_size: pagination.pageSize }
      : {
          page: pagination.page,
          page_size: pagination.pageSize,
          ...(filters.action !== 'all' ? { action: filters.action } : {}),
          ...(filters.resource !== 'all' ? { resource: filters.resource } : {}),
          ...(filters.user_id ? { user_id: filters.user_id } : {}),
          ...(filters.ip_address.trim() ? { ip_address: filters.ip_address.trim() } : {}),
          ...dateParams()
        }
    const response = await axios.get<AuditLogsResponse>(endpoint, { params })
    logs.value = response.data.logs || []
    pagination.total = response.data.total || 0
  } catch (error) {
    console.error('Failed to fetch audit logs:', error)
  } finally {
    loading.value = false
  }
}
const fetchStats = async (): Promise<void> => {
  try {
    const response = await axios.get<AuditLogStats>('/api/admin/logs/stats', { params: dateParams() })
    stats.value = response.data || {}
  } catch (error) {
    console.error('Failed to fetch audit stats:', error)
  }
}
const refreshLogs = async (): Promise<void> => {
  await Promise.all([fetchLogs(), fetchStats()])
}
const applyFilters = (): void => {
  pagination.page = 1
  void refreshLogs()
}
const resetFilters = (): void => {
  Object.assign(filters, { keyword: '', action: 'all', resource: 'all', user_id: '', ip_address: '', start_date: '', end_date: '' })
  pagination.page = 1
  void refreshLogs()
}
const updatePage = (page: number): void => {
  pagination.page = page
  void fetchLogs()
}
const updatePageSize = (pageSize: number): void => {
  pagination.pageSize = pageSize
  pagination.page = 1
  void fetchLogs()
}
const viewDetail = async (log: AuditLogRecord): Promise<void> => {
  currentLog.value = log
  detailDialogVisible.value = true
  detailLoading.value = true
  try {
    const response = await axios.get<AuditLogDetailResponse>(`/api/admin/logs/${log.id}`)
    currentLog.value = response.data.log || log
  } catch (error) {
    console.error('Failed to fetch audit log detail:', error)
  } finally {
    detailLoading.value = false
  }
}

onMounted(() => {
  void refreshLogs()
})
</script>

