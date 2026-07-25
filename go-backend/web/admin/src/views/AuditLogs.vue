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

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { RefreshCw, ScrollText, ShieldCheck, ShieldX } from '@lucide/vue'
import AuditLogDetailDialog from '@/components/admin/audit-log/AuditLogDetailDialog.vue'
import AuditLogFilterPanel from '@/components/admin/audit-log/AuditLogFilterPanel.vue'
import AuditLogTablePanel from '@/components/admin/audit-log/AuditLogTablePanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import { Button } from '@/components/ui/button'
import axios from '@/utils/axios'

const loading = ref(false)
const logs = ref([])
const stats = ref({})
const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const currentLog = ref(null)
const filters = reactive({ keyword: '', action: 'all', resource: 'all', user_id: '', ip_address: '', start_date: '', end_date: '' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const statItems = computed(() => [
  { key: 'total', label: '总日志数', value: stats.value.total_count || 0, icon: ScrollText, tone: 'gray' },
  { key: 'today', label: '今日操作', value: stats.value.today_count || 0, icon: RefreshCw, tone: 'blue' },
  { key: 'success', label: '成功操作', value: stats.value.success_count || 0, icon: ShieldCheck, tone: 'green' },
  { key: 'failed', label: '失败操作', value: stats.value.failed_count || 0, icon: ShieldX, tone: 'coral' }
])

const actionName = (action) => ({ create: '创建', update: '更新', delete: '删除', view: '查看' })[action] || action || '-'
const actionTone = (action) => ({ create: 'green', update: 'amber', delete: 'coral', view: 'gray' })[action] || 'gray'
const resourceName = (resource) => ({
  user: '用户', product: '商品', order: '订单', post: '文章', ticket: '工单', faq: 'FAQ', gallery: '图库',
  subscription: '订阅', marketing: '营销', setting: '设置'
})[resource] || resource || '-'
const methodTone = (method) => ({ GET: 'gray', POST: 'green', PUT: 'amber', PATCH: 'blue', DELETE: 'coral' })[method] || 'gray'
const durationClass = (duration) => Number(duration || 0) >= 1000 ? 'font-medium text-destructive' : Number(duration || 0) >= 300 ? 'text-amber-700' : ''
const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
const dateParams = () => ({
  ...(filters.start_date ? { start_date: filters.start_date } : {}),
  ...(filters.end_date ? { end_date: filters.end_date } : {})
})

const fetchLogs = async () => {
  loading.value = true
  try {
    const keyword = filters.keyword.trim()
    const endpoint = keyword ? '/api/admin/logs/search' : '/api/admin/logs'
    const params = keyword
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
    const response = await axios.get(endpoint, { params })
    logs.value = response.data.logs || []
    pagination.total = response.data.total || 0
  } catch (error) {
    console.error('Failed to fetch audit logs:', error)
  } finally {
    loading.value = false
  }
}
const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/logs/stats', { params: dateParams() })
    stats.value = response.data || {}
  } catch (error) {
    console.error('Failed to fetch audit stats:', error)
  }
}
const refreshLogs = () => Promise.all([fetchLogs(), fetchStats()])
const applyFilters = () => { pagination.page = 1; refreshLogs() }
const resetFilters = () => {
  Object.assign(filters, { keyword: '', action: 'all', resource: 'all', user_id: '', ip_address: '', start_date: '', end_date: '' })
  pagination.page = 1
  refreshLogs()
}
const updatePage = (page) => { pagination.page = page; fetchLogs() }
const updatePageSize = (pageSize) => { pagination.pageSize = pageSize; pagination.page = 1; fetchLogs() }
const viewDetail = async (log) => {
  currentLog.value = log
  detailDialogVisible.value = true
  detailLoading.value = true
  try {
    const response = await axios.get(`/api/admin/logs/${log.id}`)
    currentLog.value = response.data.log || log
  } catch (error) {
    console.error('Failed to fetch audit log detail:', error)
  } finally {
    detailLoading.value = false
  }
}

onMounted(refreshLogs)
</script>
